package publish

import "testing"

func TestNewGateway_ValidatesConfig(t *testing.T) {
	_, err := NewGateway(Config{}, nil)
	if err == nil {
		t.Fatalf("expected config error")
	}
}
