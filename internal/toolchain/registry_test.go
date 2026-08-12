package toolchain

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tools.conf")
	data := "TOOL_SPECS=(\n    \"java|Java|required|java-major|java|17|apt:openjdk-17-jdk\"\n    \"repo|Repo|recommended|command|repo||apt:repo,manual:https://example.com/repo\"\n)\n"
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	tools, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(tools) != 2 || AptPackage(tools[0]) != "openjdk-17-jdk" || ManualMethod(tools[1]) == "" {
		t.Fatalf("unexpected tools: %#v", tools)
	}
}

func TestJavaMajor(t *testing.T) {
	if got := javaMajor(`openjdk version "17.0.19" 2026-04-21`); got != 17 {
		t.Fatalf("javaMajor = %d", got)
	}
}
