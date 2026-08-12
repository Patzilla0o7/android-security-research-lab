package safety

import "testing"

func TestPaths(t *testing.T) {
	if err := RequireAbsolute("workspace", "/tmp/asrl"); err != nil {
		t.Fatal(err)
	}
	if err := RequireAbsolute("workspace", "relative"); err == nil {
		t.Fatal("relative path accepted")
	}
	if !IsWithin("/tmp/asrl", "/tmp/asrl") || !IsWithin("/tmp/asrl/child", "/tmp/asrl") {
		t.Fatal("child not detected")
	}
	if IsWithin("/tmp/other", "/tmp/asrl") {
		t.Fatal("unrelated path accepted")
	}
	if err := RefuseDangerous("workspace", "/opt/asrl", "/opt/asrl"); err == nil {
		t.Fatal("project root accepted")
	}
}
