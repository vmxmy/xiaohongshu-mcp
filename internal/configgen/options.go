package configgen

type Options struct {
	OutputPath string
	Backup     bool
	Headless   bool
	DryRun     bool
	VerifyOnly bool
}

func DefaultOptions() Options {
	return Options{
		OutputPath: "config.yaml",
		Backup:     true,
		Headless:   true,
	}
}
