package main

import (
	"context"
	"testing"
	"time"
)

type fakeLoginSession struct {
	openCalls int
	qrCalls   int
	loggedIn  bool
	qr        loginQRCode
	saved     bool
	closed    bool
}

func (f *fakeLoginSession) Open(ctx context.Context) error {
	f.openCalls++
	return nil
}

func (f *fakeLoginSession) LoggedIn(ctx context.Context) (bool, error) {
	return f.loggedIn, nil
}

func (f *fakeLoginSession) QRCode(ctx context.Context) (loginQRCode, error) {
	f.qrCalls++
	return f.qr, nil
}

func (f *fakeLoginSession) SaveCookies() error {
	f.saved = true
	return nil
}

func (f *fakeLoginSession) Close() error {
	f.closed = true
	return nil
}

func TestLoginManager_ReturnsQRCodeAndKeepsSession(t *testing.T) {
	clock := time.Date(2026, 1, 23, 10, 0, 0, 0, time.UTC)
	s := &fakeLoginSession{qr: loginQRCode{Image: "img", Stage: "login"}}
	m := NewLoginManager(func() (loginSession, error) { return s, nil }, 4*time.Minute)
	m.now = func() time.Time { return clock }

	got, err := m.GetQRCode(context.Background())
	if err != nil {
		t.Fatalf("GetQRCode err: %v", err)
	}
	if got.Img != "img" || got.Stage != "login" || got.IsLoggedIn {
		t.Fatalf("unexpected qr result: %+v", got)
	}
	if s.openCalls != 1 || s.qrCalls != 1 {
		t.Fatalf("expected session used once")
	}
}
