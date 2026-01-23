package main

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
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
}

const (
	loginStatusLoggedIn       = "logged_in"
	loginStatusLoginRequired  = "login_required"
	loginStatusSecurityNeeded = "security_required"
)

type LoginManager struct {
	mu         sync.Mutex
	session    loginSession
	newSession func() (loginSession, error)
	ttl        time.Duration
	now        func() time.Time
	openedAt   time.Time
	sessionID  string

	newSessionID func() string
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
	if m.newSessionID == nil {
		m.newSessionID = func() string {
			return strconv.FormatInt(m.now().UnixNano(), 10)
		}
	}

	expired := m.session != nil && m.expiredLocked()
	if m.session == nil || expired {
		_ = m.closeLocked()
		s, err := m.newSession()
		if err != nil {
			return loginQRResult{}, err
		}
		m.session = s
		m.openedAt = m.now()
		m.sessionID = m.newSessionID()
		logrus.WithField("session_id", m.sessionID).Info("login session created")
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
		logrus.WithField("session_id", m.sessionID).Info("login status logged_in")
		sessionID := m.sessionID
		_ = m.closeLocked()
		return loginQRResult{
			LoginQrcodeResponse: LoginQrcodeResponse{
				Timeout:    "0s",
				IsLoggedIn: true,
				Status:     loginStatusLoggedIn,
				SessionID:  sessionID,
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
	status := loginStatusLoginRequired
	if qr.Stage == "security" {
		status = loginStatusSecurityNeeded
	}
	logrus.WithFields(logrus.Fields{
		"session_id": m.sessionID,
		"status":     status,
		"stage":      qr.Stage,
		"timeout":    remaining.String(),
	}).Info("login qrcode status")
	return loginQRResult{
		LoginQrcodeResponse: LoginQrcodeResponse{
			Timeout:    remaining.String(),
			IsLoggedIn: false,
			Img:        qr.Image,
			Stage:      qr.Stage,
			Status:     status,
			SessionID:  m.sessionID,
		},
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
	m.sessionID = ""
	return err
}
