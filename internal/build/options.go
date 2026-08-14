package build

import (
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

type options struct {
	workspace string
	target    string
	jobs      int
	modules   []string
	apply     bool
}

var buildValuePattern = regexp.MustCompile(`^[A-Za-z0-9._+:/-]+$`)

func parseOptions(args []string, allowApply bool) (options, error) {
	var result options
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--workspace":
			if index+1 >= len(args) {
				return result, fmt.Errorf("missing value for --workspace")
			}
			index++
			result.workspace = args[index]
		case "--target":
			if index+1 >= len(args) {
				return result, fmt.Errorf("missing value for --target")
			}
			index++
			if !validBuildValue(args[index]) {
				return result, fmt.Errorf("invalid --target value")
			}
			result.target = args[index]
		case "--jobs":
			if index+1 >= len(args) {
				return result, fmt.Errorf("missing value for --jobs")
			}
			index++
			jobs, err := strconv.Atoi(args[index])
			if err != nil || jobs < 1 {
				return result, fmt.Errorf("--jobs must be a positive integer")
			}
			result.jobs = jobs
		case "--module":
			if index+1 >= len(args) {
				return result, fmt.Errorf("missing value for --module")
			}
			index++
			if !validBuildValue(args[index]) {
				return result, fmt.Errorf("invalid --module value")
			}
			result.modules = append(result.modules, args[index])
		case "--apply":
			if !allowApply {
				return result, fmt.Errorf("--apply is not supported by this command")
			}
			result.apply = true
		default:
			return result, fmt.Errorf("unknown option: %s", args[index])
		}
	}
	if result.jobs == 0 {
		result.jobs = runtime.NumCPU()
	}
	return result, nil
}

func validBuildValue(value string) bool {
	return value != "" && buildValuePattern.MatchString(value) && !strings.Contains(value, "..") && !strings.HasPrefix(value, "/")
}
