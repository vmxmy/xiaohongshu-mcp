package main

import (
	"context"
	"encoding/base64"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/go-rod/rod/lib/proto"
)

type fakeQRElement struct {
	image []byte
}

func (f *fakeQRElement) Screenshot(_ proto.PageCaptureScreenshotFormat, _ int) ([]byte, error) {
	return f.image, nil
}

type fakeQRPage struct {
	navigateCalls int
	waitCalls     int
	lastURL       string
	has           map[string]bool
	text          string
	elements      map[string]*fakeQRElement
	elementR      *fakeQRElement
	frames        []qrFrame
	closed        bool
}

func (f *fakeQRPage) Navigate(_ context.Context, url string) error {
	f.navigateCalls++
	f.lastURL = url
	return nil
}

func (f *fakeQRPage) WaitLoad(_ context.Context) error {
	f.waitCalls++
	return nil
}

func (f *fakeQRPage) Has(_ context.Context, selector string) (bool, error) {
	if f.has == nil {
		return false, nil
	}
	return f.has[selector], nil
}

func (f *fakeQRPage) HasR(_ context.Context, _ string, jsRegex string) (bool, error) {
	if f.text == "" {
		return false, nil
	}
	re, err := regexp.Compile(jsRegex)
	if err != nil {
		return false, err
	}
	return re.MatchString(f.text), nil
}

func (f *fakeQRPage) Element(_ context.Context, selector string) (qrElement, error) {
	if f.elements == nil {
		return nil, errors.New("not found")
	}
	el, ok := f.elements[selector]
	if !ok {
		return nil, errors.New("not found")
	}
	return el, nil
}

func (f *fakeQRPage) ElementR(_ context.Context, _ string, _ string) (qrElement, error) {
	if f.elementR == nil {
		return nil, errors.New("not found")
	}
	return f.elementR, nil
}

func (f *fakeQRPage) Frames(_ context.Context) ([]qrFrame, error) {
	return f.frames, nil
}

func (f *fakeQRPage) Close() error {
	f.closed = true
	return nil
}

type fakeQRFrame struct {
	text     string
	elements map[string]*fakeQRElement
	frames   []qrFrame
}

func (f *fakeQRFrame) HasR(_ context.Context, _ string, jsRegex string) (bool, error) {
	if f.text == "" {
		return false, nil
	}
	re, err := regexp.Compile(jsRegex)
	if err != nil {
		return false, err
	}
	return re.MatchString(f.text), nil
}

func (f *fakeQRFrame) Element(_ context.Context, selector string) (qrElement, error) {
	if f.elements == nil {
		return nil, errors.New("not found")
	}
	el, ok := f.elements[selector]
	if !ok {
		return nil, errors.New("not found")
	}
	return el, nil
}

func (f *fakeQRFrame) ElementR(_ context.Context, _ string, _ string) (qrElement, error) {
	return nil, errors.New("not found")
}

func (f *fakeQRFrame) Frames(_ context.Context) ([]qrFrame, error) {
	return f.frames, nil
}

func TestRodLoginSession_Open(t *testing.T) {
	page := &fakeQRPage{}
	session := rodLoginSession{page: page, sleep: func(time.Duration) {}}

	if err := session.Open(context.Background()); err != nil {
		t.Fatalf("Open err: %v", err)
	}
	if page.navigateCalls != 1 || page.waitCalls != 1 {
		t.Fatalf("expected navigate and waitload called once")
	}
	if page.lastURL != xhsLoginURL {
		t.Fatalf("unexpected url: %s", page.lastURL)
	}
}

func TestRodLoginSession_LoggedIn(t *testing.T) {
	page := &fakeQRPage{has: map[string]bool{loginStatusSelector: true}}
	session := rodLoginSession{page: page, sleep: func(time.Duration) {}}

	got, err := session.LoggedIn(context.Background())
	if err != nil {
		t.Fatalf("LoggedIn err: %v", err)
	}
	if !got {
		t.Fatalf("expected logged in")
	}
}

func TestRodLoginSession_QRCode_LoginStage(t *testing.T) {
	page := &fakeQRPage{
		elements: map[string]*fakeQRElement{
			".login-container .qrcode-img": {image: []byte("png")},
		},
	}
	session := rodLoginSession{page: page, sleep: func(time.Duration) {}}

	got, err := session.QRCode(context.Background())
	if err != nil {
		t.Fatalf("QRCode err: %v", err)
	}
	if got.Stage != "login" {
		t.Fatalf("expected login stage")
	}
	if got.Image != base64.StdEncoding.EncodeToString([]byte("png")) {
		t.Fatalf("unexpected image data")
	}
}

func TestRodLoginSession_QRCode_SecurityStage(t *testing.T) {
	page := &fakeQRPage{
		text: "扫码验证",
		elements: map[string]*fakeQRElement{
			".login-container .qrcode-img": {image: []byte("png")},
		},
	}
	session := rodLoginSession{page: page, sleep: func(time.Duration) {}}

	got, err := session.QRCode(context.Background())
	if err != nil {
		t.Fatalf("QRCode err: %v", err)
	}
	if got.Stage != "security" {
		t.Fatalf("expected security stage")
	}
}

func TestRodLoginSession_QRCode_SecurityPrefersFrameQRCode(t *testing.T) {
	frame := &fakeQRFrame{
		text: "扫码验证",
		elements: map[string]*fakeQRElement{
			".login-container .qrcode-img": {image: []byte("frame")},
		},
	}
	page := &fakeQRPage{
		elements: map[string]*fakeQRElement{
			".login-container .qrcode-img": {image: []byte("page")},
		},
		frames: []qrFrame{frame},
	}
	session := rodLoginSession{page: page, sleep: func(time.Duration) {}}

	got, err := session.QRCode(context.Background())
	if err != nil {
		t.Fatalf("QRCode err: %v", err)
	}
	if got.Stage != "security" {
		t.Fatalf("expected security stage")
	}
	if got.Image != base64.StdEncoding.EncodeToString([]byte("frame")) {
		t.Fatalf("expected frame qrcode")
	}
}

func TestRodLoginSession_SaveCookies_UsesSaver(t *testing.T) {
	called := false
	session := rodLoginSession{saveCookies: func() error {
		called = true
		return nil
	}}

	if err := session.SaveCookies(); err != nil {
		t.Fatalf("SaveCookies err: %v", err)
	}
	if !called {
		t.Fatalf("expected saver called")
	}
}

func TestRodLoginSession_Close_ClosesPage(t *testing.T) {
	page := &fakeQRPage{}
	session := rodLoginSession{page: page}

	if err := session.Close(); err != nil {
		t.Fatalf("Close err: %v", err)
	}
	if !page.closed {
		t.Fatalf("expected page closed")
	}
}
