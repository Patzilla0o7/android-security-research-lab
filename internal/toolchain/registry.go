package toolchain

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type Tool struct {
	ID, Label, Level, CheckType, Target, Minimum string
	Methods                                      []string
}

type Result struct {
	Tool   Tool
	Status string
	Detail string
}

var specPattern = regexp.MustCompile(`^\s*"([^"]+)"\s*$`)

func Load(path string) ([]Tool, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var tools []Tool
	scanner := bufio.NewScanner(file)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		match := specPattern.FindStringSubmatch(scanner.Text())
		if match == nil {
			continue
		}
		fields := strings.Split(match[1], "|")
		if len(fields) != 7 {
			return nil, fmt.Errorf("%s:%d: invalid tool specification", path, lineNumber)
		}
		tool := Tool{ID: fields[0], Label: fields[1], Level: fields[2], CheckType: fields[3], Target: fields[4], Minimum: fields[5], Methods: strings.Split(fields[6], ",")}
		if tool.ID == "" || tool.Label == "" || tool.Target == "" {
			return nil, fmt.Errorf("%s:%d: incomplete tool specification", path, lineNumber)
		}
		if tool.Level != "required" && tool.Level != "recommended" && tool.Level != "optional" {
			return nil, fmt.Errorf("%s:%d: invalid tool level %q", path, lineNumber, tool.Level)
		}
		if tool.CheckType != "command" && tool.CheckType != "java-major" && tool.CheckType != "package" {
			return nil, fmt.Errorf("%s:%d: invalid check type %q", path, lineNumber, tool.CheckType)
		}
		tools = append(tools, tool)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(tools) == 0 {
		return nil, fmt.Errorf("no tool specifications found in %s", path)
	}
	return tools, nil
}

func Check(tool Tool) Result {
	result := Result{Tool: tool, Status: "missing", Detail: "Not installed"}
	switch tool.CheckType {
	case "command":
		if _, err := exec.LookPath(tool.Target); err == nil {
			result.Status, result.Detail = "installed", commandDetail(tool)
		}
	case "package":
		cmd := exec.Command("dpkg-query", "-W", "-f=${db:Status-Status}", tool.Target)
		if output, err := cmd.Output(); err == nil && strings.TrimSpace(string(output)) == "installed" {
			result.Status, result.Detail = "installed", "Installed"
		}
	case "java-major":
		if _, err := exec.LookPath(tool.Target); err != nil {
			return result
		}
		output, err := exec.Command(tool.Target, "-version").CombinedOutput()
		major := javaMajor(string(output))
		minimum, _ := strconv.Atoi(tool.Minimum)
		if err != nil || major == 0 {
			result.Status, result.Detail = "version_mismatch", "Unable to determine version"
		} else if major < minimum {
			result.Status, result.Detail = "version_mismatch", fmt.Sprintf("%d (requires %d+)", major, minimum)
		} else {
			result.Status, result.Detail = "installed", strconv.Itoa(major)
		}
	}
	return result
}

func CheckAll(tools []Tool) []Result {
	results := make([]Result, 0, len(tools))
	for _, tool := range tools {
		results = append(results, Check(tool))
	}
	return results
}

func AptPackage(tool Tool) string {
	for _, method := range tool.Methods {
		if strings.HasPrefix(method, "apt:") {
			return strings.TrimPrefix(method, "apt:")
		}
	}
	return ""
}

func ManualMethod(tool Tool) string {
	for _, method := range tool.Methods {
		if strings.HasPrefix(method, "manual:") {
			return strings.TrimPrefix(method, "manual:")
		}
	}
	return ""
}

func javaMajor(output string) int {
	re := regexp.MustCompile(`version "([0-9]+)`)
	match := re.FindStringSubmatch(output)
	if len(match) != 2 {
		return 0
	}
	major, _ := strconv.Atoi(match[1])
	return major
}

func commandDetail(tool Tool) string {
	var args []string
	switch tool.ID {
	case "git", "python3", "ccache":
		args = []string{"--version"}
	case "repo":
		args = []string{"version"}
	default:
		return "Installed"
	}
	output, err := exec.Command(tool.Target, args...).CombinedOutput()
	if err != nil {
		return "Installed"
	}
	line := strings.Split(strings.TrimSpace(string(output)), "\n")[0]
	if line == "" {
		return "Installed"
	}
	return line
}
