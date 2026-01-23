package configgen

type CookieSaver interface {
	Save(path string) error
}

type LoginFlow struct {
	CookieSaver CookieSaver
}

func (l LoginFlow) SaveCookies(path string) error {
	if l.CookieSaver == nil {
		return nil
	}
	return l.CookieSaver.Save(path)
}
