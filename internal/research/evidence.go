package research

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type evidenceRecord struct {
	Bundle         string `json:"bundle"`
	Workspace      string `json:"workspace"`
	Serial         string `json:"serial"`
	CollectedAt    string `json:"collected_at"`
	Status         string `json:"status"`
	ManifestSHA256 string `json:"manifest_sha256"`
}

type evidenceIndex struct {
	SchemaVersion int              `json:"schema_version"`
	Evidence      []evidenceRecord `json:"evidence"`
}

type collectManifest struct {
	Workspace   string `json:"workspace"`
	CaseID      string `json:"case_id"`
	Serial      string `json:"serial"`
	CollectedAt string `json:"collected_at"`
	Status      string `json:"status"`
}

func evidenceFile(root, id string) (string, error) {
	dir, err := caseDir(root, id)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "artifacts", "evidence.json"), nil
}

func loadEvidence(root, id string) (evidenceIndex, error) {
	path, err := evidenceFile(root, id)
	if err != nil {
		return evidenceIndex{}, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return evidenceIndex{SchemaVersion: 1, Evidence: []evidenceRecord{}}, nil
	}
	if err != nil {
		return evidenceIndex{}, err
	}
	var index evidenceIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return evidenceIndex{}, fmt.Errorf("invalid evidence index: %w", err)
	}
	if index.SchemaVersion != 1 {
		return evidenceIndex{}, fmt.Errorf("unsupported evidence schema_version: %d", index.SchemaVersion)
	}
	return index, nil
}

func saveEvidence(root, id string, index evidenceIndex) error {
	path, err := evidenceFile(root, id)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}
	temp := path + ".tmp"
	if err := os.WriteFile(temp, append(data, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(temp, path)
}

func inspectBundle(root, id, requested string) (evidenceRecord, error) {
	bundle, err := filepath.Abs(requested)
	if err != nil {
		return evidenceRecord{}, err
	}
	info, err := os.Stat(bundle)
	if err != nil || !info.IsDir() {
		return evidenceRecord{}, fmt.Errorf("evidence bundle is not a directory: %s", bundle)
	}
	if err := verifyBundle(bundle); err != nil {
		return evidenceRecord{}, err
	}
	manifestPath := filepath.Join(bundle, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return evidenceRecord{}, err
	}
	var m collectManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return evidenceRecord{}, fmt.Errorf("invalid evidence manifest: %w", err)
	}
	if m.CaseID != id {
		return evidenceRecord{}, fmt.Errorf("evidence case_id %s does not match research case %s", m.CaseID, id)
	}
	stored := bundle
	if relative, err := filepath.Rel(root, bundle); err == nil && relative != ".." && !strings.HasPrefix(relative, "../") {
		stored = filepath.ToSlash(relative)
	}
	sum := sha256.Sum256(data)
	return evidenceRecord{Bundle: stored, Workspace: m.Workspace, Serial: m.Serial, CollectedAt: m.CollectedAt, Status: m.Status, ManifestSHA256: hex.EncodeToString(sum[:])}, nil
}

func resolveBundle(root, stored string) string {
	if filepath.IsAbs(stored) {
		return filepath.Clean(stored)
	}
	return filepath.Join(root, filepath.FromSlash(stored))
}

func verifyBundle(bundle string) error {
	checksumPath := filepath.Join(bundle, "SHA256SUMS")
	if info, err := os.Lstat(checksumPath); err != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("SHA256SUMS is missing or not a regular file")
	}
	file, err := os.Open(checksumPath)
	if err != nil {
		return err
	}
	defer file.Close()
	count, foundManifest := 0, false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 || len(fields[0]) != 64 {
			return fmt.Errorf("invalid SHA256SUMS entry: %s", scanner.Text())
		}
		name := fields[1]
		if name == "manifest.json" {
			foundManifest = true
		}
		if filepath.IsAbs(name) || name == ".." || strings.HasPrefix(name, "../") || strings.Contains(name, `/../`) {
			return fmt.Errorf("unsafe SHA256SUMS path: %s", name)
		}
		path := filepath.Join(bundle, filepath.FromSlash(name))
		info, err := os.Lstat(path)
		if err != nil || !info.Mode().IsRegular() {
			return fmt.Errorf("evidence file is missing or not regular: %s", name)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		if !strings.EqualFold(fields[0], hex.EncodeToString(sum[:])) {
			return fmt.Errorf("SHA-256 mismatch: %s", name)
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("SHA256SUMS contains no entries")
	}
	if !foundManifest {
		return fmt.Errorf("SHA256SUMS does not cover manifest.json")
	}
	return nil
}
