package evidence

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func bundleFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	manifest := []byte(`{"schema_version":1,"operation":"bundle","case_id":"case-1","workspace":"android-15","serial":"emulator-5554","collected_at":"2026-08-21T12:00:00Z","status":"success"}` + "\n")
	logcat := []byte("test log\n")
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), manifest, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "logcat.txt"), logcat, 0o644); err != nil {
		t.Fatal(err)
	}
	manifestSum, logcatSum := sha256.Sum256(manifest), sha256.Sum256(logcat)
	checksums := fmt.Sprintf("%x  logcat.txt\n%x  manifest.json\n", logcatSum, manifestSum)
	if err := os.WriteFile(filepath.Join(dir, "SHA256SUMS"), []byte(checksums), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestInspectAndVerify(t *testing.T) {
	dir := bundleFixture(t)
	report, err := Verify(dir)
	if err != nil {
		t.Fatal(err)
	}
	if report.Manifest.CaseID != "case-1" || report.Checksums != 2 || len(report.Files) != 3 {
		t.Fatalf("unexpected report: %#v", report)
	}
	digest, err := ManifestDigest(dir)
	if err != nil || len(digest) != 64 {
		t.Fatalf("digest=%q err=%v", digest, err)
	}
}

func TestVerifyRejectsTamperAndUnexpectedFile(t *testing.T) {
	dir := bundleFixture(t)
	if err := os.WriteFile(filepath.Join(dir, "logcat.txt"), []byte("changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Verify(dir); err == nil || !strings.Contains(err.Error(), "mismatch") {
		t.Fatalf("unexpected error: %v", err)
	}
	dir = bundleFixture(t)
	if err := os.WriteFile(filepath.Join(dir, "extra.txt"), []byte("extra"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Verify(dir); err == nil || !strings.Contains(err.Error(), "unverified") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVerifyRejectsUnsafeDuplicateAndSymlink(t *testing.T) {
	dir := bundleFixture(t)
	bad := strings.Repeat("0", 64) + "  ../outside\n"
	if err := os.WriteFile(filepath.Join(dir, "SHA256SUMS"), []byte(bad), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Verify(dir); err == nil || !strings.Contains(err.Error(), "unsafe") {
		t.Fatalf("unexpected error: %v", err)
	}
	dir = bundleFixture(t)
	data, err := os.ReadFile(filepath.Join(dir, "SHA256SUMS"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SHA256SUMS"), append(data, data...), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Verify(dir); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("unexpected error: %v", err)
	}
	dir = bundleFixture(t)
	if err := os.Symlink("logcat.txt", filepath.Join(dir, "link.txt")); err != nil {
		t.Fatal(err)
	}
	if _, err := Inspect(dir); err == nil || !strings.Contains(err.Error(), "non-regular") {
		t.Fatalf("unexpected error: %v", err)
	}
}
