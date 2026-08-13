package workspaces

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMultipleProfilesAndSelection(t *testing.T) {
	root := t.TempDir()
	path14, path15 := filepath.Join(t.TempDir(), "android-14"), filepath.Join(t.TempDir(), "android-15")
	active, err := Add(root, "android-14", path14, "", "android-14.0.0_r75", "aosp_x86_64-eng", "")
	if err != nil || !active {
		t.Fatalf("first Add = %v, %v", active, err)
	}
	if _, err := Add(root, "android-15", path15, "", "android-15.0.0_r1", "aosp_x86_64-eng", ""); err != nil {
		t.Fatal(err)
	}
	if err := Select(root, "android-15"); err != nil {
		t.Fatal(err)
	}
	profile, err := Current(root, "")
	if err != nil {
		t.Fatal(err)
	}
	if profile.Name != "android-15" || profile.Path != path15 {
		t.Fatalf("current = %#v", profile)
	}
	all, err := All(root)
	if err != nil || len(all) != 2 {
		t.Fatalf("All = %#v, %v", all, err)
	}
}

func TestDuplicatePathRejected(t *testing.T) {
	root, path := t.TempDir(), filepath.Join(t.TempDir(), "aosp")
	if _, err := Add(root, "one", path, "", "", "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := Add(root, "two", path, "", "", "", ""); err == nil {
		t.Fatal("duplicate path accepted")
	}
}

func TestInitNamedProfile(t *testing.T) {
	root, path := t.TempDir(), filepath.Join(t.TempDir(), "aosp")
	if _, err := Add(root, "test", path, "", "", "", ""); err != nil {
		t.Fatal(err)
	}
	var out, stderr bytes.Buffer
	if code := Run(root, []string{"init", "test"}, &out, &stderr); code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	if info, err := os.Stat(path); err != nil || !info.IsDir() {
		t.Fatalf("workspace not created: %v", err)
	}
	out.Reset()
	if code := Run(root, []string{"list"}, &out, &stderr); code != 0 || !strings.Contains(out.String(), "* test") {
		t.Fatalf("list=%q code=%d", out.String(), code)
	}
}
