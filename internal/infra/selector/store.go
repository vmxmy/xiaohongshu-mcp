package selector

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type FileStore struct {
	Path string
}

func (s FileStore) Load() (map[string]string, error) {
	data, err := os.ReadFile(s.Path)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string)
	if err := yaml.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s FileStore) Save(selectors map[string]string) error {
	data, err := yaml.Marshal(selectors)
	if err != nil {
		return err
	}
	return os.WriteFile(s.Path, data, 0644)
}

func (s FileStore) Snapshot() (string, error) {
	ts := time.Now().Format("20060102-150405")
	dst := s.Path + "." + ts + ".bak"
	data, err := os.ReadFile(s.Path)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(dst, data, 0644); err != nil {
		return "", err
	}
	return filepath.Base(dst), nil
}

func (s FileStore) Rollback(snapshot string) error {
	src := filepath.Join(filepath.Dir(s.Path), snapshot)
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(s.Path, data, 0644)
}
