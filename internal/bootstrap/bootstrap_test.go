package bootstrap

import (
	"testing"

	"github.com/Patzilla0o7/android-security-research-lab/internal/toolchain"
)

func TestParseArgs(t *testing.T) {
	for _, test := range []struct {
		args []string
		want Mode
	}{{nil, ModePlan}, {[]string{"plan"}, ModePlan}, {[]string{"--apply"}, ModeApply}, {[]string{"--help"}, ModeHelp}} {
		got, err := ParseArgs(test.args)
		if err != nil || got != test.want {
			t.Fatalf("ParseArgs(%v) = %q, %v", test.args, got, err)
		}
	}
	if _, err := ParseArgs([]string{"invalid"}); err == nil {
		t.Fatal("invalid option accepted")
	}
}

func TestCollect(t *testing.T) {
	results := []toolchain.Result{
		{Tool: toolchain.Tool{Label: "Git"}, Status: "installed"},
		{Tool: toolchain.Tool{Label: "Java", Methods: []string{"apt:java"}}, Status: "missing"},
		{Tool: toolchain.Tool{Label: "Repo", Methods: []string{"manual:https://example.com"}}, Status: "missing"},
	}
	plan := Collect(results, func(pkg string) bool { return pkg == "java" })
	if len(plan.Satisfied) != 1 || len(plan.AptPackages) != 1 || len(plan.Manual) != 1 {
		t.Fatalf("unexpected plan: %#v", plan)
	}
}
