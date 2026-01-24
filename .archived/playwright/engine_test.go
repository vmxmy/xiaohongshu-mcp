package playwright

import "testing"

func TestEngineConfig_Defaults(t *testing.T) {
	cfg := DefaultConfig()
	if !cfg.Headless {
		t.Fatalf("expected headless true")
	}
}
