package repo

import (
	"fmt"
	"regexp"
)

type initOptions struct {
	workspace         string
	partialClone      bool
	cloneFilter       string
	noUseSuperproject bool
	apply             bool
}

var cloneFilterPattern = regexp.MustCompile(`^[A-Za-z0-9:._+=-]+$`)

func parseInitOptions(args []string) (initOptions, error) {
	var result initOptions
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--workspace":
			if index+1 >= len(args) {
				return result, fmt.Errorf("missing value for --workspace")
			}
			index++
			result.workspace = args[index]
		case "--partial-clone":
			result.partialClone = true
		case "--clone-filter":
			if index+1 >= len(args) {
				return result, fmt.Errorf("missing value for --clone-filter")
			}
			index++
			if !cloneFilterPattern.MatchString(args[index]) {
				return result, fmt.Errorf("invalid --clone-filter value")
			}
			result.cloneFilter = args[index]
		case "--no-use-superproject":
			result.noUseSuperproject = true
		case "--apply":
			result.apply = true
		default:
			return result, fmt.Errorf("unknown option: %s", args[index])
		}
	}
	if result.cloneFilter != "" && !result.partialClone {
		return result, fmt.Errorf("--clone-filter requires --partial-clone")
	}
	return result, nil
}
