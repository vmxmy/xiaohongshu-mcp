package configgen

import "testing"

type fakeProbe struct{ called bool }

func (f *fakeProbe) Probe() Snapshot {
	f.called = true
	return Snapshot{}
}

type fakeInfer struct{ called bool }

func (f *fakeInfer) Infer(s Snapshot) map[string]string {
	f.called = true
	return map[string]string{}
}

type fakePersist struct{ called bool }

func (f *fakePersist) Write(path string, data []byte) error {
	f.called = true
	return nil
}

func TestRunner_ExecutesPipeline(t *testing.T) {
	r := Runner{
		Probe:   &fakeProbe{},
		Infer:   &fakeInfer{},
		Persist: &fakePersist{},
	}
	if err := r.Run("config.yaml"); err != nil {
		t.Fatalf("run err: %v", err)
	}
}
