package workspaces

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Patzilla0o7/android-security-research-lab/internal/safety"
)

func Run(root string, args []string, out, stderr io.Writer) int {
	command := "status"
	if len(args) > 0 {
		command, args = args[0], args[1:]
	}
	switch command {
	case "help", "--help", "-h":
		Usage(out)
		return 0
	case "list":
		if len(args) != 0 {
			return usage(stderr, "list accepts no arguments")
		}
		all, err := All(root)
		if err != nil {
			return fail(stderr, err)
		}
		active, _ := Active(root)
		if len(all) == 0 {
			fmt.Fprintln(out, "No workspace profiles configured.")
		}
		for _, p := range all {
			mark := " "
			if p.Name == active {
				mark = "*"
			}
			fmt.Fprintf(out, "%s %-20s %s\n", mark, p.Name, p.Path)
		}
		return 0
	case "add":
		return addCommand(root, args, out, stderr)
	case "use":
		if len(args) != 1 {
			return usage(stderr, "use requires one profile name")
		}
		if err := Select(root, args[0]); err != nil {
			return fail(stderr, err)
		}
		fmt.Fprintf(out, "[ OK ] Active workspace: %s\n", args[0])
		return 0
	case "current":
		if len(args) != 0 {
			return usage(stderr, "current accepts no arguments")
		}
		p, err := Current(root, "")
		if err != nil {
			return fail(stderr, err)
		}
		fmt.Fprintln(out, p.Name)
		return 0
	case "status":
		return statusCommand(root, args, out, stderr)
	case "init":
		return initCommand(root, args, out, stderr)
	default:
		return usage(stderr, "unknown workspace subcommand: "+command)
	}
}

func Usage(w io.Writer) {
	fmt.Fprint(w, `Usage: lab workspace <command> [options]

Commands:
    list
    add <name> --path <absolute-path> [--branch <branch>] [--target <target>]
    use <name>
    current
    status [name]
    init [name]
`)
}

func addCommand(root string, args []string, out, stderr io.Writer) int {
	if len(args) < 3 {
		return usage(stderr, "add requires <name> and --path")
	}
	name := args[0]
	opts := map[string]string{}
	for i := 1; i < len(args); i += 2 {
		if i+1 >= len(args) {
			return usage(stderr, "missing value for "+args[i])
		}
		switch args[i] {
		case "--path", "--manifest", "--branch", "--target", "--java-home":
			opts[args[i]] = args[i+1]
		default:
			return usage(stderr, "unknown add option: "+args[i])
		}
	}
	if opts["--path"] == "" {
		return usage(stderr, "add requires --path")
	}
	activated, err := Add(root, name, opts["--path"], opts["--manifest"], opts["--branch"], opts["--target"], opts["--java-home"])
	if err != nil {
		return fail(stderr, err)
	}
	fmt.Fprintf(out, "[ OK ] Workspace profile added: %s\n", name)
	if activated {
		fmt.Fprintf(out, "[ OK ] Active workspace: %s\n", name)
	}
	return 0
}

func statusCommand(root string, args []string, out, stderr io.Writer) int {
	if len(args) > 1 {
		return usage(stderr, "status accepts at most one name")
	}
	name := ""
	if len(args) == 1 {
		name = args[0]
	}
	p, err := Current(root, name)
	if err != nil {
		return fail(stderr, err)
	}
	fmt.Fprintln(out, "\n============================================================\nWorkspace Status\n============================================================")
	fmt.Fprintf(out, "Profile       : %s\nConfiguration : %s\nWorkspace     : %s\nBranch        : %s\nBuild target  : %s\n", p.Name, p.File, p.Path, p.Branch, p.Target)
	info, err := os.Stat(p.Path)
	if os.IsNotExist(err) {
		fmt.Fprintln(out, "Directory     : missing\nRepo checkout : not initialized")
		return 0
	}
	if err != nil {
		return fail(stderr, err)
	}
	if !info.IsDir() {
		return fail(stderr, fmt.Errorf("workspace path exists but is not a directory"))
	}
	fmt.Fprintln(out, "Directory     : ready")
	if info, err := os.Stat(filepath.Join(p.Path, ".repo")); err == nil && info.IsDir() {
		fmt.Fprintln(out, "Repo checkout : initialized")
	} else {
		fmt.Fprintln(out, "Repo checkout : not initialized")
	}
	return 0
}

func initCommand(root string, args []string, out, stderr io.Writer) int {
	if len(args) > 1 {
		return usage(stderr, "init accepts at most one name")
	}
	name := ""
	if len(args) == 1 {
		name = args[0]
	}
	p, err := Current(root, name)
	if err != nil {
		return fail(stderr, err)
	}
	if err := safety.RefuseDangerous("workspace path", p.Path, root); err != nil {
		return fail(stderr, err)
	}
	if info, err := os.Stat(p.Path); err == nil && !info.IsDir() {
		return fail(stderr, fmt.Errorf("workspace path exists but is not a directory"))
	} else if err != nil && !os.IsNotExist(err) {
		return fail(stderr, err)
	}
	if err := os.MkdirAll(p.Path, 0o755); err != nil {
		return fail(stderr, err)
	}
	fmt.Fprintf(out, "[ OK ] Workspace directory is ready: %s (%s)\n", p.Path, p.Name)
	return 0
}
func usage(w io.Writer, message string) int {
	fmt.Fprintf(w, "[FAIL] %s\n", message)
	Usage(w)
	return 2
}
func fail(w io.Writer, err error) int { fmt.Fprintf(w, "[FAIL] %v\n", err); return 1 }
