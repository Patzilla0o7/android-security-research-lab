package collect

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/Patzilla0o7/android-security-research-lab/internal/evidence"
)

const maxScanEntry = 16 << 20

type redactRule struct {
	name    string
	pattern *regexp.Regexp
}

type finding struct {
	File, Rule string
	Count      int
}

var redactRules = []redactRule{
	{"email", regexp.MustCompile(`(?i)\b[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}\b`)},
	{"ipv4", regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)},
	{"mac_address", regexp.MustCompile(`(?i)\b[0-9a-f]{2}(?::[0-9a-f]{2}){5}\b`)},
	{"phone_number", regexp.MustCompile(`\b(?:\+?[0-9][0-9 ()-]{7,}[0-9])\b`)},
	{"credential", regexp.MustCompile(`(?i)\b(?:authorization|cookie|token|password|passwd|secret)\s*[:=]\s*[^\s,;]+`)},
	{"android_identifier", regexp.MustCompile(`(?i)\b(?:android[_ ]?id|imei|imsi|serial(?:no)?)\s*[:=]\s*[A-Z0-9._-]+`)},
	{"wifi_identifier", regexp.MustCompile(`(?i)\b(?:ssid|bssid)\s*[:=]\s*[^\s,;]+`)},
	{"user_path", regexp.MustCompile(`(?:/home/[A-Za-z0-9._-]+|/data/user/[0-9]+/[A-Za-z0-9._-]+)`)},
}

func governanceCommand(operation string, args []string, stdout, stderr io.Writer) int {
	if operation == "inspect" || operation == "verify" {
		if len(args) != 1 {
			return usageError(stderr, operation+" requires one bundle directory")
		}
		report, err := evidence.Inspect(args[0])
		if err != nil {
			return fail(stderr, err)
		}
		printInspection(stdout, report)
		verified, verifyErr := evidence.Verify(args[0])
		if verifyErr != nil {
			fmt.Fprintf(stderr, "[FAIL] Integrity: %v\n", verifyErr)
			return 1
		}
		fmt.Fprintf(stdout, "Integrity      : verified (%d checksums)\n", verified.Checksums)
		return 0
	}
	if operation == "redact" {
		if len(args) != 2 || args[1] != "--plan" {
			return usageError(stderr, "redact requires <bundle> --plan; --apply is not available yet")
		}
		report, err := evidence.Verify(args[0])
		if err != nil {
			return fail(stderr, fmt.Errorf("verify bundle before redaction planning: %w", err))
		}
		findings, manual, err := scanBundle(report.Directory)
		if err != nil {
			return fail(stderr, err)
		}
		fmt.Fprintf(stdout, "Redaction plan : %s\nIntegrity      : verified\n", report.Directory)
		if len(findings) == 0 {
			fmt.Fprintln(stdout, "Sensitive text : no configured patterns detected")
		} else {
			fmt.Fprintln(stdout, "Sensitive text :")
			for _, item := range findings {
				fmt.Fprintf(stdout, "  %s\t%s\t%d match(es)\n", item.File, item.Rule, item.Count)
			}
		}
		if len(manual) > 0 {
			fmt.Fprintln(stdout, "Manual review  :")
			for _, name := range manual {
				fmt.Fprintf(stdout, "  %s\n", name)
			}
		}
		fmt.Fprintln(stdout, "No files were changed. A future --apply mode will create a separate copy.")
		return 0
	}
	return usageError(stderr, "unknown governance operation: "+operation)
}

func printInspection(w io.Writer, report evidence.Report) {
	fmt.Fprintf(w, "Bundle         : %s\nCase           : %s\nWorkspace      : %s\nSerial         : %s\nOperation      : %s\nStatus         : %s\nCollected      : %s\nFiles          : %d\nTotal size     : %d bytes\n", report.Directory, report.Manifest.CaseID, report.Manifest.Workspace, report.Manifest.Serial, report.Manifest.Operation, report.Manifest.Status, report.Manifest.CollectedAt, len(report.Files), report.TotalSize)
	for _, file := range report.Files {
		fmt.Fprintf(w, "  %-24s %d bytes\n", file.Name, file.Size)
	}
}

func scanBundle(dir string) ([]finding, []string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, err
	}
	counts := map[string]int{}
	manual := []string{}
	for _, entry := range entries {
		name := entry.Name()
		path := filepath.Join(dir, name)
		switch strings.ToLower(filepath.Ext(name)) {
		case ".png", ".jpg", ".jpeg", ".webp":
			manual = append(manual, name+" (image content)")
		case ".zip":
			if err := scanZip(path, name, counts); err != nil {
				return nil, nil, err
			}
		case ".tar":
			if err := scanTar(path, name, counts); err != nil {
				return nil, nil, err
			}
		default:
			if name == "SHA256SUMS" {
				continue
			}
			data, err := readLimited(path, maxScanEntry)
			if err != nil {
				return nil, nil, err
			}
			scanText(name, data, counts)
		}
	}
	findings := make([]finding, 0, len(counts))
	for key, count := range counts {
		file, rule, _ := strings.Cut(key, "\x00")
		findings = append(findings, finding{File: file, Rule: rule, Count: count})
	}
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].File == findings[j].File {
			return findings[i].Rule < findings[j].Rule
		}
		return findings[i].File < findings[j].File
	})
	sort.Strings(manual)
	return findings, manual, nil
}

func scanZip(path, label string, counts map[string]int) error {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("inspect ZIP %s: %w", label, err)
	}
	defer reader.Close()
	for _, file := range reader.File {
		if file.FileInfo().IsDir() || file.UncompressedSize64 > maxScanEntry {
			continue
		}
		stream, err := file.Open()
		if err != nil {
			return err
		}
		data, err := io.ReadAll(io.LimitReader(stream, maxScanEntry+1))
		stream.Close()
		if err != nil {
			return err
		}
		scanText(label+"!"+file.Name, data, counts)
	}
	return nil
}

func scanTar(path, label string, counts map[string]int) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	reader := tar.NewReader(file)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("inspect TAR %s: %w", label, err)
		}
		if header.Typeflag != tar.TypeReg || header.Size > maxScanEntry {
			continue
		}
		data, err := io.ReadAll(io.LimitReader(reader, maxScanEntry+1))
		if err != nil {
			return err
		}
		scanText(label+"!"+header.Name, data, counts)
	}
}

func readLimited(path string, limit int64) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(io.LimitReader(file, limit+1))
}

func scanText(label string, data []byte, counts map[string]int) {
	if len(data) > maxScanEntry || !utf8.Valid(data) || bytes.IndexByte(data, 0) >= 0 {
		return
	}
	for _, rule := range redactRules {
		if count := len(rule.pattern.FindAllIndex(data, -1)); count > 0 {
			counts[label+"\x00"+rule.name] += count
		}
	}
}
