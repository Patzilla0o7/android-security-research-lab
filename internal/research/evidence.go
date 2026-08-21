package research

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Patzilla0o7/android-security-research-lab/internal/evidence"
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
	report, err := evidence.Verify(bundle)
	if err != nil {
		return evidenceRecord{}, err
	}
	if report.Manifest.CaseID != id {
		return evidenceRecord{}, fmt.Errorf("evidence case_id %s does not match research case %s", report.Manifest.CaseID, id)
	}
	stored := bundle
	if relative, err := filepath.Rel(root, bundle); err == nil && relative != ".." && !strings.HasPrefix(relative, "../") {
		stored = filepath.ToSlash(relative)
	}
	digest, err := evidence.ManifestDigest(bundle)
	if err != nil {
		return evidenceRecord{}, err
	}
	return evidenceRecord{Bundle: stored, Workspace: report.Manifest.Workspace, Serial: report.Manifest.Serial, CollectedAt: report.Manifest.CollectedAt, Status: report.Manifest.Status, ManifestSHA256: digest}, nil
}

func resolveBundle(root, stored string) string {
	if filepath.IsAbs(stored) {
		return filepath.Clean(stored)
	}
	return filepath.Join(root, filepath.FromSlash(stored))
}
