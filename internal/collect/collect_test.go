package collect

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Patzilla0o7/android-security-research-lab/internal/device"
	"github.com/Patzilla0o7/android-security-research-lab/internal/workspaces"
)

func collectFixture(t *testing.T) (string, dependencies, map[string]string) {
	t.Helper()
	root := t.TempDir()
	if _, err := workspaces.Add(root, "android-15", filepath.Join(t.TempDir(), "aosp"), "", "android-15.0.0_r1", "aosp_x86_64-eng", ""); err != nil {
		t.Fatal(err)
	}
	outputs := map[string]string{
		"-s emulator-5554 shell getprop ro.build.version.release": "15",
		"-s emulator-5554 shell getprop ro.build.version.sdk":     "35",
		"-s emulator-5554 shell getprop ro.build.fingerprint":     "aosp/test/fingerprint",
		"-s emulator-5554 shell getprop ro.build.id":              "TEST",
		"-s emulator-5554 shell getprop ro.build.type":            "userdebug",
		"-s emulator-5554 shell getprop ro.debuggable":            "1",
		"-s emulator-5554 shell getprop sys.boot_completed":       "1",
		"-s emulator-5554 shell getenforce":                       "Enforcing",
		"-s emulator-5554 logcat -d -v threadtime":                "08-21 12:00:00.000 I/Test: ready",
	}
	deps := dependencies{
		resolve: func(context.Context, string) (device.Device, error) {
			return device.Device{Serial: "emulator-5554", State: "device", Model: "sdk_phone"}, nil
		},
		adb: func(_ context.Context, args ...string) (string, error) {
			if len(args) >= 3 && args[len(args)-2] == "bugreport" {
				if err := os.WriteFile(args[len(args)-1], []byte("PK\x03\x04bugreport"), 0o644); err != nil {
					return "", err
				}
				return "Bug report finished", nil
			}
			value, ok := outputs[strings.Join(args, " ")]
			if !ok {
				return "", errors.New("unexpected adb command")
			}
			return value, nil
		},
		adbBytes: func(_ context.Context, args ...string) ([]byte, error) {
			switch strings.Join(args, " ") {
			case "-s emulator-5554 exec-out screencap -p":
				return append([]byte("\x89PNG\r\n\x1a\n"), []byte("image")...), nil
			case "-s emulator-5554 exec-out tar -C /data/tombstones -cf - .":
				return []byte("tar archive"), nil
			default:
				return nil, errors.New("unexpected binary adb command")
			}
		},
		now: func() time.Time { return time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC) },
	}
	return root, deps, outputs
}

func TestBundleWritesEvidenceAndChecksums(t *testing.T) {
	root, deps, _ := collectFixture(t)
	var stdout, stderr strings.Builder
	code := run(root, []string{"bundle", "--case", "CVE-2026-0001"}, &stdout, &stderr, deps)
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	dir := filepath.Join(root, "output", "evidence", "android-15", "CVE-2026-0001", "20260821T120000Z")
	for _, name := range []string{"device.json", "logcat.txt", "screenshot.png", "bugreport.zip", "tombstones.tar", "manifest.json", "SHA256SUMS"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
	data, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m manifest
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if m.Status != "success" || m.Serial != "emulator-5554" || len(m.Files) != 5 || len(m.Commands) != 12 {
		t.Fatalf("unexpected manifest: %#v", m)
	}
}

func TestBundleDoesNotOverwriteSameTimestamp(t *testing.T) {
	root, deps, _ := collectFixture(t)
	for i := 0; i < 2; i++ {
		var stdout, stderr strings.Builder
		if code := run(root, []string{"logcat", "--case", "case-1"}, &stdout, &stderr, deps); code != 0 {
			t.Fatalf("run %d: code=%d stderr=%s", i, code, stderr.String())
		}
	}
	parent := filepath.Join(root, "output", "evidence", "android-15", "case-1")
	entries, err := os.ReadDir(parent)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 || entries[0].Name() == entries[1].Name() {
		t.Fatalf("unexpected evidence directories: %#v", entries)
	}
}

func TestBundlePreservesPartialEvidence(t *testing.T) {
	root, deps, _ := collectFixture(t)
	originalADB := deps.adb
	deps.adb = func(ctx context.Context, args ...string) (string, error) {
		if strings.Contains(strings.Join(args, " "), " logcat ") {
			return "", errors.New("logcat unavailable")
		}
		return originalADB(ctx, args...)
	}
	var stdout, stderr strings.Builder
	if code := run(root, []string{"bundle", "--case", "case-1"}, &stdout, &stderr, deps); code != 1 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	dir := filepath.Join(root, "output", "evidence", "android-15", "case-1", "20260821T120000Z")
	if _, err := os.Stat(filepath.Join(dir, "device.json")); err != nil {
		t.Fatalf("device evidence was not preserved: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"status": "partial"`) || !strings.Contains(string(data), "logcat unavailable") {
		t.Fatalf("unexpected manifest: %s", data)
	}
}

func TestCollectRejectsUnsafeCaseID(t *testing.T) {
	root, deps, _ := collectFixture(t)
	var stdout, stderr strings.Builder
	if code := run(root, []string{"device-info", "--case", "../escape"}, &stdout, &stderr, deps); code != 2 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(root, "output")); !os.IsNotExist(err) {
		t.Fatalf("unsafe case created output: %v", err)
	}
}

func TestCollectRequiresWorkspace(t *testing.T) {
	var stdout, stderr strings.Builder
	deps := dependencies{
		resolve: func(context.Context, string) (device.Device, error) {
			return device.Device{}, errors.New("must not run")
		},
		now: time.Now,
	}
	if code := run(t.TempDir(), []string{"logcat", "--case", "case-1"}, &stdout, &stderr, deps); code != 1 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
}

func TestScreenshotRejectsInvalidPNG(t *testing.T) {
	root, deps, _ := collectFixture(t)
	deps.adbBytes = func(context.Context, ...string) ([]byte, error) { return []byte("not png"), nil }
	var stdout, stderr strings.Builder
	if code := run(root, []string{"screenshot", "--case", "case-1"}, &stdout, &stderr, deps); code != 1 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	dir := filepath.Join(root, "output", "evidence", "android-15", "case-1", "20260821T120000Z")
	data, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"status": "failed"`) {
		t.Fatalf("unexpected manifest: %s", data)
	}
}

func TestBugreportRequiresOutputFile(t *testing.T) {
	root, deps, _ := collectFixture(t)
	deps.adb = func(context.Context, ...string) (string, error) { return "done", nil }
	var stdout, stderr strings.Builder
	if code := run(root, []string{"bugreport", "--case", "case-1"}, &stdout, &stderr, deps); code != 1 || !strings.Contains(stderr.String(), "expected evidence file") {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
}

func TestParseTimeout(t *testing.T) {
	opts, err := parseOptions([]string{"--case", "case-1", "--timeout", "30s"})
	if err != nil || opts.timeout != 30*time.Second {
		t.Fatalf("opts=%#v err=%v", opts, err)
	}
	if _, err := parseOptions([]string{"--case", "case-1", "--timeout", "0s"}); err == nil {
		t.Fatal("expected invalid timeout")
	}
}
