package repo

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Patzilla0o7/android-security-research-lab/internal/safety"
	"github.com/Patzilla0o7/android-security-research-lab/internal/workspaces"
)

var (
	lookPath = exec.LookPath
	now      = time.Now
)

func Run(root string, args []string, stdout, stderr io.Writer) int {
	command := "status"
	if len(args) > 0 {
		command, args = args[0], args[1:]
	}
	switch command {
	case "help", "--help", "-h":
		Usage(stdout)
		return 0
	case "status":
		return status(root, args, stdout, stderr)
	case "init":
		return initialize(root, args, stdout, stderr)
	case "sync":
		return syncWorkspace(root, args, stdout, stderr)
	default:
		return usageError(stderr, "unknown repo subcommand: "+command)
	}
}

func Usage(w io.Writer) {
	fmt.Fprint(w, `Usage: lab repo <command> [options]

Commands:
    status [--workspace <name>]
    init [--workspace <name>]
    sync [--workspace <name>] [--jobs <count>] [--project <path>]...
         [--retry-fetches <count>] [--no-clone-bundle] [--force-sync] [--apply]

sync is plan-only unless --apply is provided. init and applied sync write logs
under output/repo/<workspace>/.
`)
}

func status(root string, args []string, stdout, stderr io.Writer) int {
	opts, err := parseOptions(args, false, false)
	if err != nil {
		return usageError(stderr, err.Error())
	}
	profile, err := checkedProfile(root, opts.workspace)
	if err != nil {
		return fail(stderr, err)
	}
	repoPath, repoErr := lookPath("repo")
	fmt.Fprintln(stdout, "\n============================================================\nRepo Status\n============================================================")
	fmt.Fprintf(stdout, "Workspace     : %s\nPath          : %s\nManifest URL  : %s\nBranch        : %s\n", profile.Name, profile.Path, profile.Manifest, profile.Branch)
	if repoErr == nil {
		fmt.Fprintf(stdout, "Repo command  : %s\n", repoPath)
	} else {
		fmt.Fprintln(stdout, "Repo command  : not installed")
	}
	if isDirectory(filepath.Join(profile.Path, ".repo")) {
		fmt.Fprintln(stdout, "Repo checkout : initialized")
		manifest := filepath.Join(profile.Path, ".repo", "manifest.xml")
		if _, err := os.Lstat(manifest); err == nil {
			fmt.Fprintf(stdout, "Manifest file : %s\n", manifest)
		} else {
			fmt.Fprintln(stdout, "Manifest file : missing")
		}
	} else {
		fmt.Fprintln(stdout, "Repo checkout : not initialized")
	}
	return 0
}

func initialize(root string, args []string, stdout, stderr io.Writer) int {
	opts, err := parseOptions(args, false, false)
	if err != nil {
		return usageError(stderr, err.Error())
	}
	profile, err := checkedProfile(root, opts.workspace)
	if err != nil {
		return fail(stderr, err)
	}
	if _, err := lookPath("repo"); err != nil {
		return fail(stderr, fmt.Errorf("repo command is not installed; run 'lab bootstrap plan'"))
	}
	if err := os.MkdirAll(profile.Path, 0o755); err != nil {
		return fail(stderr, fmt.Errorf("create workspace: %w", err))
	}
	arguments := []string{"init", "-u", profile.Manifest, "-b", profile.Branch}
	return execute(root, profile, "init", arguments, stdout, stderr)
}

func syncWorkspace(root string, args []string, stdout, stderr io.Writer) int {
	opts, err := parseOptions(args, true, true)
	if err != nil {
		return usageError(stderr, err.Error())
	}
	profile, err := checkedProfile(root, opts.workspace)
	if err != nil {
		return fail(stderr, err)
	}
	jobs := opts.jobs
	if jobs == 0 {
		jobs = runtime.NumCPU()
	}
	fmt.Fprintln(stdout, "\n============================================================\nRepo Sync Plan\n============================================================")
	projects := "all manifest projects"
	if len(opts.projects) > 0 {
		projects = strings.Join(opts.projects, ", ")
	}
	fmt.Fprintf(stdout, "Workspace     : %s\nPath          : %s\nProjects      : %s\nJobs          : %d\nRetry fetches : %d\nClone bundle  : %t\nForce sync    : %t\nProgress      : interactive terminal\nMode          : %s\n", profile.Name, profile.Path, projects, jobs, opts.retryFetches, !opts.noCloneBundle, opts.forceSync, map[bool]string{false: "plan", true: "apply"}[opts.apply])
	if !isDirectory(filepath.Join(profile.Path, ".repo")) {
		return fail(stderr, fmt.Errorf("workspace is not initialized; run 'lab repo init --workspace %s'", profile.Name))
	}
	if !opts.apply {
		fmt.Fprintln(stdout, "No changes were made. Add --apply to synchronize sources.")
		return 0
	}
	if _, err := lookPath("repo"); err != nil {
		return fail(stderr, fmt.Errorf("repo command is not installed; run 'lab bootstrap plan'"))
	}
	arguments := []string{"sync", "-c", "-j", fmt.Sprint(jobs)}
	if opts.retryFetches > 0 {
		arguments = append(arguments, fmt.Sprintf("--retry-fetches=%d", opts.retryFetches))
	}
	if opts.noCloneBundle {
		arguments = append(arguments, "--no-clone-bundle")
	}
	if opts.forceSync {
		arguments = append(arguments, "--force-sync")
	}
	arguments = append(arguments, opts.projects...)
	return execute(root, profile, "sync", arguments, stdout, stderr)
}

func checkedProfile(root, requested string) (workspaces.Profile, error) {
	profile, err := workspaces.Current(root, requested)
	if err != nil {
		return profile, err
	}
	if err := safety.RefuseDangerous("workspace path", profile.Path, root); err != nil {
		return profile, err
	}
	if info, err := os.Stat(profile.Path); err == nil && !info.IsDir() {
		return profile, fmt.Errorf("workspace path exists but is not a directory")
	} else if err != nil && !os.IsNotExist(err) {
		return profile, err
	}
	return profile, nil
}

func execute(root string, profile workspaces.Profile, operation string, args []string, stdout, stderr io.Writer) int {
	logDir := filepath.Join(root, "output", "repo", profile.Name)
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return fail(stderr, fmt.Errorf("create log directory: %w", err))
	}
	logPath := filepath.Join(logDir, now().Format("20060102-150405")+"-"+operation+".log")
	logFile, err := os.Create(logPath)
	if err != nil {
		return fail(stderr, fmt.Errorf("create log: %w", err))
	}
	defer logFile.Close()
	command := exec.Command("repo", args...)
	command.Dir = profile.Path
	command.Stdout = io.MultiWriter(stdout, logFile)
	if isTerminal(stderr) {
		// Repo and Git only render interactive transfer progress when stderr is
		// a terminal. Preserve that file descriptor instead of hiding it behind
		// the pipe os/exec creates for an io.MultiWriter.
		command.Stderr = stderr
		fmt.Fprintln(stdout, "[INFO] Interactive progress is shown on the terminal and may not be repeated in the log.")
	} else {
		command.Stderr = io.MultiWriter(stderr, logFile)
	}
	command.Stdin = os.Stdin
	fmt.Fprintf(stdout, "[INFO] Running: repo %s\n[INFO] Log: %s\n", strings.Join(args, " "), logPath)
	if err := command.Run(); err != nil {
		fmt.Fprintf(stderr, "[FAIL] Repo %s failed; log: %s\n", operation, logPath)
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		return 1
	}
	fmt.Fprintf(stdout, "[ OK ] Repo %s completed; log: %s\n", operation, logPath)
	return 0
}

func isDirectory(path string) bool { info, err := os.Stat(path); return err == nil && info.IsDir() }
func isTerminal(w io.Writer) bool {
	file, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
func usageError(w io.Writer, message string) int {
	fmt.Fprintf(w, "[FAIL] %s\n", message)
	Usage(w)
	return 2
}
func fail(w io.Writer, err error) int { fmt.Fprintf(w, "[FAIL] %v\n", err); return 1 }
