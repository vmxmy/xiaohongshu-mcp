package selector

import (
	"context"
	"testing"
)

type fakeStore struct {
	Selectors map[string]string
	Saved     bool
}

func (f *fakeStore) Load() (map[string]string, error) { return f.Selectors, nil }
func (f *fakeStore) Save(s map[string]string) error   { f.Saved = true; return nil }
func (f *fakeStore) Snapshot() (string, error)        { return "snap", nil }
func (f *fakeStore) Rollback(string) error            { return nil }

type fakeLearner struct{}

func (fakeLearner) Learn(ctx context.Context, current map[string]string) (map[string]string, error) {
	return map[string]string{"a": "b"}, nil
}

type fakeValidator struct{}

func (fakeValidator) Validate(ctx context.Context, selectors map[string]string) error { return nil }

func TestPipeline_Run_SavesSelectors(t *testing.T) {
	p := Pipeline{
		Store:    &fakeStore{Selectors: map[string]string{"a": "a"}},
		Learner:  fakeLearner{},
		Validate: fakeValidator{},
	}
	if err := p.Run(context.Background()); err != nil {
		t.Fatalf("run err: %v", err)
	}
	if !p.Store.(*fakeStore).Saved {
		t.Fatalf("expected Save to be called")
	}
}
