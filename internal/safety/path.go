package safety

import (
	"fmt"
	"path/filepath"
	"strings"
)

func RequireAbsolute(label, path string) error {
	if path == "" || !filepath.IsAbs(path) {
		return fmt.Errorf("%s must be an absolute path: %s", label, displayPath(path))
	}
	return nil
}

func IsWithin(path, parent string) bool {
	relative, err := filepath.Rel(filepath.Clean(parent), filepath.Clean(path))
	if err != nil {
		return false
	}
	return relative == "." || (relative != ".." && !strings.HasPrefix(relative, "../"))
}

func RefuseDangerous(label, path, projectRoot string) error {
	if err := RequireAbsolute(label, path); err != nil {
		return err
	}
	clean := filepath.Clean(path)
	root := filepath.Clean(projectRoot)
	dangerous := map[string]bool{
		string(filepath.Separator):      true,
		root:                            true,
		filepath.Join(root, "config"):   true,
		filepath.Join(root, "docs"):     true,
		filepath.Join(root, "research"): true,
	}
	if dangerous[clean] {
		return fmt.Errorf("refusing unsafe %s: %s", label, clean)
	}
	return nil
}

func displayPath(path string) string {
	if path == "" {
		return "<empty>"
	}
	return path
}
