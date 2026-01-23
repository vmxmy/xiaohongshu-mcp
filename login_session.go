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

	if m.session == nil || m.expiredLocked() {
		_ = m.closeLocked()
		s, err := m.newSession()
		if err != nil {
			return loginQRResult{}, err
		}
		m.session = s
		m.openedAt = m.now()
	}

	if err := m.session.Open(ctx); err != nil {
		_ = m.closeLocked()
		return loginQRResult{}, err
	}

	loggedIn, err := m.session.LoggedIn(ctx)
	if err != nil {
		_ = m.closeLocked()
		return loginQRResult{}, err
	}
	if loggedIn {
		if err := m.session.SaveCookies(); err != nil {
			_ = m.closeLocked()
			return loginQRResult{}, err
		}
		_ = m.closeLocked()
		return loginQRResult{
			LoginQrcodeResponse: LoginQrcodeResponse{
				Timeout:    "0s",
				IsLoggedIn: true,
			},
		}, nil
	}

	qr, err := m.session.QRCode(ctx)
	if err != nil {
		_ = m.closeLocked()
		return loginQRResult{}, err
	}

	remaining := m.ttl - m.now().Sub(m.openedAt)
	if remaining < 0 {
		remaining = 0
	}
	return loginQRResult{
		LoginQrcodeResponse: LoginQrcodeResponse{
			Timeout:    remaining.String(),
			IsLoggedIn: false,
			Img:        qr.Image,
		},
		Stage: qr.Stage,
	}, nil
}

func (m *LoginManager) expiredLocked() bool {
	if m.ttl <= 0 {
		return false
	}
	return m.now().Sub(m.openedAt) > m.ttl
}

func (m *LoginManager) closeLocked() error {
	if m.session == nil {
		return nil
	}
	err := m.session.Close()
	m.session = nil
	return err
}
