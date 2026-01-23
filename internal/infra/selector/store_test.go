package selector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStore_SnapshotAndRollback(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "selectors.yaml")
	if err := os.WriteFile(path, []byte("a: b\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	store := Store{Path: path}
	snap, err := store.Snapshot()
	if err != nil || snap == "" {
		t.Fatalf("snapshot err: %v", err)
	}
	if err := store.Save(map[string]string{"a": "c"}); err != nil {
		t.Fatalf("save err: %v", err)
	}
	if err := store.Rollback(snap); err != nil {
		t.Fatalf("rollback err: %v", err)
	}
}
