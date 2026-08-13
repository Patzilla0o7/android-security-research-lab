package repo

import (
	"fmt"
	"strconv"
	"strings"
)

type researchOptions struct {
	workspace string
	projects  []string
	apply     bool
	commits   int
	file      string
}

func parseResearchOptions(args []string, allowApply, allowCommits, allowFile bool) (researchOptions, error) {
	var result researchOptions
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--workspace":
			if i+1 >= len(args) {
				return result, fmt.Errorf("missing value for --workspace")
			}
			i++
			result.workspace = args[i]
		case "--project":
			if i+1 >= len(args) {
				return result, fmt.Errorf("missing value for --project")
			}
			i++
			if !validProject(args[i]) {
				return result, fmt.Errorf("invalid project path: %s", args[i])
			}
			result.projects = append(result.projects, args[i])
		case "--apply":
			if !allowApply {
				return result, fmt.Errorf("--apply is not supported")
			}
			result.apply = true
		case "--commits":
			if !allowCommits || i+1 >= len(args) {
				return result, fmt.Errorf("--commits is not supported or missing a value")
			}
			i++
			n, err := strconv.Atoi(args[i])
			if err != nil || n < 0 {
				return result, fmt.Errorf("--commits must be zero or positive")
			}
			result.commits = n
		case "--file":
			if !allowFile || i+1 >= len(args) {
				return result, fmt.Errorf("--file is not supported or missing a value")
			}
			i++
			result.file = args[i]
		default:
			return result, fmt.Errorf("unknown option: %s", args[i])
		}
	}
	return result, nil
}

func validBranch(name string) bool {
	return name != "" && !strings.HasPrefix(name, "-") && !strings.ContainsAny(name, " ~^:?*[\\") && !strings.Contains(name, "..") && !strings.HasSuffix(name, ".") && !strings.HasSuffix(name, ".lock")
}
