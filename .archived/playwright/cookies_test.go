package playwright

import (
	"testing"

	"github.com/playwright-community/playwright-go"
)

func TestParseCookies_Rod(t *testing.T) {
	data := []byte(`[{"name":"a","value":"1","domain":".xiaohongshu.com","path":"/","expires":123,"httpOnly":true,"secure":false,"sameSite":"Lax"}]`)
	cookies, err := parseCookies(data)
	if err != nil {
		t.Fatalf("parse cookies err: %v", err)
	}
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	if cookies[0].Domain == nil || *cookies[0].Domain != ".xiaohongshu.com" {
		t.Fatalf("expected domain set")
	}
	if cookies[0].SameSite == nil || cookies[0].SameSite != playwright.SameSiteAttributeLax {
		t.Fatalf("expected samesite lax")
	}
}

func TestParseCookies_StorageState(t *testing.T) {
	data := []byte(`{"cookies":[{"name":"a","value":"1","domain":".xiaohongshu.com","path":"/"}]}`)
	cookies, err := parseCookies(data)
	if err != nil {
		t.Fatalf("parse cookies err: %v", err)
	}
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
}

func TestParseCookies_Empty(t *testing.T) {
	cookies, err := parseCookies([]byte(" \n"))
	if err != nil {
		t.Fatalf("parse cookies err: %v", err)
	}
	if len(cookies) != 0 {
		t.Fatalf("expected empty cookies")
	}
}
