package repo

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Patzilla0o7/android-security-research-lab/internal/workspaces"
)

func TestStatusInitAndSync(t *testing.T) {
	root := t.TempDir()
	workspacePath := filepath.Join(t.TempDir(), "aosp")
	if _, err := workspaces.Add(root, "android-14", workspacePath, "https://example.com/manifest", "android-14-tag", "target", ""); err != nil {
		t.Fatal(err)
	}
	binDir := t.TempDir()
	record := filepath.Join(t.TempDir(), "repo-calls")
	fakeRepo := filepath.Join(binDir, "repo")
	script := "#!/usr/bin/env bash\nset -euo pipefail\nprintf '%s|%s\\n' \"$PWD\" \"$*\" >> \"$REPO_TEST_RECORD\"\nif [[ \"${1:-}\" == init ]]; then mkdir -p .repo; touch .repo/manifest.xml; fi\nif [[ \"${1:-}\" == sync ]]; then echo 'Fetching: 50% (1/2)' >&2; fi\n"
	if err := os.WriteFile(fakeRepo, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("REPO_TEST_RECORD", record)
	oldNow := now
	now = func() time.Time { return time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { now = oldNow })
	var stdout, stderr bytes.Buffer
	if code := Run(root, []string{"status", "--workspace", "android-14"}, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "not initialized") {
		t.Fatalf("status code=%d out=%s err=%s", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := Run(root, []string{"init", "--workspace", "android-14", "--partial-clone", "--clone-filter", "blob:limit=10M", "--no-use-superproject", "--apply"}, &stdout, &stderr); code != 0 {
		t.Fatalf("init code=%d out=%s err=%s", code, stdout.String(), stderr.String())
	}
	data, err := os.ReadFile(record)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), workspacePath+"|init -u https://example.com/manifest -b android-14-tag --partial-clone --clone-filter=blob:limit=10M --no-use-superproject") {
		t.Fatalf("init call=%q", data)
	}
	if _, err := os.Stat(filepath.Join(root, "output", "repo", "android-14", "20260813-120000-init.log")); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	if code := Run(root, []string{"sync", "--workspace", "android-14", "--jobs", "4"}, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "No changes were made") {
		t.Fatalf("plan code=%d out=%s err=%s", code, stdout.String(), stderr.String())
	}
	before, _ := os.ReadFile(record)
	stdout.Reset()
	stderr.Reset()
	if code := Run(root, []string{"sync", "--workspace", "android-14", "--jobs", "4", "--project", "frameworks/base", "--retry-fetches", "3", "--no-clone-bundle", "--force-sync", "--apply"}, &stdout, &stderr); code != 0 {
		t.Fatalf("sync code=%d out=%s err=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "Fetching: 50% (1/2)") {
		t.Fatalf("live progress was not forwarded: %q", stderr.String())
	}
	after, _ := os.ReadFile(record)
	if len(after) <= len(before) || !strings.Contains(string(after), "sync -c -j 4 --retry-fetches=3 --no-clone-bundle --force-sync frameworks/base") {
		t.Fatalf("sync call=%q", after)
	}
}

func TestSyncRequiresInitialization(t *testing.T) {
	root, path := t.TempDir(), filepath.Join(t.TempDir(), "aosp")
	if _, err := workspaces.Add(root, "test", path, "", "", "", ""); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := Run(root, []string{"sync"}, &stdout, &stderr); code == 0 || !strings.Contains(stderr.String(), "not initialized") {
		t.Fatalf("code=%d err=%s", code, stderr.String())
	}
}
