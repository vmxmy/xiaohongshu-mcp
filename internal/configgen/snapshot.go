package configgen

type Node struct {
	Tag   string
	Text  string
	Attrs map[string]string
}

type Snapshot struct {
	Nodes []Node
}
