package repo

import "testing"

func TestParseInitOptions(t *testing.T) {
	got, err := parseInitOptions([]string{"--workspace", "android-14", "--partial-clone", "--clone-filter", "blob:limit=10M", "--no-use-superproject", "--apply"})
	if err != nil || got.workspace != "android-14" || !got.partialClone || got.cloneFilter != "blob:limit=10M" || !got.noUseSuperproject || !got.apply {
		t.Fatalf("got=%#v err=%v", got, err)
	}
	if _, err := parseInitOptions([]string{"--clone-filter", "blob:none"}); err == nil {
		t.Fatal("clone filter accepted without partial clone")
	}
	if _, err := parseInitOptions([]string{"--partial-clone", "--clone-filter", "$(unsafe)"}); err == nil {
		t.Fatal("unsafe clone filter accepted")
	}
}
