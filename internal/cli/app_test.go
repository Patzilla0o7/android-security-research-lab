package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"--help"}, &stdout, &stderr); code != exitSuccess {
		t.Fatalf("Run() code = %d", code)
	}
	if !strings.Contains(stdout.String(), "Android Security Research Lab") {
		t.Fatalf("help output missing project name: %q", stdout.String())
	}
}

func TestUnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"unknown-command"}, &stdout, &stderr); code != exitUsage {
		t.Fatalf("Run() code = %d, want %d", code, exitUsage)
	}
	if !strings.Contains(stderr.String(), "Unknown command") {
		t.Fatalf("error output missing message: %q", stderr.String())
	}
}

func TestPlaceholder(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"build"}, &stdout, &stderr); code != exitNotImplemented {
		t.Fatalf("Run() code = %d, want %d", code, exitNotImplemented)
	}
}

func TestVersion(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "VERSION"), []byte("test-version\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "config"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ASRL_ROOT", root)
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"version"}, &stdout, &stderr); code != exitSuccess {
		t.Fatalf("Run() code = %d; stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Version : test-version") {
		t.Fatalf("version output = %q", stdout.String())
	}
}
