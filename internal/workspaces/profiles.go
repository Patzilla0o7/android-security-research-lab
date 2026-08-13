package workspaces

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Patzilla0o7/android-security-research-lab/internal/config"
	"github.com/Patzilla0o7/android-security-research-lab/internal/safety"
)

var validName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)

type Profile struct {
	Name, Path, Manifest, Branch, Target, JavaHome, Ccache, File string
}

func profilesDir(root string) string { return filepath.Join(root, "config", "workspaces") }
func activeFile(root string) string  { return filepath.Join(root, ".local", "active-workspace") }

func Load(root, name string) (Profile, error) {
	if !validName.MatchString(name) {
		return Profile{}, fmt.Errorf("invalid workspace name: %s", name)
	}
	file := filepath.Join(profilesDir(root), name+".conf")
	values, err := config.Load(file, map[string]string{"LAB_ROOT": root})
	if os.IsNotExist(err) {
		return Profile{}, fmt.Errorf("workspace profile not found: %s", name)
	}
	if err != nil {
		return Profile{}, err
	}
	if err := values.Require("ANDROID_WORKSPACE", "ANDROID_MANIFEST_URL", "ANDROID_BRANCH", "ANDROID_BUILD_TARGET"); err != nil {
		return Profile{}, fmt.Errorf("profile %s: %w", name, err)
	}
	path := filepath.Clean(values["ANDROID_WORKSPACE"])
	if err := safety.RequireAbsolute("ANDROID_WORKSPACE", path); err != nil {
		return Profile{}, err
	}
	return Profile{Name: name, Path: path, Manifest: values["ANDROID_MANIFEST_URL"], Branch: values["ANDROID_BRANCH"], Target: values["ANDROID_BUILD_TARGET"], JavaHome: values["JAVA_HOME"], Ccache: values["CCACHE_DIR"], File: file}, nil
}

func All(root string) ([]Profile, error) {
	entries, err := os.ReadDir(profilesDir(root))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var result []Profile
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".conf" {
			continue
		}
		p, err := Load(root, strings.TrimSuffix(entry.Name(), ".conf"))
		if err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

func Active(root string) (string, error) {
	data, err := os.ReadFile(activeFile(root))
	if err != nil {
		return "", err
	}
	name := strings.TrimSpace(string(data))
	if !validName.MatchString(name) {
		return "", fmt.Errorf("invalid active workspace name")
	}
	return name, nil
}

func Select(root, name string) error {
	if _, err := Load(root, name); err != nil {
		return err
	}
	dir := filepath.Dir(activeFile(root))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	return os.WriteFile(activeFile(root), []byte(name+"\n"), 0o600)
}

func Add(root, name, path, manifest, branch, target, javaHome string) (bool, error) {
	if !validName.MatchString(name) {
		return false, fmt.Errorf("workspace name must match [A-Za-z0-9][A-Za-z0-9_-]*")
	}
	path = filepath.Clean(path)
	if err := safety.RefuseDangerous("workspace path", path, root); err != nil {
		return false, err
	}
	all, err := All(root)
	if err != nil {
		return false, err
	}
	for _, p := range all {
		if p.Name == name {
			return false, fmt.Errorf("workspace profile already exists: %s", name)
		}
		if samePath(p.Path, path) {
			return false, fmt.Errorf("workspace path is already used by profile %s", p.Name)
		}
	}
	if manifest == "" {
		manifest = "https://android.googlesource.com/platform/manifest"
	}
	if branch == "" {
		branch = "android-latest-release"
	}
	if target == "" {
		target = "aosp_x86_64-eng"
	}
	if err := os.MkdirAll(profilesDir(root), 0o755); err != nil {
		return false, err
	}
	content := fmt.Sprintf("NAME=%q\nANDROID_WORKSPACE=%q\nANDROID_MANIFEST_URL=%q\nANDROID_BRANCH=%q\nANDROID_BUILD_TARGET=%q\n", name, path, manifest, branch, target)
	if javaHome != "" {
		content += fmt.Sprintf("JAVA_HOME=%q\n", javaHome)
	}
	content += fmt.Sprintf("CCACHE_DIR=%q\n", filepath.Join(filepath.Dir(path), ".ccache", name))
	if err := os.WriteFile(filepath.Join(profilesDir(root), name+".conf"), []byte(content), 0o600); err != nil {
		return false, err
	}
	_, activeErr := Active(root)
	if activeErr != nil {
		return true, Select(root, name)
	}
	return false, nil
}

func Current(root, requested string) (Profile, error) {
	if requested != "" {
		return Load(root, requested)
	}
	name, err := Active(root)
	if err != nil {
		return Profile{}, fmt.Errorf("no active workspace; run 'lab workspace add <name> --path <absolute-path>'")
	}
	return Load(root, name)
}

func samePath(a, b string) bool {
	left, e1 := filepath.Abs(a)
	right, e2 := filepath.Abs(b)
	return e1 == nil && e2 == nil && filepath.Clean(left) == filepath.Clean(right)
}
