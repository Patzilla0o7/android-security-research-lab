package repo

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

func branchCommand(root string, args []string, out, stderr io.Writer) int {
	if len(args) == 0 {
		return usageError(stderr, "branch requires list or create")
	}
	switch args[0] {
	case "list":
		opts, err := parseResearchOptions(args[1:], false, false, false)
		if err != nil {
			return usageError(stderr, err.Error())
		}
		p, err := checkedInitializedProfile(root, opts.workspace)
		if err != nil {
			return fail(stderr, err)
		}
		cmd := exec.Command("repo", "branches")
		cmd.Dir = p.Path
		cmd.Stdout = out
		cmd.Stderr = stderr
		if err := cmd.Run(); err != nil {
			return 1
		}
		return 0
	case "create":
		if len(args) < 2 {
			return usageError(stderr, "branch create requires a name")
		}
		name := args[1]
		if !validBranch(name) {
			return usageError(stderr, "invalid branch name: "+name)
		}
		opts, err := parseResearchOptions(args[2:], true, false, false)
		if err != nil {
			return usageError(stderr, err.Error())
		}
		p, err := checkedInitializedProfile(root, opts.workspace)
		if err != nil {
			return fail(stderr, err)
		}
		projects := "all manifest projects"
		if len(opts.projects) > 0 {
			projects = strings.Join(opts.projects, ", ")
		}
		fmt.Fprintf(out, "Branch : %s\nProjects: %s\nMode    : %s\n", name, projects, map[bool]string{false: "plan", true: "apply"}[opts.apply])
		if !opts.apply {
			fmt.Fprintln(out, "No changes were made. Add --apply to create the branch.")
			return 0
		}
		argv := append([]string{"start", name}, opts.projects...)
		return execute(root, p, "branch-create", argv, out, stderr)
	default:
		return usageError(stderr, "unknown branch subcommand: "+args[0])
	}
}
