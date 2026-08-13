package repo

import "testing"

func TestParseOptions(t *testing.T) {
	got, err := parseOptions([]string{"--workspace", "android-14", "--jobs", "8", "--apply"}, true, true)
	if err != nil || got.workspace != "android-14" || got.jobs != 8 || !got.apply {
		t.Fatalf("got=%#v err=%v", got, err)
	}
	if _, err := parseOptions([]string{"--jobs", "0"}, true, true); err == nil {
		t.Fatal("zero jobs accepted")
	}
	if _, err := parseOptions([]string{"--apply"}, false, false); err == nil {
		t.Fatal("unsupported apply accepted")
	}
}
