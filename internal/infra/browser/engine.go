package browser

type Page interface {
	Goto(url string) error
	Click(selector string) error
	Fill(selector, value string) error
	SetFiles(selector string, files []string) error
	Text(selector string) (string, error)
	WaitVisible(selector string) error
	Close() error
}

type Engine interface {
	Start() error
	NewPage() (Page, error)
	Close() error
}
