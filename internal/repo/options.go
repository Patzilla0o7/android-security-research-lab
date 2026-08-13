package repo

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type options struct {
	workspace     string
	jobs          int
	retryFetches  int
	apply         bool
	noCloneBundle bool
	forceSync     bool
	projects      []string
}

var projectPattern = regexp.MustCompile(`^[A-Za-z0-9._+/-]+$`)

func parseOptions(args []string, allowJobs, allowApply bool) (options, error) {
	var result options
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--workspace":
			if index+1 >= len(args) {
				return result, fmt.Errorf("missing value for --workspace")
			}
			index++
			result.workspace = args[index]
		case "--jobs":
			if !allowJobs {
				return result, fmt.Errorf("--jobs is not supported by this command")
			}
			if index+1 >= len(args) {
				return result, fmt.Errorf("missing value for --jobs")
			}
			index++
			jobs, err := strconv.Atoi(args[index])
			if err != nil || jobs < 1 {
				return result, fmt.Errorf("--jobs must be a positive integer")
			}
			result.jobs = jobs
		case "--apply":
			if !allowApply {
				return result, fmt.Errorf("--apply is not supported by this command")
			}
			result.apply = true
		case "--project":
			if !allowJobs {
				return result, fmt.Errorf("--project is not supported by this command")
			}
			if index+1 >= len(args) {
				return result, fmt.Errorf("missing value for --project")
			}
			index++
			project := args[index]
			if !validProject(project) {
				return result, fmt.Errorf("invalid project path: %s", project)
			}
			result.projects = append(result.projects, project)
		case "--retry-fetches":
			if !allowJobs {
				return result, fmt.Errorf("--retry-fetches is not supported by this command")
			}
			if index+1 >= len(args) {
				return result, fmt.Errorf("missing value for --retry-fetches")
			}
			index++
			retries, err := strconv.Atoi(args[index])
			if err != nil || retries < 0 {
				return result, fmt.Errorf("--retry-fetches must be zero or a positive integer")
			}
			result.retryFetches = retries
		case "--no-clone-bundle":
			if !allowJobs {
				return result, fmt.Errorf("--no-clone-bundle is not supported by this command")
			}
			result.noCloneBundle = true
		case "--force-sync":
			if !allowJobs {
				return result, fmt.Errorf("--force-sync is not supported by this command")
			}
			result.forceSync = true
		default:
			return result, fmt.Errorf("unknown option: %s", args[index])
		}
	}
	return result, nil
}

func validProject(project string) bool {
	if !projectPattern.MatchString(project) || strings.HasPrefix(project, "/") {
		return false
	}
	for _, part := range strings.Split(project, "/") {
		if part == "" || part == "." || part == ".." {
			return false
		}
	}
	return true
}
