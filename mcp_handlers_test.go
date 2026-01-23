package main

import (
	"context"
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
				},
				Stage: "security",
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
