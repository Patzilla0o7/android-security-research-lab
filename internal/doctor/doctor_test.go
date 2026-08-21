package doctor

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestReportTracksStatuses(t *testing.T) {
	var out bytes.Buffer
	r := &report{out: &out}
	r.good("Go", "installed")
	r.caution("ADB", "optional")
	r.bad("Repo", "missing")

	if r.pass != 1 || r.warn != 1 || r.fail != 1 {
		t.Fatalf("unexpected totals: pass=%d warn=%d fail=%d", r.pass, r.warn, r.fail)
	}
	if len(r.checks) != 3 || r.checks[2].Status != "failed" {
		t.Fatalf("unexpected checks: %#v", r.checks)
	}
}

func TestJSONReportShape(t *testing.T) {
	report := jsonReport{
		Checks: []check{{Status: "passed", Label: "Git", Detail: "installed"}},
		Passed: 1,
	}
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["passed"] != float64(1) || decoded["checks"] == nil {
		t.Fatalf("unexpected JSON: %s", data)
	}
}

func TestFailedChecksError(t *testing.T) {
	err := FailedChecksError{Count: 2}
	if err.Error() != "2 required check(s) failed" {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}
