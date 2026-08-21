package collect

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Patzilla0o7/android-security-research-lab/internal/device"
	"github.com/Patzilla0o7/android-security-research-lab/internal/workspaces"
)

var validCaseID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)

type options struct {
	caseID, workspace, serial string
	timeout                   time.Duration
}

type commandRecord struct {
	Arguments []string `json:"arguments"`
	Status    string   `json:"status"`
	Error     string   `json:"error,omitempty"`
}

type fileRecord struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

type manifest struct {
	SchemaVersion int             `json:"schema_version"`
	Operation     string          `json:"operation"`
	CaseID        string          `json:"case_id"`
	Workspace     string          `json:"workspace"`
	Serial        string          `json:"serial"`
	CollectedAt   string          `json:"collected_at"`
	Status        string          `json:"status"`
	Commands      []commandRecord `json:"commands"`
	Files         []fileRecord    `json:"files"`
}

type deviceInfo struct {
	Serial     string            `json:"serial"`
	State      string            `json:"state"`
	Model      string            `json:"model,omitempty"`
	Product    string            `json:"product,omitempty"`
	Device     string            `json:"device,omitempty"`
	Properties map[string]string `json:"properties"`
	SELinux    string            `json:"selinux"`
}

type dependencies struct {
	resolve  func(context.Context, string) (device.Device, error)
	adb      func(context.Context, ...string) (string, error)
	adbBytes func(context.Context, ...string) ([]byte, error)
	now      func() time.Time
}

func Run(root string, args []string, stdout, stderr io.Writer) int {
	return run(root, args, stdout, stderr, dependencies{resolve: device.Resolve, adb: device.ADBOutput, adbBytes: device.ADBBytes, now: time.Now})
}

func run(root string, args []string, stdout, stderr io.Writer, deps dependencies) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		Usage(stdout)
		return 0
	}
	operation, args := args[0], args[1:]
	if operation != "device-info" && operation != "logcat" && operation != "screenshot" && operation != "bugreport" && operation != "tombstones" && operation != "bundle" {
		return usageError(stderr, "unknown collect subcommand: "+operation)
	}
	opts, err := parseOptions(args)
	if err != nil {
		return usageError(stderr, err.Error())
	}
	profile, err := workspaces.Current(root, opts.workspace)
	if err != nil {
		return fail(stderr, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), opts.timeout)
	defer cancel()
	selected, err := deps.resolve(ctx, opts.serial)
	if err != nil {
		return fail(stderr, err)
	}
	stamp := deps.now().UTC().Format("20060102T150405Z")
	dir, err := createEvidenceDir(filepath.Join(root, "output", "evidence", profile.Name, opts.caseID), stamp)
	if err != nil {
		return fail(stderr, err)
	}
	m := manifest{SchemaVersion: 1, Operation: operation, CaseID: opts.caseID, Workspace: profile.Name, Serial: selected.Serial, CollectedAt: deps.now().UTC().Format(time.RFC3339), Status: "success", Commands: []commandRecord{}, Files: []fileRecord{}}

	var collectErrs []error
	if operation == "device-info" || operation == "bundle" {
		if err := collectDeviceInfo(ctx, dir, selected, &m, deps.adb); err != nil {
			collectErrs = append(collectErrs, err)
		}
	}
	if operation == "logcat" || operation == "bundle" {
		if err := collectLogcat(ctx, dir, selected.Serial, &m, deps.adb); err != nil {
			collectErrs = append(collectErrs, err)
		}
	}
	if operation == "screenshot" || operation == "bundle" {
		if err := collectScreenshot(ctx, dir, selected.Serial, &m, deps.adbBytes); err != nil {
			collectErrs = append(collectErrs, err)
		}
	}
	if operation == "bugreport" || operation == "bundle" {
		if err := collectBugreport(ctx, dir, selected.Serial, &m, deps.adb); err != nil {
			collectErrs = append(collectErrs, err)
		}
	}
	if operation == "tombstones" || operation == "bundle" {
		if err := collectTombstones(ctx, dir, selected.Serial, &m, deps.adbBytes); err != nil {
			collectErrs = append(collectErrs, err)
		}
	}
	collectErr := errors.Join(collectErrs...)
	if collectErr != nil {
		m.Status = "failed"
		if len(m.Files) > 0 {
			m.Status = "partial"
		}
	}
	if err := writeManifest(dir, &m); err != nil {
		return fail(stderr, err)
	}
	if err := writeChecksums(dir, m.Files); err != nil {
		return fail(stderr, err)
	}
	fmt.Fprintf(stdout, "Evidence directory: %s\n", dir)
	if collectErr != nil {
		return fail(stderr, fmt.Errorf("evidence collection partially failed: %w", collectErr))
	}
	fmt.Fprintln(stdout, "[ OK ] Evidence collection completed")
	return 0
}

func createEvidenceDir(parent, stamp string) (string, error) {
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return "", err
	}
	for attempt := 0; attempt < 100; attempt++ {
		name := stamp
		if attempt > 0 {
			name = fmt.Sprintf("%s-%02d", stamp, attempt)
		}
		path := filepath.Join(parent, name)
		if err := os.Mkdir(path, 0o755); err == nil {
			return path, nil
		} else if !os.IsExist(err) {
			return "", err
		}
	}
	return "", fmt.Errorf("unable to create unique evidence directory for %s", stamp)
}

func Usage(w io.Writer) {
	fmt.Fprint(w, `Usage: lab collect <command> --case <case-id> [options]

Commands:
    device-info
    logcat
    screenshot
    bugreport
    tombstones
    bundle

Options:
    --case <case-id>      Research case identifier (required).
    --workspace <name>    Use a Workspace profile instead of the active profile.
    --serial <serial>     Select an ADB device; required when multiple are usable.
    --timeout <duration>  Collection timeout; defaults to 10m.

Evidence is stored under output/evidence/<workspace>/<case-id>/<timestamp>/.
`)
}

func parseOptions(args []string) (options, error) {
	result := options{timeout: 10 * time.Minute}
	for i := 0; i < len(args); i++ {
		if i+1 >= len(args) {
			return result, fmt.Errorf("%s requires a value", args[i])
		}
		value := args[i+1]
		switch args[i] {
		case "--case":
			result.caseID = value
		case "--workspace":
			result.workspace = value
		case "--serial":
			result.serial = value
		case "--timeout":
			timeout, err := time.ParseDuration(value)
			if err != nil || timeout <= 0 {
				return result, fmt.Errorf("invalid timeout: %s", value)
			}
			result.timeout = timeout
		default:
			return result, fmt.Errorf("unknown collect option: %s", args[i])
		}
		i++
	}
	if !validCaseID.MatchString(result.caseID) {
		return result, fmt.Errorf("--case must use letters, numbers, dot, dash or underscore")
	}
	return result, nil
}

func collectDeviceInfo(ctx context.Context, dir string, selected device.Device, m *manifest, adb func(context.Context, ...string) (string, error)) error {
	keys := []string{"ro.build.version.release", "ro.build.version.sdk", "ro.build.fingerprint", "ro.build.id", "ro.build.type", "ro.debuggable", "sys.boot_completed"}
	info := deviceInfo{Serial: selected.Serial, State: selected.State, Model: selected.Model, Product: selected.Product, Device: selected.Device, Properties: map[string]string{}}
	for _, key := range keys {
		args := []string{"-s", selected.Serial, "shell", "getprop", key}
		value, err := runADB(ctx, m, adb, args...)
		if err != nil {
			return fmt.Errorf("read %s: %w", key, err)
		}
		info.Properties[key] = value
	}
	value, err := runADB(ctx, m, adb, "-s", selected.Serial, "shell", "getenforce")
	if err != nil {
		return fmt.Errorf("read SELinux state: %w", err)
	}
	info.SELinux = value
	return writeJSONFile(dir, "device.json", info, m)
}

func collectLogcat(ctx context.Context, dir, serial string, m *manifest, adb func(context.Context, ...string) (string, error)) error {
	output, err := runADB(ctx, m, adb, "-s", serial, "logcat", "-d", "-v", "threadtime")
	if err != nil {
		return fmt.Errorf("capture logcat: %w", err)
	}
	return writeFile(dir, "logcat.txt", []byte(output+"\n"), m)
}

func collectScreenshot(ctx context.Context, dir, serial string, m *manifest, adb func(context.Context, ...string) ([]byte, error)) error {
	data, err := runADBBinary(ctx, m, adb, "-s", serial, "exec-out", "screencap", "-p")
	if err != nil {
		return fmt.Errorf("capture screenshot: %w", err)
	}
	if len(data) < 8 || string(data[:8]) != "\x89PNG\r\n\x1a\n" {
		return fmt.Errorf("capture screenshot: ADB did not return a PNG image")
	}
	return writeFile(dir, "screenshot.png", data, m)
}

func collectBugreport(ctx context.Context, dir, serial string, m *manifest, adb func(context.Context, ...string) (string, error)) error {
	path := filepath.Join(dir, "bugreport.zip")
	if _, err := runADB(ctx, m, adb, "-s", serial, "bugreport", path); err != nil {
		return fmt.Errorf("capture bugreport: %w", err)
	}
	return recordExistingFile(dir, "bugreport.zip", m)
}

func collectTombstones(ctx context.Context, dir, serial string, m *manifest, adb func(context.Context, ...string) ([]byte, error)) error {
	data, err := runADBBinary(ctx, m, adb, "-s", serial, "exec-out", "tar", "-C", "/data/tombstones", "-cf", "-", ".")
	if err != nil {
		return fmt.Errorf("capture tombstones (root or readable /data/tombstones required): %w", err)
	}
	if len(data) == 0 {
		return fmt.Errorf("capture tombstones: empty archive")
	}
	return writeFile(dir, "tombstones.tar", data, m)
}

func runADB(ctx context.Context, m *manifest, adb func(context.Context, ...string) (string, error), args ...string) (string, error) {
	output, err := adb(ctx, args...)
	record := commandRecord{Arguments: append([]string{"adb"}, args...), Status: "success"}
	if err != nil {
		record.Status, record.Error = "failed", err.Error()
	}
	m.Commands = append(m.Commands, record)
	return output, err
}

func runADBBinary(ctx context.Context, m *manifest, adb func(context.Context, ...string) ([]byte, error), args ...string) ([]byte, error) {
	output, err := adb(ctx, args...)
	record := commandRecord{Arguments: append([]string{"adb"}, args...), Status: "success"}
	if err != nil {
		record.Status, record.Error = "failed", err.Error()
	}
	m.Commands = append(m.Commands, record)
	return output, err
}

func writeJSONFile(dir, name string, value any, m *manifest) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(dir, name, append(data, '\n'), m)
}

func writeFile(dir, name string, data []byte, m *manifest) error {
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return err
	}
	sum := sha256.Sum256(data)
	m.Files = append(m.Files, fileRecord{Path: name, Size: int64(len(data)), SHA256: hex.EncodeToString(sum[:])})
	return nil
}

func recordExistingFile(dir, name string, m *manifest) error {
	data, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		return fmt.Errorf("expected evidence file %s: %w", name, err)
	}
	sum := sha256.Sum256(data)
	m.Files = append(m.Files, fileRecord{Path: name, Size: int64(len(data)), SHA256: hex.EncodeToString(sum[:])})
	return nil
}

func writeManifest(dir string, m *manifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "manifest.json"), append(data, '\n'), 0o644)
}

func writeChecksums(dir string, files []fileRecord) error {
	lines := make([]string, 0, len(files)+1)
	for _, file := range files {
		lines = append(lines, file.SHA256+"  "+file.Path)
	}
	manifestData, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return err
	}
	sum := sha256.Sum256(manifestData)
	lines = append(lines, hex.EncodeToString(sum[:])+"  manifest.json")
	return os.WriteFile(filepath.Join(dir, "SHA256SUMS"), []byte(strings.Join(lines, "\n")+"\n"), 0o644)
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
