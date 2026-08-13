package bootstrap

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Patzilla0o7/android-security-research-lab/internal/toolchain"
)

type Mode string

const (
	ModePlan  Mode = "plan"
	ModeApply Mode = "apply"
	ModeHelp  Mode = "help"
)

type Plan struct {
	Satisfied   []string
	AptPackages []string
	Manual      []string
	Unavailable []string
}

func ParseArgs(args []string) (Mode, error) {
	if len(args) == 0 || (len(args) == 1 && args[0] == "plan") {
		return ModePlan, nil
	}
	if len(args) != 1 {
		return "", fmt.Errorf("bootstrap accepts at most one option")
	}
	switch args[0] {
	case "--apply":
		return ModeApply, nil
	case "help", "--help", "-h":
		return ModeHelp, nil
	default:
		return "", fmt.Errorf("unknown bootstrap option: %s", args[0])
	}
}

func Run(root string, args []string, stdout, stderr io.Writer) int {
	mode, err := ParseArgs(args)
	if err != nil {
		fmt.Fprintf(stderr, "[FAIL] %v\n", err)
		Usage(stderr)
		return 2
	}
	if mode == ModeHelp {
		Usage(stdout)
		return 0
	}
	tools, err := toolchain.Load(filepath.Join(root, "config", "tools.conf"))
	if err != nil {
		fmt.Fprintf(stderr, "[FAIL] Unable to load toolchain: %v\n", err)
		return 1
	}
	plan := Collect(toolchain.CheckAll(tools), aptAvailable)
	printPlan(stdout, mode, plan)
	if mode == ModeApply && len(plan.AptPackages) > 0 {
		script := filepath.Join(root, "scripts", "bootstrap-apt.sh")
		cmd := exec.Command(script, plan.AptPackages...)
		cmd.Dir, cmd.Env = root, os.Environ()
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, stdout, stderr
		if err := cmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				return exitErr.ExitCode()
			}
			fmt.Fprintf(stderr, "[FAIL] Unable to run apt adapter: %v\n", err)
			return 1
		}
	}
	printSummary(stdout, plan)
	return 0
}

func Collect(results []toolchain.Result, available func(string) bool) Plan {
	var plan Plan
	seen := make(map[string]bool)
	for _, result := range results {
		if result.Status == "installed" {
			plan.Satisfied = append(plan.Satisfied, result.Tool.Label)
			continue
		}
		if pkg := toolchain.AptPackage(result.Tool); pkg != "" && available(pkg) {
			if !seen[pkg] {
				plan.AptPackages = append(plan.AptPackages, pkg)
				seen[pkg] = true
			}
			continue
		}
		if method := toolchain.ManualMethod(result.Tool); method != "" {
			plan.Manual = append(plan.Manual, result.Tool.Label+": "+method)
		} else {
			plan.Unavailable = append(plan.Unavailable, result.Tool.Label+": no supported installation method")
		}
	}
	return plan
}

func Usage(w io.Writer) {
	fmt.Fprint(w, `Usage: lab bootstrap [plan|--apply]

Modes:
    plan       Check tools and show installation methods without changing the host.
    --apply    Install missing tools that have an available apt package.
`)
}

func aptAvailable(pkg string) bool {
	return exec.Command("apt-cache", "show", pkg).Run() == nil
}

func printPlan(w io.Writer, mode Mode, plan Plan) {
	fmt.Fprintln(w, "\n============================================================\nBootstrap Plan\n============================================================")
	fmt.Fprintf(w, "[INFO] Mode: %s\n", mode)
	line(w, "Already satisfied", plan.Satisfied)
	line(w, "Install with apt", plan.AptPackages)
	line(w, "Manual action", plan.Manual)
	line(w, "Unavailable", plan.Unavailable)
	if mode == ModePlan {
		fmt.Fprintln(w, "[INFO] No changes were made. Run 'lab bootstrap --apply' to install available apt packages.")
	}
}

func printSummary(w io.Writer, plan Plan) {
	fmt.Fprintln(w, "\n============================================================\nBootstrap Summary\n============================================================")
	fmt.Fprintf(w, "%-12s %d\n%-12s %d\n%-12s %d\n%-12s %d\n", "Satisfied", len(plan.Satisfied), "AptPackages", len(plan.AptPackages), "Manual", len(plan.Manual), "Unavailable", len(plan.Unavailable))
}

func line(w io.Writer, label string, values []string) {
	if len(values) > 0 {
		fmt.Fprintf(w, "[INFO] %s: %s\n", label, strings.Join(values, " "))
	}
}
