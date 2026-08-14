package build

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Patzilla0o7/android-security-research-lab/internal/safety"
	"github.com/Patzilla0o7/android-security-research-lab/internal/workspaces"
)

var now = time.Now

type metadata struct {
	Workspace string   `json:"workspace"`
	Source    string   `json:"source"`
	Target    string   `json:"target"`
	Modules   []string `json:"modules,omitempty"`
	Jobs      int      `json:"jobs"`
	Mode      string   `json:"build_mode"`
	Status    string   `json:"status"`
	StartedAt string   `json:"started_at"`
	EndedAt   string   `json:"ended_at"`
	Duration  string   `json:"duration"`
	Log       string   `json:"log"`
	ExitCode  int      `json:"exit_code"`
	CCache    string   `json:"ccache,omitempty"`
	OutDir    string   `json:"out_dir"`
}

func Run(root string, args []string, stdout, stderr io.Writer) int {
	command := "plan"
	if len(args) > 0 {
		switch args[0] {
		case "plan", "status", "help", "--help", "-h":
			command, args = args[0], args[1:]
		}
	}
	switch command {
	case "help", "--help", "-h":
		Usage(stdout)
		return 0
	case "status":
		return status(root, args, stdout, stderr)
	case "plan":
		return plan(root, args, stdout, stderr)
	default:
		return usageError(stderr, "unknown build subcommand: "+command)
	}
}

func Usage(w io.Writer) {
	fmt.Fprint(w, `Usage: lab build [plan] [options]
       lab build status [--workspace <name>]

Options:
    --workspace <name>  Use a Workspace profile instead of the active profile.
    --target <target>   Override the Workspace lunch target for this build.
    --jobs <count>      Build concurrency; defaults to logical CPU count.
    --module <name>     Build a module; may be repeated. Omit for a full target.
    --apply             Execute the build. Without it, only show the plan.

Logs and metadata are stored under output/build/<workspace>/.
`)
}

func plan(root string, args []string, stdout, stderr io.Writer) int {
	opts, err := parseOptions(args, true)
	if err != nil {
		return usageError(stderr, err.Error())
	}
	profile, target, mode, freeGiB, err := inspect(root, opts)
	if err != nil {
		return fail(stderr, err)
	}
	fmt.Fprintln(stdout, "\n============================================================\nAOSP Build Plan\n============================================================")
	fmt.Fprintf(stdout, "Workspace     : %s\nSource        : %s\nBranch        : %s\nTarget        : %s\nModules       : %s\nJobs          : %d\nJava home     : %s\nCCache        : %s\nBuild mode    : %s\nDisk free     : %d GiB\nProgress      : interactive terminal\nMode          : %s\n", profile.Name, profile.Path, profile.Branch, target, moduleDisplay(opts.modules), opts.jobs, valueOr(profile.JavaHome, "AOSP/default"), valueOr(profile.Ccache, "disabled"), mode, freeGiB, map[bool]string{false: "plan", true: "apply"}[opts.apply])
	if !opts.apply {
		fmt.Fprintln(stdout, "No changes were made. Add --apply to start the build.")
		return 0
	}
	return execute(root, profile, target, mode, opts, stdout, stderr)
}

func inspect(root string, opts options) (workspaces.Profile, string, string, uint64, error) {
	profile, err := workspaces.Current(root, opts.workspace)
	if err != nil {
		return profile, "", "", 0, err
	}
	if err := safety.RefuseDangerous("workspace path", profile.Path, root); err != nil {
		return profile, "", "", 0, err
	}
	info, err := os.Stat(profile.Path)
	if err != nil || !info.IsDir() {
		return profile, "", "", 0, fmt.Errorf("workspace directory is not ready: %s", profile.Path)
	}
	if info, err := os.Stat(filepath.Join(profile.Path, ".repo")); err != nil || !info.IsDir() {
		return profile, "", "", 0, fmt.Errorf("Repo checkout is not initialized for workspace %s", profile.Name)
	}
	if _, err := os.Stat(filepath.Join(profile.Path, "build", "envsetup.sh")); err != nil {
		return profile, "", "", 0, fmt.Errorf("AOSP build environment is missing: build/envsetup.sh; complete repo sync first")
	}
	target := opts.target
	if target == "" {
		target = profile.Target
	}
	if !validBuildValue(target) {
		return profile, "", "", 0, fmt.Errorf("invalid Workspace build target")
	}
	mode := "full"
	if len(opts.modules) > 0 {
		mode = "module"
	} else if isDirectory(filepath.Join(profile.Path, "out")) {
		mode = "incremental"
	}
	var stat syscall.Statfs_t
	var free uint64
	if err := syscall.Statfs(profile.Path, &stat); err == nil {
		free = stat.Bavail * uint64(stat.Bsize) / (1024 * 1024 * 1024)
	}
	return profile, target, mode, free, nil
}

func execute(root string, profile workspaces.Profile, target, mode string, opts options, stdout, stderr io.Writer) int {
	adapter := filepath.Join(root, "scripts", "aosp-build.sh")
	if info, err := os.Stat(adapter); err != nil || info.IsDir() {
		return fail(stderr, fmt.Errorf("AOSP build adapter is missing: %s", adapter))
	}
	logDir := filepath.Join(root, "output", "build", profile.Name)
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return fail(stderr, err)
	}
	started := now()
	stamp := started.Format("20060102-150405")
	logPath := filepath.Join(logDir, stamp+"-build.log")
	logFile, err := os.Create(logPath)
	if err != nil {
		return fail(stderr, err)
	}
	defer logFile.Close()
	arguments := []string{adapter, profile.Path, target, fmt.Sprint(opts.jobs), profile.Ccache}
	arguments = append(arguments, opts.modules...)
	command := exec.Command("bash", arguments...)
	command.Dir, command.Stdin = profile.Path, os.Stdin
	command.Stdout, command.Stderr = io.MultiWriter(stdout, logFile), io.MultiWriter(stderr, logFile)
	command.Env = os.Environ()
	if profile.JavaHome != "" {
		command.Env = append(command.Env, "JAVA_HOME="+profile.JavaHome, "PATH="+filepath.Join(profile.JavaHome, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"))
	}
	fmt.Fprintf(stdout, "[INFO] Running AOSP build\n[INFO] Log: %s\n", logPath)
	runErr := command.Run()
	ended := now()
	exitCode, statusValue := 0, "success"
	if runErr != nil {
		exitCode, statusValue = 1, "failed"
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}
	record := metadata{Workspace: profile.Name, Source: profile.Path, Target: target, Modules: opts.modules, Jobs: opts.jobs, Mode: mode, Status: statusValue, StartedAt: started.Format(time.RFC3339), EndedAt: ended.Format(time.RFC3339), Duration: ended.Sub(started).Round(time.Second).String(), Log: logPath, ExitCode: exitCode, CCache: profile.Ccache, OutDir: filepath.Join(profile.Path, "out")}
	if err := saveMetadata(logDir, stamp, record); err != nil {
		fmt.Fprintf(stderr, "[WARN] Unable to save build metadata: %v\n", err)
	}
	if runErr != nil {
		fmt.Fprintf(stderr, "[FAIL] AOSP build failed with exit code %d; log: %s\n", exitCode, logPath)
		return exitCode
	}
	fmt.Fprintf(stdout, "[ OK ] AOSP build completed in %s; log: %s\n", record.Duration, logPath)
	return 0
}

func saveMetadata(dir, stamp string, record metadata) error {
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.WriteFile(filepath.Join(dir, stamp+"-build.json"), data, 0o644); err != nil {
		return err
	}
	temp := filepath.Join(dir, ".latest.json.tmp")
	if err := os.WriteFile(temp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(temp, filepath.Join(dir, "latest.json"))
}

func status(root string, args []string, stdout, stderr io.Writer) int {
	opts, err := parseStatusOptions(args)
	if err != nil {
		return usageError(stderr, err.Error())
	}
	profile, err := workspaces.Current(root, opts)
	if err != nil {
		return fail(stderr, err)
	}
	path := filepath.Join(root, "output", "build", profile.Name, "latest.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		fmt.Fprintf(stdout, "No build history for workspace %s.\n", profile.Name)
		return 0
	}
	if err != nil {
		return fail(stderr, err)
	}
	var record metadata
	if err := json.Unmarshal(data, &record); err != nil {
		return fail(stderr, fmt.Errorf("invalid build metadata: %w", err))
	}
	fmt.Fprintln(stdout, "\n============================================================\nAOSP Build Status\n============================================================")
	fmt.Fprintf(stdout, "Workspace     : %s\nStatus        : %s\nTarget        : %s\nModules       : %s\nJobs          : %d\nBuild mode    : %s\nStarted       : %s\nEnded         : %s\nDuration      : %s\nExit code     : %d\nCCache        : %s\nOutput        : %s\nLog           : %s\n", record.Workspace, record.Status, record.Target, moduleDisplay(record.Modules), record.Jobs, record.Mode, record.StartedAt, record.EndedAt, record.Duration, record.ExitCode, valueOr(record.CCache, "disabled"), record.OutDir, record.Log)
	return 0
}

func parseStatusOptions(args []string) (string, error) {
	if len(args) == 0 {
		return "", nil
	}
	if len(args) == 2 && args[0] == "--workspace" && args[1] != "" {
		return args[1], nil
	}
	return "", fmt.Errorf("status only supports --workspace <name>")
}
func moduleDisplay(modules []string) string {
	if len(modules) == 0 {
		return "all (full target)"
	}
	return strings.Join(modules, ", ")
}
func valueOr(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
func isDirectory(path string) bool { info, err := os.Stat(path); return err == nil && info.IsDir() }
func usageError(w io.Writer, message string) int {
	fmt.Fprintf(w, "[FAIL] %s\n", message)
	Usage(w)
	return 2
}
func fail(w io.Writer, err error) int { fmt.Fprintf(w, "[FAIL] %v\n", err); return 1 }
