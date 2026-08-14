package build

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Patzilla0o7/android-security-research-lab/internal/workspaces"
)

func TestPlanExecuteAndStatus(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(t.TempDir(), "aosp")
	for _, path := range []string{filepath.Join(source, ".repo"), filepath.Join(source, "build")} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(source, "build", "envsetup.sh"), []byte("# test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := workspaces.Add(root, "android-14", source, "", "android-14-tag", "aosp_x86_64-eng", ""); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "scripts", "aosp-build.sh"), []byte("#!/usr/bin/env bash\nprintf '%s\\n' \"$*\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	oldNow := now
	calls := 0
	now = func() time.Time { calls++; return time.Date(2026, 8, 14, 10, 0, calls-1, 0, time.UTC) }
	t.Cleanup(func() { now = oldNow })
	var stdout, stderr bytes.Buffer
	if code := Run(root, []string{"--workspace", "android-14", "--jobs", "4", "--module", "services"}, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "No changes were made") {
		t.Fatalf("plan code=%d out=%s err=%s", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := Run(root, []string{"--workspace", "android-14", "--jobs", "4", "--module", "services", "--apply"}, &stdout, &stderr); code != 0 {
		t.Fatalf("build code=%d out=%s err=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), source+" aosp_x86_64-eng 4") {
		t.Fatalf("adapter args missing: %s", stdout.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := Run(root, []string{"status", "--workspace", "android-14"}, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "success") || !strings.Contains(stdout.String(), "services") {
		t.Fatalf("status code=%d out=%s err=%s", code, stdout.String(), stderr.String())
	}
}

func TestBuildValidation(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(t.TempDir(), "aosp")
	if _, err := workspaces.Add(root, "test", source, "", "", "", ""); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := Run(root, []string{"--workspace", "test"}, &stdout, &stderr); code == 0 || !strings.Contains(stderr.String(), "not ready") {
		t.Fatalf("code=%d err=%s", code, stderr.String())
	}
	if _, err := parseOptions([]string{"--module", "../bad"}, true); err == nil {
		t.Fatal("unsafe module accepted")
	}
	if _, err := parseOptions([]string{"--jobs", "0"}, true); err == nil {
		t.Fatal("zero jobs accepted")
	}
}
