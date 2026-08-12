package config

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var keyPattern = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)

// Values is a parsed ASRL shell-compatible key/value configuration.
type Values map[string]string

// Load reads simple KEY=VALUE assignments. It intentionally rejects executable
// shell syntax so configuration remains data when consumed by Go.
func Load(path string, variables map[string]string) (Values, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	values := make(Values)
	scanner := bufio.NewScanner(file)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, raw, ok := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if !ok || !keyPattern.MatchString(key) {
			return nil, fmt.Errorf("%s:%d: invalid configuration assignment", path, lineNumber)
		}
		value := strings.TrimSpace(raw)
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}
		value = os.Expand(value, func(name string) string {
			if value, ok := variables[name]; ok {
				return value
			}
			return os.Getenv(name)
		})
		values[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return values, nil
}

func (v Values) Require(keys ...string) error {
	for _, key := range keys {
		if !keyPattern.MatchString(key) {
			return fmt.Errorf("invalid configuration key: %s", key)
		}
		if strings.TrimSpace(v[key]) == "" {
			return fmt.Errorf("missing required configuration: %s", key)
		}
	}
	return nil
}
