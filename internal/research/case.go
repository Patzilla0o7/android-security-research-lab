package research

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var validCaseID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)

var allowedStatuses = map[string]bool{
	"investigating": true,
	"reproduced":    true,
	"patched":       true,
	"disclosed":     true,
	"closed":        true,
}

type Case struct {
	SchemaVersion     int
	CaseID            string
	Title             string
	Status            string
	Workspace         string
	AffectedComponent string
	DisclosureStatus  string
	CreatedAt         string
	UpdatedAt         string
}

func caseDir(root, id string) (string, error) {
	if !validCaseID.MatchString(id) {
		return "", fmt.Errorf("case ID must use letters, numbers, dot, dash or underscore")
	}
	return filepath.Join(root, "research", id), nil
}

func loadCase(root, id string) (Case, error) {
	dir, err := caseDir(root, id)
	if err != nil {
		return Case{}, err
	}
	file, err := os.Open(filepath.Join(dir, "case.yaml"))
	if os.IsNotExist(err) {
		return Case{}, fmt.Errorf("research case not found: %s", id)
	}
	if err != nil {
		return Case{}, err
	}
	defer file.Close()
	values := map[string]string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok || strings.TrimSpace(key) == "" {
			return Case{}, fmt.Errorf("invalid case.yaml line: %s", line)
		}
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if _, exists := values[key]; exists {
			return Case{}, fmt.Errorf("duplicate case.yaml field: %s", key)
		}
		if strings.HasPrefix(value, `"`) {
			decoded, err := strconv.Unquote(value)
			if err != nil {
				return Case{}, fmt.Errorf("invalid case.yaml value for %s", key)
			}
			value = decoded
		}
		values[key] = value
	}
	if err := scanner.Err(); err != nil {
		return Case{}, err
	}
	schema, err := strconv.Atoi(values["schema_version"])
	if err != nil {
		return Case{}, fmt.Errorf("invalid case schema_version")
	}
	c := Case{SchemaVersion: schema, CaseID: values["case_id"], Title: values["title"], Status: values["status"], Workspace: values["workspace"], AffectedComponent: values["affected_component"], DisclosureStatus: values["disclosure_status"], CreatedAt: values["created_at"], UpdatedAt: values["updated_at"]}
	if err := validateCase(c, id); err != nil {
		return Case{}, err
	}
	return c, nil
}

func validateCase(c Case, expectedID string) error {
	if c.SchemaVersion != 1 {
		return fmt.Errorf("unsupported case schema_version: %d", c.SchemaVersion)
	}
	if c.CaseID != expectedID || !validCaseID.MatchString(c.CaseID) {
		return fmt.Errorf("case_id does not match directory: %s", expectedID)
	}
	if strings.TrimSpace(c.Title) == "" {
		return fmt.Errorf("case title is required")
	}
	if !allowedStatuses[c.Status] {
		return fmt.Errorf("invalid case status: %s", c.Status)
	}
	if strings.TrimSpace(c.Workspace) == "" {
		return fmt.Errorf("case workspace is required")
	}
	if _, err := time.Parse(time.RFC3339, c.CreatedAt); err != nil {
		return fmt.Errorf("invalid created_at")
	}
	if _, err := time.Parse(time.RFC3339, c.UpdatedAt); err != nil {
		return fmt.Errorf("invalid updated_at")
	}
	return nil
}

func writeCase(dir string, c Case) error {
	content := fmt.Sprintf("schema_version: %d\ncase_id: %s\ntitle: %s\nstatus: %s\nworkspace: %s\naffected_component: %s\ndisclosure_status: %s\ncreated_at: %s\nupdated_at: %s\n", c.SchemaVersion, strconv.Quote(c.CaseID), strconv.Quote(c.Title), strconv.Quote(c.Status), strconv.Quote(c.Workspace), strconv.Quote(c.AffectedComponent), strconv.Quote(c.DisclosureStatus), strconv.Quote(c.CreatedAt), strconv.Quote(c.UpdatedAt))
	return os.WriteFile(filepath.Join(dir, "case.yaml"), []byte(content), 0o644)
}

func allCases(root string) ([]Case, error) {
	dir := filepath.Join(root, "research")
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var result []Case
	for _, entry := range entries {
		if !entry.IsDir() || !validCaseID.MatchString(entry.Name()) {
			continue
		}
		if _, err := os.Stat(filepath.Join(dir, entry.Name(), "case.yaml")); os.IsNotExist(err) {
			continue
		}
		c, err := loadCase(root, entry.Name())
		if err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CaseID < result[j].CaseID })
	return result, nil
}
