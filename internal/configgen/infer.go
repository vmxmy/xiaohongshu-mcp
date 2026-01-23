package configgen

type Infer struct{}

func NewInfer() Infer {
	return Infer{}
}

func (i Infer) Infer(s Snapshot) map[string]string {
	out := make(map[string]string)
	for _, node := range s.Nodes {
		if node.Attrs["aria-label"] == "发布" {
			out["submit"] = "[aria-label='发布']"
		}
	}
	return out
}
