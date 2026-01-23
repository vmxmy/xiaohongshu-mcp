package configgen

type Prober interface {
	Probe() Snapshot
}

type Inferer interface {
	Infer(Snapshot) map[string]string
}

type Persist interface {
	Write(path string, data []byte) error
}

type Runner struct {
	Probe   Prober
	Infer   Inferer
	Persist Persist
}

func (r Runner) Run(path string) error {
	snap := r.Probe.Probe()
	selectors := r.Infer.Infer(snap)
	_ = selectors
	return r.Persist.Write(path, []byte("selectors: {}\n"))
}
