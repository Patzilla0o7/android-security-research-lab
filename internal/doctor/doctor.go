package doctor

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/Patzilla0o7/android-security-research-lab/internal/config"
	"github.com/Patzilla0o7/android-security-research-lab/internal/toolchain"
)

type report struct {
	out              io.Writer
	pass, warn, fail int
	checks           []check
}

type check struct {
	Status string `json:"status"`
	Label  string `json:"label"`
	Detail string `json:"detail"`
}

type jsonReport struct {
	Checks   []check `json:"checks"`
	Passed   int     `json:"passed"`
	Warnings int     `json:"warnings"`
	Failed   int     `json:"failed"`
}

type FailedChecksError struct{ Count int }

func (e FailedChecksError) Error() string { return fmt.Sprintf("%d required check(s) failed", e.Count) }

func Run(root string, args []string, out io.Writer) error {
	jsonOutput := false
	if len(args) > 0 {
		if len(args) != 1 || args[0] != "--json" {
			return fmt.Errorf("usage: lab doctor [--json]")
		}
		jsonOutput = true
	}
	reportOut := out
	if jsonOutput {
		reportOut = io.Discard
	}
	r := &report{out: reportOut}
	r.section("Android Framework Research Doctor")
	settings, err := config.Load(filepath.Join(root, "config", "doctor.conf"), nil)
	if err != nil {
		return fmt.Errorf("load Doctor configuration: %w", err)
	}
	r.section("System Validation")
	r.os(settings)
	r.cpu(settings)
	r.memory(settings)
	r.disk(root, settings)
	r.labConfig(root)
	tools, err := toolchain.Load(filepath.Join(root, "config", "tools.conf"))
	if err != nil {
		return fmt.Errorf("load toolchain: %w", err)
	}
	for _, result := range toolchain.CheckAll(tools) {
		switch {
		case result.Status == "installed":
			r.good(result.Tool.Label, result.Detail)
		case result.Tool.Level == "required":
			r.bad(result.Tool.Label, result.Detail)
		default:
			r.caution(result.Tool.Label, result.Detail)
		}
	}
	r.gitIdentity()
	r.section("Summary")
	fmt.Fprintf(reportOut, "%-12s %d\n%-12s %d\n%-12s %d\n", "Passed", r.pass, "Warnings", r.warn, "Failed", r.fail)
	if jsonOutput {
		if err := json.NewEncoder(out).Encode(jsonReport{Checks: r.checks, Passed: r.pass, Warnings: r.warn, Failed: r.fail}); err != nil {
			return fmt.Errorf("encode Doctor report: %w", err)
		}
	}
	if r.fail > 0 {
		return FailedChecksError{Count: r.fail}
	}
	return nil
}

func (r *report) os(settings config.Values) {
	values, err := keyValues("/etc/os-release")
	if err != nil {
		r.bad("Ubuntu", err.Error())
		return
	}
	version := values["VERSION_ID"]
	for _, supported := range strings.Fields(settings["SUPPORTED_UBUNTU"]) {
		if version == supported {
			r.good("Ubuntu", version)
			return
		}
	}
	r.bad("Ubuntu", version)
}

func (r *report) cpu(settings config.Values) {
	if runtime.GOARCH == "amd64" {
		r.good("Architecture", "x86_64")
	} else {
		r.caution("Architecture", runtime.GOARCH)
	}
	cores, recommended := runtime.NumCPU(), number(settings, "RECOMMENDED_CPU_CORES")
	if cores >= recommended {
		r.good("CPU Cores", strconv.Itoa(cores))
	} else {
		r.caution("CPU Cores", fmt.Sprintf("%d (recommend %d+)", cores, recommended))
	}
}

func (r *report) memory(settings config.Values) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		r.bad("Memory", err.Error())
		return
	}
	defer file.Close()
	var kib uint64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if _, err := fmt.Sscanf(scanner.Text(), "MemTotal: %d kB", &kib); err == nil {
			break
		}
	}
	r.threshold("Memory", int(kib/1024/1024), number(settings, "MIN_MEMORY_GB"), number(settings, "RECOMMENDED_MEMORY_GB"), "GB")
}

func (r *report) disk(root string, settings config.Values) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(root, &stat); err != nil {
		r.bad("Disk", err.Error())
		return
	}
	gb := int(stat.Bavail * uint64(stat.Bsize) / 1024 / 1024 / 1024)
	r.threshold("Disk", gb, number(settings, "MIN_DISK_GB"), number(settings, "RECOMMENDED_DISK_GB"), "GB free")
}

func (r *report) labConfig(root string) {
	path := os.Getenv("LAB_CONFIG_FILE")
	if path == "" {
		path = filepath.Join(root, "config", "lab.conf")
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		r.caution("Lab Config", "Not found (copy config/lab.conf.example)")
		return
	} else if err != nil {
		r.bad("Lab Config", err.Error())
		return
	}
	if _, err := config.Load(path, map[string]string{"LAB_ROOT": root}); err != nil {
		r.bad("Lab Config", err.Error())
	} else {
		r.good("Lab Config", path)
	}
}

func (r *report) gitIdentity() {
	name, email := output("git", "config", "--global", "user.name"), output("git", "config", "--global", "user.email")
	if name != "" && email != "" {
		r.good("Git Identity", fmt.Sprintf("%s <%s>", name, email))
	} else {
		r.caution("Git Identity", "Not Configured")
	}
}

func (r *report) threshold(label string, value, minimum, recommended int, suffix string) {
	detail := fmt.Sprintf("%d %s", value, suffix)
	if value < minimum {
		r.bad(label, detail)
	} else if value < recommended {
		r.caution(label, detail)
	} else {
		r.good(label, detail)
	}
}
func (r *report) section(title string) {
	fmt.Fprintf(r.out, "\n============================================================\n%s\n============================================================\n", title)
}
func (r *report) good(label, detail string) {
	fmt.Fprintf(r.out, "✓ %-18s %s\n", label, detail)
	r.checks = append(r.checks, check{Status: "passed", Label: label, Detail: detail})
	r.pass++
}
func (r *report) caution(label, detail string) {
	fmt.Fprintf(r.out, "! %-18s %s\n", label, detail)
	r.checks = append(r.checks, check{Status: "warning", Label: label, Detail: detail})
	r.warn++
}
func (r *report) bad(label, detail string) {
	fmt.Fprintf(r.out, "✗ %-18s %s\n", label, detail)
	r.checks = append(r.checks, check{Status: "failed", Label: label, Detail: detail})
	r.fail++
}
func number(values config.Values, key string) int {
	value, _ := strconv.Atoi(values[key])
	return value
}

func keyValues(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	values := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		key, value, ok := strings.Cut(scanner.Text(), "=")
		if ok {
			values[key] = strings.Trim(strings.TrimSpace(value), `"`)
		}
	}
	return values, scanner.Err()
}
func output(name string, args ...string) string {
	value, err := exec.Command(name, args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(value))
}
