package configgen

import "testing"

func TestInfer_UsesAriaLabel(t *testing.T) {
	snap := Snapshot{
		Nodes: []Node{
			{Tag: "button", Text: "发布", Attrs: map[string]string{"aria-label": "发布"}},
		},
	}
	inf := NewInfer()
	out := inf.Infer(snap)
	if out["submit"] == "" {
		t.Fatalf("expected submit selector")
	}
}
