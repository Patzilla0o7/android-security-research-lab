package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Patzilla0o7/android-security-research-lab/internal/bootstrap"
	buildcommand "github.com/Patzilla0o7/android-security-research-lab/internal/build"
	"github.com/Patzilla0o7/android-security-research-lab/internal/device"
	"github.com/Patzilla0o7/android-security-research-lab/internal/doctor"
	repocommand "github.com/Patzilla0o7/android-security-research-lab/internal/repo"
	"github.com/Patzilla0o7/android-security-research-lab/internal/workspaces"
)

const (
	exitSuccess        = 0
	exitFailure        = 1
	exitUsage          = 2
	exitNotImplemented = 3
)

var placeholderCommands = map[string]string{
	"research": "Research",
}

// Run executes the ASRL command line and returns its process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	command := "help"
	var commandArgs []string
	if len(args) > 0 {
		command = args[0]
		commandArgs = args[1:]
	}

	switch command {
	case "-h", "--help":
		command = "help"
	case "-V", "--version":
		command = "version"
	}

	switch command {
	case "help":
		printHelp(stdout)
		return exitSuccess
	case "version":
		return printVersion(stdout, stderr)
	case "doctor":
		if len(commandArgs) > 1 || (len(commandArgs) == 1 && commandArgs[0] != "--json") {
			fmt.Fprintln(stderr, "[FAIL] Usage: lab doctor [--json]")
			return exitUsage
		}
		root, err := projectRoot()
		if err != nil {
			fmt.Fprintf(stderr, "[FAIL] %v\n", err)
			return exitFailure
		}
		if err := doctor.Run(root, commandArgs, stdout); err != nil {
			fmt.Fprintf(stderr, "[FAIL] Doctor failed: %v\n", err)
			return exitFailure
		}
		return exitSuccess
	case "bootstrap":
		root, err := projectRoot()
		if err != nil {
			fmt.Fprintf(stderr, "[FAIL] %v\n", err)
			return exitFailure
		}
		return bootstrap.Run(root, commandArgs, stdout, stderr)
	case "workspace":
		root, err := projectRoot()
		if err != nil {
			fmt.Fprintf(stderr, "[FAIL] %v\n", err)
			return exitFailure
		}
		return workspaces.Run(root, commandArgs, stdout, stderr)
	case "repo":
		root, err := projectRoot()
		if err != nil {
			fmt.Fprintf(stderr, "[FAIL] %v\n", err)
			return exitFailure
		}
		return repocommand.Run(root, commandArgs, stdout, stderr)
	case "build":
		root, err := projectRoot()
		if err != nil {
			fmt.Fprintf(stderr, "[FAIL] %v\n", err)
			return exitFailure
		}
		return buildcommand.Run(root, commandArgs, stdout, stderr)
	case "device":
		return device.Run(commandArgs, stdout, stderr)
	default:
		if label, ok := placeholderCommands[command]; ok {
			fmt.Fprintf(stdout, "[INFO] %s module is not implemented.\n", label)
			return exitNotImplemented
		}
		fmt.Fprintf(stderr, "[FAIL] Unknown command: %s\n\n", command)
		printHelp(stderr)
		return exitUsage
	}
}

func printHelp(w io.Writer) {
	fmt.Fprint(w, `
Android Security Research Lab

Usage

    lab <command> [options]
    lab --help
    lab --version

Commands

    help
    version
    doctor
    bootstrap
        plan (default) | --apply
    workspace
    repo
    build
    device
    research

Run 'lab <command> --help' for command-specific usage where available.

`)
}

func printVersion(stdout, stderr io.Writer) int {
	root, err := projectRoot()
	if err != nil {
		fmt.Fprintf(stderr, "[FAIL] %v\n", err)
		return exitFailure
	}
	data, err := os.ReadFile(filepath.Join(root, "VERSION"))
	if err != nil {
		fmt.Fprintf(stderr, "[FAIL] Unable to read VERSION: %v\n", err)
		return exitFailure
	}
	fmt.Fprintf(stdout, "Android Security Research Lab\n\nVersion : %s\n", strings.TrimSpace(string(data)))
	return exitSuccess
}

func projectRoot() (string, error) {
	if root := os.Getenv("ASRL_ROOT"); root != "" {
		return filepath.Abs(root)
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("unable to determine project root: %w", err)
	}
	for dir := wd; ; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "VERSION")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "config")); err == nil {
				return dir, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	return "", fmt.Errorf("unable to locate ASRL project root")
}
