package configgen

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPersist_BackupAndWrite(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(target, []byte("old: 1\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	p := Persister{Now: func() string { return "20260123-120000" }}
	if err := p.BackupAndWrite(target, []byte("new: 2\n")); err != nil {
		t.Fatalf("persist: %v", err)
	}
	if _, err := os.Stat(target + ".bak.20260123-120000"); err != nil {
		t.Fatalf("backup missing: %v", err)
	}
}
