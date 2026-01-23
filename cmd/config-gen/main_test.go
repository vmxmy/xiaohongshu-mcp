package main

import "testing"

func TestParseFlags_Defaults(t *testing.T) {
	opt := parseFlags([]string{})
	if opt.OutputPath == "" || !opt.Backup {
		t.Fatalf("expected defaults")
	}
}
