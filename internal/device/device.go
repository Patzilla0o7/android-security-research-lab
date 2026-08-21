package device

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const defaultWaitTimeout = 2 * time.Minute

type Device struct {
	Serial      string
	State       string
	Product     string
	Model       string
	Device      string
	TransportID string
}

type runner interface {
	Output(context.Context, ...string) (string, error)
}

type adbRunner struct{}

func (adbRunner) raw(ctx context.Context, args ...string) ([]byte, error) {
	path, err := exec.LookPath("adb")
	if err != nil {
		return nil, fmt.Errorf("ADB is not installed or not in PATH")
	}
	data, err := exec.CommandContext(ctx, path, args...).Output()
	if err != nil {
		detail := ""
		if exitErr, ok := err.(*exec.ExitError); ok {
			detail = strings.TrimSpace(string(exitErr.Stderr))
		}
		if detail == "" {
			detail = err.Error()
		}
		return nil, errors.New(detail)
	}
	return data, nil

}

func (a adbRunner) Output(ctx context.Context, args ...string) (string, error) {
	data, err := a.raw(ctx, args...)
	return strings.TrimSpace(string(data)), err
}

func Run(args []string, stdout, stderr io.Writer) int {
	return run(args, stdout, stderr, adbRunner{})
}

// Resolve selects one usable ADB device, requiring serial when selection is ambiguous.
func Resolve(ctx context.Context, serial string) (Device, error) {
	return selectDevice(ctx, adbRunner{}, serial)
}

// ADBOutput executes ADB and returns its trimmed combined output.
func ADBOutput(ctx context.Context, args ...string) (string, error) {
	return adbRunner{}.Output(ctx, args...)
}

// ADBBytes executes ADB without text conversion or trimming for binary evidence.
func ADBBytes(ctx context.Context, args ...string) ([]byte, error) {
	return adbRunner{}.raw(ctx, args...)
}

func run(args []string, stdout, stderr io.Writer, adb runner) int {
	command := "list"
	if len(args) > 0 {
		command, args = args[0], args[1:]
	}
	switch command {
	case "help", "--help", "-h":
		Usage(stdout)
		return 0
	case "list":
		if len(args) != 0 {
			return usageError(stderr, "list accepts no arguments")
		}
		return listCommand(stdout, stderr, adb)
	case "status":
		serial, _, err := parseOptions(args, false)
		if err != nil {
			return usageError(stderr, err.Error())
		}
		return statusCommand(serial, stdout, stderr, adb)
	case "wait":
		serial, timeout, err := parseOptions(args, true)
		if err != nil {
			return usageError(stderr, err.Error())
		}
		return waitCommand(serial, timeout, stdout, stderr, adb)
	default:
		return usageError(stderr, "unknown device subcommand: "+command)
	}
}

func Usage(w io.Writer) {
	fmt.Fprint(w, `Usage: lab device <command> [options]

Commands:
    list
    status [--serial <serial>]
    wait [--serial <serial>] [--timeout <duration>]

When more than one usable device is connected, --serial is required.
`)
}

func listCommand(stdout, stderr io.Writer, adb runner) int {
	devices, err := devices(context.Background(), adb)
	if err != nil {
		return fail(stderr, err)
	}
	if len(devices) == 0 {
		fmt.Fprintln(stdout, "No ADB devices found.")
		return 0
	}
	fmt.Fprintln(stdout, "SERIAL\tSTATE\tMODEL\tPRODUCT\tTRANSPORT")
	for _, d := range devices {
		fmt.Fprintf(stdout, "%s\t%s\t%s\t%s\t%s\n", d.Serial, d.State, valueOr(d.Model, "-"), valueOr(d.Product, "-"), valueOr(d.TransportID, "-"))
	}
	return 0
}

func statusCommand(requested string, stdout, stderr io.Writer, adb runner) int {
	ctx := context.Background()
	selected, err := selectDevice(ctx, adb, requested)
	if err != nil {
		return fail(stderr, err)
	}
	properties := []struct{ label, key string }{
		{"Android", "ro.build.version.release"},
		{"API level", "ro.build.version.sdk"},
		{"Fingerprint", "ro.build.fingerprint"},
		{"Build type", "ro.build.type"},
		{"Debuggable", "ro.debuggable"},
		{"Boot completed", "sys.boot_completed"},
	}
	fmt.Fprintf(stdout, "Serial         : %s\nState          : %s\nModel          : %s\n", selected.Serial, selected.State, valueOr(selected.Model, "unknown"))
	for _, property := range properties {
		value, err := adb.Output(ctx, "-s", selected.Serial, "shell", "getprop", property.key)
		if err != nil {
			return fail(stderr, fmt.Errorf("read %s: %w", property.key, err))
		}
		fmt.Fprintf(stdout, "%-15s: %s\n", property.label, valueOr(value, "unknown"))
	}
	selinux, err := adb.Output(ctx, "-s", selected.Serial, "shell", "getenforce")
	if err != nil {
		return fail(stderr, fmt.Errorf("read SELinux state: %w", err))
	}
	fmt.Fprintf(stdout, "SELinux        : %s\n", valueOr(selinux, "unknown"))
	return 0
}

func waitCommand(requested string, timeout time.Duration, stdout, stderr io.Writer, adb runner) int {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	fmt.Fprintf(stdout, "Waiting for device (timeout %s)...\n", timeout)
	if _, err := adb.Output(ctx, serialArgs(requested, "wait-for-device")...); err != nil {
		if ctx.Err() != nil {
			return fail(stderr, fmt.Errorf("timed out waiting for ADB device after %s", timeout))
		}
		return fail(stderr, fmt.Errorf("wait for ADB device: %w", err))
	}
	selected, err := selectDevice(ctx, adb, requested)
	if err != nil {
		return fail(stderr, err)
	}
	for {
		completed, err := adb.Output(ctx, "-s", selected.Serial, "shell", "getprop", "sys.boot_completed")
		if err == nil && completed == "1" {
			fmt.Fprintf(stdout, "[ OK ] Device is ready: %s\n", selected.Serial)
			return 0
		}
		select {
		case <-ctx.Done():
			return fail(stderr, fmt.Errorf("timed out waiting for Android boot after %s", timeout))
		case <-time.After(time.Second):
		}
	}
}

func devices(ctx context.Context, adb runner) ([]Device, error) {
	output, err := adb.Output(ctx, "devices", "-l")
	if err != nil {
		return nil, fmt.Errorf("list ADB devices: %w", err)
	}
	return parseDevices(output), nil
}

func parseDevices(output string) []Device {
	var result []Device
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 2 || fields[0] == "List" || strings.HasPrefix(fields[0], "*") {
			continue
		}
		d := Device{Serial: fields[0], State: fields[1]}
		for _, field := range fields[2:] {
			key, value, ok := strings.Cut(field, ":")
			if !ok {
				continue
			}
			switch key {
			case "product":
				d.Product = value
			case "model":
				d.Model = value
			case "device":
				d.Device = value
			case "transport_id":
				d.TransportID = value
			}
		}
		result = append(result, d)
	}
	return result
}

func selectDevice(ctx context.Context, adb runner, requested string) (Device, error) {
	all, err := devices(ctx, adb)
	if err != nil {
		return Device{}, err
	}
	if requested != "" {
		for _, d := range all {
			if d.Serial == requested {
				if d.State != "device" {
					return Device{}, fmt.Errorf("device %s is %s", requested, d.State)
				}
				return d, nil
			}
		}
		return Device{}, fmt.Errorf("ADB device not found: %s", requested)
	}
	var usable []Device
	for _, d := range all {
		if d.State == "device" {
			usable = append(usable, d)
		}
	}
	if len(usable) == 0 {
		return Device{}, fmt.Errorf("no usable ADB device found")
	}
	if len(usable) > 1 {
		return Device{}, fmt.Errorf("multiple ADB devices found; specify --serial")
	}
	return usable[0], nil
}

func parseOptions(args []string, allowTimeout bool) (string, time.Duration, error) {
	serial, timeout := "", defaultWaitTimeout
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--serial":
			if i+1 >= len(args) || args[i+1] == "" {
				return "", 0, fmt.Errorf("--serial requires a value")
			}
			i++
			serial = args[i]
		case "--timeout":
			if !allowTimeout {
				return "", 0, fmt.Errorf("--timeout is only valid for wait")
			}
			if i+1 >= len(args) {
				return "", 0, fmt.Errorf("--timeout requires a duration")
			}
			i++
			parsed, err := time.ParseDuration(args[i])
			if err != nil || parsed <= 0 {
				return "", 0, fmt.Errorf("invalid timeout: %s", strconv.Quote(args[i]))
			}
			timeout = parsed
		default:
			return "", 0, fmt.Errorf("unknown device option: %s", args[i])
		}
	}
	return serial, timeout, nil
}

func serialArgs(serial string, args ...string) []string {
	if serial == "" {
		return args
	}
	return append([]string{"-s", serial}, args...)
}

func valueOr(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func usageError(w io.Writer, message string) int {
	fmt.Fprintf(w, "[FAIL] %s\n", message)
	Usage(w)
	return 2
}

func fail(w io.Writer, err error) int {
	fmt.Fprintf(w, "[FAIL] %v\n", err)
	return 1
}
