package repo

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Patzilla0o7/android-security-research-lab/internal/workspaces"
)

func researchFixture(t *testing.T) (string, string, string) {
	t.Helper()
	root := t.TempDir()
	workspace := filepath.Join(t.TempDir(), "aosp")
	project := filepath.Join(workspace, "frameworks", "base")
	if _, err := workspaces.Add(root, "android-14", workspace, "", "tag", "target", ""); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(workspace, ".repo"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(project, 0755); err != nil {
		t.Fatal(err)
	}
	runGit(t, project, "init")
	runGit(t, project, "config", "user.name", "ASRL Test")
	runGit(t, project, "config", "user.email", "test@example.com")
	if err := os.WriteFile(filepath.Join(project, "file.txt"), []byte("base\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, project, "add", ".")
	runGit(t, project, "commit", "-m", "base")
	return root, workspace, project
}

func TestBranchCreatePlanAndApply(t *testing.T) {
	root, _, _ := researchFixture(t)
	bin := t.TempDir()
	record := filepath.Join(t.TempDir(), "calls")
	script := "#!/usr/bin/env bash\nprintf '%s\\n' \"$*\" >> \"$BRANCH_RECORD\"\n"
	if err := os.WriteFile(filepath.Join(bin, "repo"), []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("BRANCH_RECORD", record)
	var out, stderr bytes.Buffer
	args := []string{"branch", "create", "cve-test", "--workspace", "android-14", "--project", "frameworks/base"}
	if code := Run(root, args, &out, &stderr); code != 0 {
		t.Fatalf("plan: %s", stderr.String())
	}
	if _, err := os.Stat(record); !os.IsNotExist(err) {
		t.Fatal("plan executed repo")
	}
	args = append(args, "--apply")
	if code := Run(root, args, &out, &stderr); code != 0 {
		t.Fatalf("apply: %s", stderr.String())
	}
	data, _ := os.ReadFile(record)
	if !strings.Contains(string(data), "start cve-test frameworks/base") {
		t.Fatalf("call=%q", data)
	}
}

func TestPatchExportAndImport(t *testing.T) {
	root, _, project := researchFixture(t)
	oldNow := now
	now = func() time.Time { return time.Date(2026, 8, 13, 13, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { now = oldNow })
	if err := os.WriteFile(filepath.Join(project, "file.txt"), []byte("changed\n"), 0644); err != nil {
		t.Fatal(err)
	}
	var out, stderr bytes.Buffer
	if code := Run(root, []string{"patch", "export", "--workspace", "android-14", "--project", "frameworks/base"}, &out, &stderr); code != 0 {
		t.Fatalf("export: %s", stderr.String())
	}
	dir := filepath.Join(root, "output", "repo", "android-14", "patches", "20260813-130000", "frameworks__base")
	for _, name := range []string{"working-tree.diff", "metadata.json", "SHA256SUMS"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatal(err)
		}
	}
	runGit(t, project, "checkout", "--", "file.txt")
	patch := filepath.Join(dir, "working-tree.diff")
	out.Reset()
	stderr.Reset()
	args := []string{"patch", "import", "--workspace", "android-14", "--project", "frameworks/base", "--file", patch}
	if code := Run(root, args, &out, &stderr); code != 0 {
		t.Fatalf("check: %s", stderr.String())
	}
	data, _ := os.ReadFile(filepath.Join(project, "file.txt"))
	if string(data) != "base\n" {
		t.Fatal("plan modified file")
	}
	args = append(args, "--apply")
	if code := Run(root, args, &out, &stderr); code != 0 {
		t.Fatalf("apply: %s", stderr.String())
	}
	data, _ = os.ReadFile(filepath.Join(project, "file.txt"))
	if string(data) != "changed\n" {
		t.Fatalf("content=%q", data)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if data, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %s", args, data)
	}
}
