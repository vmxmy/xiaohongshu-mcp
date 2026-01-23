package main

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"
)

type fakeLoginProvider struct {
	result loginQRResult
	err    error
}

func (f fakeLoginProvider) GetQRCode(ctx context.Context) (loginQRResult, error) {
	return f.result, f.err
}

func TestLoginQrcodeHandler_TextForSecurityStage(t *testing.T) {
	service := &XiaohongshuService{
		loginManager: fakeLoginProvider{
			result: loginQRResult{
				LoginQrcodeResponse: LoginQrcodeResponse{
					Timeout:    "4m0s",
					IsLoggedIn: false,
					Img:        "img",
					Stage:      "security",
				},
			},
		},
	}
	app := &AppServer{xiaohongshuService: service}

	result := app.handleGetLoginQrcode(context.Background())
	if result == nil || len(result.Content) == 0 {
		t.Fatalf("expected content")
	}
	if !strings.Contains(result.Content[0].Text, "安全认证") {
		t.Fatalf("expected security text")
	}
}

func TestLoginQrcodeHandler_TextIncludesStatusAndSession(t *testing.T) {
	service := &XiaohongshuService{
		loginManager: fakeLoginProvider{
			result: loginQRResult{
				LoginQrcodeResponse: LoginQrcodeResponse{
					Timeout:    "4m0s",
					IsLoggedIn: false,
					Img:        "img",
					Stage:      "security",
					Status:     loginStatusSecurityNeeded,
					SessionID:  "sess-1",
				},
			},
		},
	}
	app := &AppServer{xiaohongshuService: service}

	result := app.handleGetLoginQrcode(context.Background())
	if result == nil || len(result.Content) == 0 {
		t.Fatalf("expected content")
	}
	if !strings.Contains(result.Content[0].Text, "状态:") {
		t.Fatalf("expected status text")
	}
	if !strings.Contains(result.Content[0].Text, "sess-1") {
		t.Fatalf("expected session id")
	}
}

func TestParseSyncCookiesPayload_Base64(t *testing.T) {
	data := []byte(`[{"name":"a"}]`)
	args := SyncCookiesArgs{CookiesBase64: base64.StdEncoding.EncodeToString(data)}

	got, err := parseSyncCookiesPayload(args)
	if err != nil {
		t.Fatalf("parse err: %v", err)
	}
	if string(got) != string(data) {
		t.Fatalf("unexpected payload")
	}
}

func TestParseSyncCookiesPayload_JSON(t *testing.T) {
	data := []byte(`[{"name":"a"}]`)
	args := SyncCookiesArgs{CookiesJSON: string(data)}

	got, err := parseSyncCookiesPayload(args)
	if err != nil {
		t.Fatalf("parse err: %v", err)
	}
	if string(got) != string(data) {
		t.Fatalf("unexpected payload")
	}
}

func TestParseSyncCookiesPayload_Missing(t *testing.T) {
	_, err := parseSyncCookiesPayload(SyncCookiesArgs{})
	if err == nil {
		t.Fatalf("expected error")
	}
}
