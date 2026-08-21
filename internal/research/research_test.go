package research

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Patzilla0o7/android-security-research-lab/internal/workspaces"
)

func researchFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if _, err := workspaces.Add(root, "android-15", filepath.Join(t.TempDir(), "aosp"), "", "android-15.0.0_r1", "aosp_x86_64-eng", ""); err != nil {
		t.Fatal(err)
	}
	previous := now
	now = func() time.Time { return time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { now = previous })
	return root
}

func createCase(t *testing.T, root, id string) {
	t.Helper()
	var stdout, stderr strings.Builder
	if code := Run(root, []string{"new", id, "--title", "Binder test"}, &stdout, &stderr); code != 0 {
		t.Fatalf("new code=%d stderr=%s", code, stderr.String())
	}
}

func TestNewListShowAndValidate(t *testing.T) {
	root := researchFixture(t)
	createCase(t, root, "CVE-2026-0001")
	dir := filepath.Join(root, "research", "CVE-2026-0001")
	for _, name := range []string{"case.yaml", "README.md", "timeline.md", "reproduction.md", "root-cause.md", "patches", "poc", "reports"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
	var stdout, stderr strings.Builder
	if code := Run(root, []string{"list"}, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "CVE-2026-0001") {
		t.Fatalf("list code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	if code := Run(root, []string{"show", "CVE-2026-0001"}, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "Binder test") {
		t.Fatalf("show code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	if code := Run(root, []string{"validate", "CVE-2026-0001"}, &stdout, &stderr); code != 0 {
		t.Fatalf("validate code=%d stderr=%s", code, stderr.String())
	}
}

func TestNewRejectsUnsafeAndDuplicateCases(t *testing.T) {
	root := researchFixture(t)
	var stdout, stderr strings.Builder
	if code := Run(root, []string{"new", "../escape", "--title", "bad"}, &stdout, &stderr); code != 2 {
		t.Fatalf("unsafe code=%d", code)
	}
	createCase(t, root, "case-1")
	stderr.Reset()
	if code := Run(root, []string{"new", "case-1", "--title", "duplicate"}, &stdout, &stderr); code != 1 {
		t.Fatalf("duplicate code=%d stderr=%s", code, stderr.String())
	}
}

func TestValidateDetectsMissingTemplate(t *testing.T) {
	root := researchFixture(t)
	createCase(t, root, "case-1")
	if err := os.Remove(filepath.Join(root, "research", "case-1", "root-cause.md")); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr strings.Builder
	if code := Run(root, []string{"validate", "case-1"}, &stdout, &stderr); code != 1 || !strings.Contains(stderr.String(), "root-cause.md") {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
}
