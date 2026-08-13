package repo

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Patzilla0o7/android-security-research-lab/internal/workspaces"
)

type patchMetadata struct {
	Workspace string   `json:"workspace"`
	Project   string   `json:"project"`
	Branch    string   `json:"manifest_branch"`
	Base      string   `json:"base_commit"`
	Files     []string `json:"files"`
}

func patchCommand(root string, args []string, out, stderr io.Writer) int {
	if len(args) == 0 {
		return usageError(stderr, "patch requires export or import")
	}
	switch args[0] {
	case "export":
		return patchExport(root, args[1:], out, stderr)
	case "import":
		return patchImport(root, args[1:], out, stderr)
	default:
		return usageError(stderr, "unknown patch subcommand: "+args[0])
	}
}

func patchExport(root string, args []string, out, stderr io.Writer) int {
	opts, err := parseResearchOptions(args, false, true, false)
	if err != nil {
		return usageError(stderr, err.Error())
	}
	if len(opts.projects) == 0 {
		return usageError(stderr, "patch export requires --project")
	}
	p, err := checkedInitializedProfile(root, opts.workspace)
	if err != nil {
		return fail(stderr, err)
	}
	stamp := now().Format("20060102-150405")
	bundle := filepath.Join(root, "output", "repo", p.Name, "patches", stamp)
	if err := os.MkdirAll(bundle, 0755); err != nil {
		return fail(stderr, err)
	}
	for _, project := range opts.projects {
		dir := filepath.Join(p.Path, project)
		if !isGitRepo(dir) {
			return fail(stderr, fmt.Errorf("not a Git project: %s", project))
		}
		dest := filepath.Join(bundle, strings.ReplaceAll(project, "/", "__"))
		if err := os.MkdirAll(dest, 0755); err != nil {
			return fail(stderr, err)
		}
		files := []string{}
		base := gitOutput(dir, "rev-parse", "HEAD")
		if opts.commits > 0 {
			cmd := exec.Command("git", "format-patch", fmt.Sprintf("-%d", opts.commits), "--output-directory", dest)
			cmd.Dir = dir
			if data, err := cmd.CombinedOutput(); err != nil {
				return fail(stderr, fmt.Errorf("format-patch %s: %s", project, data))
			}
			entries, _ := os.ReadDir(dest)
			for _, e := range entries {
				if strings.HasSuffix(e.Name(), ".patch") {
					files = append(files, e.Name())
				}
			}
		}
		diff, err := exec.Command("git", "-C", dir, "diff", "--binary", "HEAD").Output()
		if err != nil {
			return fail(stderr, err)
		}
		if len(diff) > 0 {
			name := "working-tree.diff"
			if err := os.WriteFile(filepath.Join(dest, name), diff, 0644); err != nil {
				return fail(stderr, err)
			}
			files = append(files, name)
		}
		meta := patchMetadata{p.Name, project, p.Branch, base, files}
		data, err := json.MarshalIndent(meta, "", "  ")
		if err != nil {
			return fail(stderr, err)
		}
		if err := os.WriteFile(filepath.Join(dest, "metadata.json"), append(data, '\n'), 0644); err != nil {
			return fail(stderr, err)
		}
		if err := writeChecksums(dest, append(files, "metadata.json")); err != nil {
			return fail(stderr, err)
		}
	}
	fmt.Fprintf(out, "[ OK ] Patch bundle exported: %s\n", bundle)
	return 0
}

func patchImport(root string, args []string, out, stderr io.Writer) int {
	opts, err := parseResearchOptions(args, true, false, true)
	if err != nil {
		return usageError(stderr, err.Error())
	}
	if len(opts.projects) != 1 || opts.file == "" {
		return usageError(stderr, "patch import requires one --project and --file")
	}
	p, err := checkedInitializedProfile(root, opts.workspace)
	if err != nil {
		return fail(stderr, err)
	}
	file, err := filepath.Abs(opts.file)
	if err != nil {
		return fail(stderr, err)
	}
	if info, err := os.Stat(file); err != nil || info.IsDir() {
		return fail(stderr, fmt.Errorf("patch file is not readable: %s", file))
	}
	dir := filepath.Join(p.Path, opts.projects[0])
	if !isGitRepo(dir) {
		return fail(stderr, fmt.Errorf("not a Git project: %s", opts.projects[0]))
	}
	applyArgs := []string{"apply", "--check", file}
	if strings.HasSuffix(file, ".patch") {
		applyArgs = []string{"apply", "--check", file}
	}
	cmd := exec.Command("git", applyArgs...)
	cmd.Dir = dir
	if data, err := cmd.CombinedOutput(); err != nil {
		return fail(stderr, fmt.Errorf("patch check failed: %s", data))
	}
	fmt.Fprintf(out, "Patch check passed: %s\nMode: %s\n", file, map[bool]string{false: "plan", true: "apply"}[opts.apply])
	if !opts.apply {
		fmt.Fprintln(out, "No changes were made. Add --apply to import.")
		return 0
	}
	if strings.HasSuffix(file, ".patch") {
		cmd = exec.Command("git", "am", file)
	} else {
		cmd = exec.Command("git", "apply", file)
	}
	cmd.Dir = dir
	cmd.Stdout = out
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return 1
	}
	fmt.Fprintln(out, "[ OK ] Patch imported")
	return 0
}

func checkedInitializedProfile(root, name string) (workspaces.Profile, error) {
	p, err := checkedProfile(root, name)
	if err != nil {
		return p, err
	}
	if !isDirectory(filepath.Join(p.Path, ".repo")) {
		return p, fmt.Errorf("workspace is not initialized")
	}
	return p, nil
}
func isGitRepo(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil && (info.IsDir() || info.Mode().IsRegular())
}
func gitOutput(dir string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	data, _ := cmd.Output()
	return strings.TrimSpace(string(data))
}
func writeChecksums(dir string, files []string) error {
	var lines []string
	for _, name := range files {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		lines = append(lines, hex.EncodeToString(sum[:])+"  "+name)
	}
	return os.WriteFile(filepath.Join(dir, "SHA256SUMS"), []byte(strings.Join(lines, "\n")+"\n"), 0644)
}
