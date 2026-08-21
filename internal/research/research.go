package research

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Patzilla0o7/android-security-research-lab/internal/workspaces"
)

var now = time.Now

func Run(root string, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		Usage(stdout)
		return 0
	}
	command, args := args[0], args[1:]
	switch command {
	case "new":
		return newCommand(root, args, stdout, stderr)
	case "list":
		return listCommand(root, args, stdout, stderr)
	case "show":
		return showCommand(root, args, stdout, stderr)
	case "validate":
		return validateCommand(root, args, stdout, stderr)
	default:
		return usageError(stderr, "unknown research subcommand: "+command)
	}
}

func Usage(w io.Writer) {
	fmt.Fprint(w, `Usage: lab research <command> [options]

Commands:
    new <case-id> --title <title> [--workspace <name>]
    list
    show <case-id>
    validate <case-id>
`)
}

func newCommand(root string, args []string, stdout, stderr io.Writer) int {
	if len(args) < 3 {
		return usageError(stderr, "new requires <case-id> and --title")
	}
	id, args := args[0], args[1:]
	dir, err := caseDir(root, id)
	if err != nil {
		return usageError(stderr, err.Error())
	}
	title, workspace := "", ""
	for i := 0; i < len(args); i++ {
		if i+1 >= len(args) {
			return usageError(stderr, args[i]+" requires a value")
		}
		switch args[i] {
		case "--title":
			title = args[i+1]
		case "--workspace":
			workspace = args[i+1]
		default:
			return usageError(stderr, "unknown new option: "+args[i])
		}
		i++
	}
	if strings.TrimSpace(title) == "" {
		return usageError(stderr, "--title is required")
	}
	profile, err := workspaces.Current(root, workspace)
	if err != nil {
		return fail(stderr, err)
	}
	if err := os.MkdirAll(filepath.Dir(dir), 0o755); err != nil {
		return fail(stderr, err)
	}
	if err := os.Mkdir(dir, 0o755); os.IsExist(err) {
		return fail(stderr, fmt.Errorf("research case already exists: %s", id))
	} else if err != nil {
		return fail(stderr, err)
	}
	for _, name := range []string{"patches", "poc", "reports"} {
		if err := os.Mkdir(filepath.Join(dir, name), 0o755); err != nil {
			return fail(stderr, err)
		}
	}
	timestamp := now().UTC().Format(time.RFC3339)
	c := Case{SchemaVersion: 1, CaseID: id, Title: title, Status: "investigating", Workspace: profile.Name, DisclosureStatus: "private", CreatedAt: timestamp, UpdatedAt: timestamp}
	if err := writeCase(dir, c); err != nil {
		return fail(stderr, err)
	}
	templates := map[string]string{
		"README.md":       fmt.Sprintf("# %s\n\nCase ID: `%s`\n\n## Summary\n\nDescribe the issue and current research status.\n", title, id),
		"timeline.md":     "# Timeline\n\nRecord observations and actions in UTC.\n",
		"reproduction.md": "# Reproduction\n\n## Preconditions\n\n## Steps\n\n## Expected result\n\n## Actual result\n",
		"root-cause.md":   "# Root Cause\n\n## Affected component\n\n## Analysis\n\n## Security impact\n",
	}
	for name, content := range templates {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			return fail(stderr, err)
		}
	}
	fmt.Fprintf(stdout, "[ OK ] Research case created: %s\nDirectory: %s\n", id, dir)
	return 0
}

func listCommand(root string, args []string, stdout, stderr io.Writer) int {
	if len(args) != 0 {
		return usageError(stderr, "list accepts no arguments")
	}
	cases, err := allCases(root)
	if err != nil {
		return fail(stderr, err)
	}
	if len(cases) == 0 {
		fmt.Fprintln(stdout, "No research cases found.")
		return 0
	}
	fmt.Fprintln(stdout, "CASE\tSTATUS\tWORKSPACE\tTITLE")
	for _, c := range cases {
		fmt.Fprintf(stdout, "%s\t%s\t%s\t%s\n", c.CaseID, c.Status, c.Workspace, c.Title)
	}
	return 0
}

func showCommand(root string, args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		return usageError(stderr, "show requires one case ID")
	}
	c, err := loadCase(root, args[0])
	if err != nil {
		return fail(stderr, err)
	}
	fmt.Fprintf(stdout, "Case ID        : %s\nTitle          : %s\nStatus         : %s\nWorkspace      : %s\nComponent      : %s\nDisclosure     : %s\nCreated        : %s\nUpdated        : %s\n", c.CaseID, c.Title, c.Status, c.Workspace, valueOr(c.AffectedComponent, "unspecified"), valueOr(c.DisclosureStatus, "unspecified"), c.CreatedAt, c.UpdatedAt)
	return 0
}

func validateCommand(root string, args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		return usageError(stderr, "validate requires one case ID")
	}
	id := args[0]
	dir, err := caseDir(root, id)
	if err != nil {
		return fail(stderr, err)
	}
	if _, err := loadCase(root, id); err != nil {
		return fail(stderr, err)
	}
	for _, name := range []string{"README.md", "timeline.md", "reproduction.md", "root-cause.md"} {
		info, err := os.Stat(filepath.Join(dir, name))
		if err != nil || !info.Mode().IsRegular() {
			return fail(stderr, fmt.Errorf("required research asset missing: %s", name))
		}
	}
	for _, name := range []string{"patches", "poc", "reports"} {
		info, err := os.Stat(filepath.Join(dir, name))
		if err != nil || !info.IsDir() {
			return fail(stderr, fmt.Errorf("required research directory missing: %s", name))
		}
	}
	fmt.Fprintf(stdout, "[ OK ] Research case is valid: %s\n", id)
	return 0
}

func valueOr(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
func usageError(w io.Writer, message string) int {
	fmt.Fprintf(w, "[FAIL] %s\n", message)
	Usage(w)
	return 2
}
func fail(w io.Writer, err error) int { fmt.Fprintf(w, "[FAIL] %v\n", err); return 1 }
