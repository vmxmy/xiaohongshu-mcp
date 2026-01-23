package configgen

import "os"

type Persister struct {
	Now func() string
}

func (p Persister) BackupAndWrite(path string, data []byte) error {
	if p.Now == nil {
		p.Now = func() string { return "unknown" }
	}
	if _, err := os.Stat(path); err == nil {
		backup := path + ".bak." + p.Now()
		old, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.WriteFile(backup, old, 0644); err != nil {
			return err
		}
	}
	return os.WriteFile(path, data, 0644)
}
