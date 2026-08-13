package workspace

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Patzilla0o7/android-security-research-lab/internal/config"
	"github.com/Patzilla0o7/android-security-research-lab/internal/safety"
)

func Run(root string, args []string, stdout, stderr io.Writer) int {
	command := "status"
	if len(args) > 0 {
		command = args[0]
	}
	if len(args) > 1 {
		fmt.Fprintln(stderr, "[FAIL] workspace accepts one subcommand")
		Usage(stderr)
		return 2
	}
	switch command {
	case "help", "--help", "-h":
		Usage(stdout)
		return 0
	case "status":
		return status(root, stdout, stderr)
	case "init":
		return initWorkspace(root, stdout, stderr)
	default:
		fmt.Fprintf(stderr, "[FAIL] Unknown workspace subcommand: %s\n", command)
		Usage(stderr)
		return 2
	}
}

func Usage(w io.Writer) {
	fmt.Fprint(w, `Usage: lab workspace <status|init>

Commands:
    status    Show configuration and AOSP Repo workspace state (default).
    init      Validate the configured path and create the workspace directory.
`)
}

func status(root string, stdout, stderr io.Writer) int {
	path, configPath, err := configuredPath(root)
	if err != nil {
		fmt.Fprintf(stderr, "[FAIL] %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, "\n============================================================\nWorkspace Status\n============================================================")
	fmt.Fprintf(stdout, "Configuration : %s\nWorkspace     : %s\n", configPath, path)
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		fmt.Fprintln(stdout, "Directory     : missing\nRepo checkout : not initialized")
		return 0
	}
	if err != nil {
		fmt.Fprintf(stderr, "[FAIL] Unable to inspect workspace: %v\n", err)
		return 1
	}
	if !info.IsDir() {
		fmt.Fprintln(stderr, "[FAIL] Workspace path exists but is not a directory")
		return 1
	}
	fmt.Fprintln(stdout, "Directory     : ready")
	if info, err := os.Stat(filepath.Join(path, ".repo")); err == nil && info.IsDir() {
		fmt.Fprintln(stdout, "Repo checkout : initialized")
	} else {
		fmt.Fprintln(stdout, "Repo checkout : not initialized")
	}
	return 0
}

func initWorkspace(root string, stdout, stderr io.Writer) int {
	path, _, err := configuredPath(root)
	if err != nil {
		fmt.Fprintf(stderr, "[FAIL] %v\n", err)
		return 1
	}
	if err := safety.RefuseDangerous("workspace path", path, root); err != nil {
		fmt.Fprintf(stderr, "[FAIL] %v\n", err)
		return 1
	}
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		fmt.Fprintln(stderr, "[FAIL] Workspace path exists but is not a directory")
		return 1
	} else if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(stderr, "[FAIL] Unable to inspect workspace: %v\n", err)
		return 1
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		fmt.Fprintf(stderr, "[FAIL] Unable to create workspace: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "[ OK ] Workspace directory is ready: %s\n", path)
	return 0
}

func configuredPath(root string) (string, string, error) {
	configPath := os.Getenv("LAB_CONFIG_FILE")
	if configPath == "" {
		configPath = filepath.Join(root, "config", "lab.conf")
	}
	values, err := config.Load(configPath, map[string]string{"LAB_ROOT": root})
	if err != nil {
		if os.IsNotExist(err) {
			return "", configPath, fmt.Errorf("lab configuration not found: %s (copy config/lab.conf.example)", configPath)
		}
		return "", configPath, fmt.Errorf("load lab configuration: %w", err)
	}
	if err := values.Require("ANDROID_WORKSPACE"); err != nil {
		return "", configPath, err
	}
	path := filepath.Clean(values["ANDROID_WORKSPACE"])
	if err := safety.RequireAbsolute("ANDROID_WORKSPACE", path); err != nil {
		return "", configPath, err
	}
	return path, configPath, nil
}
