package main

import (
	"context"
	"sync"
	"time"
)

type loginSession interface {
	Open(ctx context.Context) error
	LoggedIn(ctx context.Context) (bool, error)
	QRCode(ctx context.Context) (loginQRCode, error)
	SaveCookies() error
	Close() error
}

type loginQRCode struct {
	Image string
	Stage string
}

type loginQRResult struct {
	LoginQrcodeResponse
	Stage string
}

type LoginManager struct {
	mu         sync.Mutex
	session    loginSession
	newSession func() (loginSession, error)
	ttl        time.Duration
	now        func() time.Time
	openedAt   time.Time
}

func NewLoginManager(newSession func() (loginSession, error), ttl time.Duration) *LoginManager {
	return &LoginManager{
		newSession: newSession,
		ttl:        ttl,
	}
}

func (m *LoginManager) GetQRCode(ctx context.Context) (loginQRResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.now == nil {
		m.now = time.Now
	}

	if m.session == nil {
		s, err := m.newSession()
		if err != nil {
			return loginQRResult{}, err
		}
		m.session = s
		m.openedAt = m.now()
	}

	if err := m.session.Open(ctx); err != nil {
		return loginQRResult{}, err
	}

	qr, err := m.session.QRCode(ctx)
	if err != nil {
		return loginQRResult{}, err
	}

	return loginQRResult{
		LoginQrcodeResponse: LoginQrcodeResponse{
			Timeout:    m.ttl.String(),
			IsLoggedIn: false,
			Img:        qr.Image,
		},
		Stage: qr.Stage,
	}, nil
}
