package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAndExpand(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lab.conf")
	data := "ANDROID_WORKSPACE=\"${LAB_ROOT}/workspace/aosp\"\nANDROID_BRANCH='main'\n"
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	values, err := Load(path, map[string]string{"LAB_ROOT": "/opt/asrl"})
	if err != nil {
		t.Fatal(err)
	}
	if values["ANDROID_WORKSPACE"] != "/opt/asrl/workspace/aosp" {
		t.Fatalf("workspace = %q", values["ANDROID_WORKSPACE"])
	}
	if err := values.Require("ANDROID_WORKSPACE", "ANDROID_BRANCH"); err != nil {
		t.Fatal(err)
	}
}

func TestRejectsShellCode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lab.conf")
	if err := os.WriteFile(path, []byte("source /tmp/unsafe\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path, nil); err == nil {
		t.Fatal("Load accepted executable shell syntax")
	}
}
