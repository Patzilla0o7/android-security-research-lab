package device

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeRunner struct {
	outputs map[string]string
	errors  map[string]error
}

func (f fakeRunner) Output(_ context.Context, args ...string) (string, error) {
	key := strings.Join(args, " ")
	if err := f.errors[key]; err != nil {
		return "", err
	}
	return f.outputs[key], nil
}

const deviceList = `List of devices attached
emulator-5554 device product:sdk_phone_x86_64 model:sdk_gphone64_x86_64 device:emu64xa transport_id:1
R3CN offline transport_id:2
USB1 unauthorized usb:1-1 transport_id:3
`

func TestParseDevices(t *testing.T) {
	devices := parseDevices(deviceList)
	if len(devices) != 3 {
		t.Fatalf("len = %d, want 3", len(devices))
	}
	if devices[0].Serial != "emulator-5554" || devices[0].Model != "sdk_gphone64_x86_64" || devices[1].State != "offline" {
		t.Fatalf("unexpected devices: %#v", devices)
	}
}

func TestSelectDeviceRequiresSerialForMultipleDevices(t *testing.T) {
	adb := fakeRunner{outputs: map[string]string{
		"devices -l": "List of devices attached\nemulator-5554 device\nemulator-5556 device",
	}}
	_, err := selectDevice(context.Background(), adb, "")
	if err == nil || !strings.Contains(err.Error(), "specify --serial") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSelectDeviceRejectsOffline(t *testing.T) {
	adb := fakeRunner{outputs: map[string]string{"devices -l": deviceList}}
	_, err := selectDevice(context.Background(), adb, "R3CN")
	if err == nil || !strings.Contains(err.Error(), "offline") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStatusUsesSelectedSerial(t *testing.T) {
	outputs := map[string]string{"devices -l": deviceList}
	for key, value := range map[string]string{
		"ro.build.version.release": "15",
		"ro.build.version.sdk":     "35",
		"ro.build.fingerprint":     "aosp/test/fingerprint",
		"ro.build.type":            "userdebug",
		"ro.debuggable":            "1",
		"sys.boot_completed":       "1",
	} {
		outputs["-s emulator-5554 shell getprop "+key] = value
	}
	outputs["-s emulator-5554 shell getenforce"] = "Enforcing"
	var stdout, stderr strings.Builder
	if code := run([]string{"status"}, &stdout, &stderr, fakeRunner{outputs: outputs}); code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "aosp/test/fingerprint") || !strings.Contains(stdout.String(), "Enforcing") {
		t.Fatalf("unexpected output: %s", stdout.String())
	}
}

func TestWaitReady(t *testing.T) {
	adb := fakeRunner{outputs: map[string]string{
		"-s emulator-5554 wait-for-device": "",
		"devices -l":                       deviceList,
		"-s emulator-5554 shell getprop sys.boot_completed": "1",
	}}
	var stdout, stderr strings.Builder
	if code := run([]string{"wait", "--serial", "emulator-5554", "--timeout", "2s"}, &stdout, &stderr, adb); code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Device is ready") {
		t.Fatalf("unexpected output: %s", stdout.String())
	}
}

func TestListReportsADBFailure(t *testing.T) {
	adb := fakeRunner{errors: map[string]error{"devices -l": errors.New("adb unavailable")}}
	var stdout, stderr strings.Builder
	if code := run([]string{"list"}, &stdout, &stderr, adb); code != 1 || !strings.Contains(stderr.String(), "adb unavailable") {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
}

func TestParseOptions(t *testing.T) {
	serial, timeout, err := parseOptions([]string{"--serial", "abc", "--timeout", "30s"}, true)
	if err != nil || serial != "abc" || timeout.String() != "30s" {
		t.Fatalf("serial=%s timeout=%s err=%v", serial, timeout, err)
	}
	if _, _, err := parseOptions([]string{"--timeout", "0s"}, true); err == nil {
		t.Fatal("expected invalid timeout")
	}
}
