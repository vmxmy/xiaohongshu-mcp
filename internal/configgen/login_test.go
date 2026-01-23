package configgen

import "testing"

type fakeCookies struct{ saved bool }

func (f *fakeCookies) Save(path string) error {
	f.saved = true
	return nil
}

func TestLoginFlow_SavesCookies(t *testing.T) {
	f := &fakeCookies{}
	lf := LoginFlow{CookieSaver: f}
	if err := lf.SaveCookies("/tmp/cookies.json"); err != nil {
		t.Fatalf("save err: %v", err)
	}
	if !f.saved {
		t.Fatalf("expected cookies saved")
	}
}
