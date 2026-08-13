package repo

import (
	"fmt"
	"strconv"
)

type options struct {
	workspace string
	jobs      int
	apply     bool
}

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
		default:
			return result, fmt.Errorf("unknown option: %s", args[index])
		}
	}
	return result, nil
}
