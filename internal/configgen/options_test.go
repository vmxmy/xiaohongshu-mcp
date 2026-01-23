package configgen

import "testing"

func TestOptions_Defaults(t *testing.T) {
	opt := DefaultOptions()
	if opt.OutputPath == "" || !opt.Backup {
		t.Fatalf("expected default output path and backup=true")
	}
}
