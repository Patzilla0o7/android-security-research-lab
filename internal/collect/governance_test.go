package collect

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func governanceBundle(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string][]byte{
		"manifest.json":  []byte(`{"schema_version":1,"operation":"bundle","case_id":"case-1","workspace":"android-15","serial":"emulator-5554","collected_at":"2026-08-21T12:00:00Z","status":"success"}` + "\n"),
		"logcat.txt":     []byte("account=user@example.com ip=192.168.1.2 Authorization: Bearer-secret\n"),
		"screenshot.png": append([]byte("\x89PNG\r\n\x1a\n"), []byte("image")...),
	}
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var lines []string
	for _, name := range []string{"logcat.txt", "manifest.json", "screenshot.png"} {
		sum := sha256.Sum256(files[name])
		lines = append(lines, fmt.Sprintf("%x  %s", sum, name))
	}
	if err := os.WriteFile(filepath.Join(dir, "SHA256SUMS"), []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestGovernanceInspectVerifyAndRedactPlan(t *testing.T) {
	dir := governanceBundle(t)
	for _, args := range [][]string{{"inspect", dir}, {"verify", dir}, {"redact", dir, "--plan"}} {
		var stdout, stderr strings.Builder
		if code := run("", args, &stdout, &stderr, dependencies{}); code != 0 {
			t.Fatalf("args=%v code=%d stderr=%s", args, code, stderr.String())
		}
		if args[0] == "redact" {
			for _, expected := range []string{"email", "ipv4", "credential", "Manual review", "No files were changed"} {
				if !strings.Contains(stdout.String(), expected) {
					t.Fatalf("redact output missing %q: %s", expected, stdout.String())
				}
			}
		}
	}
}

func TestRedactPlanRequiresVerifiedBundle(t *testing.T) {
	dir := governanceBundle(t)
	if err := os.WriteFile(filepath.Join(dir, "logcat.txt"), []byte("tampered"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr strings.Builder
	if code := run("", []string{"redact", dir, "--plan"}, &stdout, &stderr, dependencies{}); code != 1 || !strings.Contains(stderr.String(), "verify bundle") {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
}

func TestScanZipAndTar(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "bugreport.zip")
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(zipFile)
	entry, err := zw.Create("report.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("ssid=ResearchWifi user=test@example.com")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := zipFile.Close(); err != nil {
		t.Fatal(err)
	}

	var tarData bytes.Buffer
	tw := tar.NewWriter(&tarData)
	content := []byte("imei=123456789012345")
	if err := tw.WriteHeader(&tar.Header{Name: "tombstone_00", Mode: 0o600, Size: int64(len(content))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tombstones.tar"), tarData.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	if err := scanZip(zipPath, "bugreport.zip", counts); err != nil {
		t.Fatal(err)
	}
	if err := scanTar(filepath.Join(dir, "tombstones.tar"), "tombstones.tar", counts); err != nil {
		t.Fatal(err)
	}
	if len(counts) < 3 {
		t.Fatalf("expected archive findings, got %#v", counts)
	}
}
