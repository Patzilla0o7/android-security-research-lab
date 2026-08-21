package evidence

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Manifest struct {
	SchemaVersion int    `json:"schema_version"`
	Operation     string `json:"operation"`
	Workspace     string `json:"workspace"`
	CaseID        string `json:"case_id"`
	Serial        string `json:"serial"`
	CollectedAt   string `json:"collected_at"`
	Status        string `json:"status"`
}

type File struct {
	Name string
	Size int64
}

type Report struct {
	Directory  string
	Manifest   Manifest
	Files      []File
	TotalSize  int64
	Checksums  int
	Unexpected []string
}

func Inspect(requested string) (Report, error) {
	dir, err := filepath.Abs(requested)
	if err != nil {
		return Report{}, err
	}
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return Report{}, fmt.Errorf("evidence bundle is not a directory: %s", dir)
	}
	data, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return Report{}, fmt.Errorf("read evidence manifest: %w", err)
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Report{}, fmt.Errorf("invalid evidence manifest: %w", err)
	}
	if manifest.SchemaVersion != 1 {
		return Report{}, fmt.Errorf("unsupported evidence schema_version: %d", manifest.SchemaVersion)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return Report{}, err
	}
	report := Report{Directory: dir, Manifest: manifest}
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return Report{}, err
		}
		if !info.Mode().IsRegular() {
			return Report{}, fmt.Errorf("bundle contains non-regular entry: %s", entry.Name())
		}
		report.Files = append(report.Files, File{Name: entry.Name(), Size: info.Size()})
		report.TotalSize += info.Size()
	}
	sort.Slice(report.Files, func(i, j int) bool { return report.Files[i].Name < report.Files[j].Name })
	return report, nil
}

func Verify(requested string) (Report, error) {
	report, err := Inspect(requested)
	if err != nil {
		return report, err
	}
	checksumPath := filepath.Join(report.Directory, "SHA256SUMS")
	if info, err := os.Lstat(checksumPath); err != nil || !info.Mode().IsRegular() {
		return report, fmt.Errorf("SHA256SUMS is missing or not a regular file")
	}
	file, err := os.Open(checksumPath)
	if err != nil {
		return report, err
	}
	defer file.Close()
	covered := map[string]bool{"SHA256SUMS": true}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 || len(fields[0]) != 64 {
			return report, fmt.Errorf("invalid SHA256SUMS entry: %s", scanner.Text())
		}
		if _, err := hex.DecodeString(fields[0]); err != nil {
			return report, fmt.Errorf("invalid SHA-256 digest for %s", fields[1])
		}
		name := filepath.ToSlash(filepath.Clean(filepath.FromSlash(fields[1])))
		if name != fields[1] || filepath.IsAbs(fields[1]) || name == ".." || strings.HasPrefix(name, "../") {
			return report, fmt.Errorf("unsafe SHA256SUMS path: %s", fields[1])
		}
		if covered[name] {
			return report, fmt.Errorf("duplicate SHA256SUMS entry: %s", name)
		}
		path := filepath.Join(report.Directory, filepath.FromSlash(name))
		info, err := os.Lstat(path)
		if err != nil || !info.Mode().IsRegular() {
			return report, fmt.Errorf("evidence file is missing or not regular: %s", name)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return report, err
		}
		sum := sha256.Sum256(data)
		if !strings.EqualFold(fields[0], hex.EncodeToString(sum[:])) {
			return report, fmt.Errorf("SHA-256 mismatch: %s", name)
		}
		covered[name] = true
		report.Checksums++
	}
	if err := scanner.Err(); err != nil {
		return report, err
	}
	if report.Checksums == 0 {
		return report, fmt.Errorf("SHA256SUMS contains no entries")
	}
	if !covered["manifest.json"] {
		return report, fmt.Errorf("SHA256SUMS does not cover manifest.json")
	}
	for _, file := range report.Files {
		if !covered[file.Name] {
			report.Unexpected = append(report.Unexpected, file.Name)
		}
	}
	if len(report.Unexpected) > 0 {
		return report, fmt.Errorf("bundle contains unverified files: %s", strings.Join(report.Unexpected, ", "))
	}
	return report, nil
}

func ManifestDigest(dir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}
