package workspace

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitAndStatus(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "config"), 0o755); err != nil {
		t.Fatal(err)
	}
	workspace := filepath.Join(t.TempDir(), "aosp")
	configPath := filepath.Join(root, "config", "lab.conf")
	if err := os.WriteFile(configPath, []byte("ANDROID_WORKSPACE=\""+workspace+"\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LAB_CONFIG_FILE", configPath)
	var stdout, stderr bytes.Buffer
	if code := Run(root, []string{"init"}, &stdout, &stderr); code != 0 {
		t.Fatalf("init code = %d: %s", code, stderr.String())
	}
	if info, err := os.Stat(workspace); err != nil || !info.IsDir() {
		t.Fatalf("workspace not created: %v", err)
	}
	stdout.Reset()
	stderr.Reset()
	if code := Run(root, []string{"status"}, &stdout, &stderr); code != 0 {
		t.Fatalf("status code = %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Directory     : ready") {
		t.Fatalf("unexpected status: %s", stdout.String())
	}
}

func TestInitRejectsProjectRoot(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "lab.conf")
	if err := os.WriteFile(configPath, []byte("ANDROID_WORKSPACE=\""+root+"\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LAB_CONFIG_FILE", configPath)
	var stdout, stderr bytes.Buffer
	if code := Run(root, []string{"init"}, &stdout, &stderr); code == 0 {
		t.Fatal("dangerous workspace accepted")
	}
}
