package main

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"
)

// Session represents a user session
type Session struct {
	Token     string
	Username  string
	ExpiresAt time.Time
}

// AuthManager handles authentication and session management
type AuthManager struct {
	config   *AuthConfig
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(config *AuthConfig) *AuthManager {
	am := &AuthManager{
		config:   config,
		sessions: make(map[string]*Session),
	}

	// Start session cleanup routine
	go am.cleanupExpiredSessions()

	return am
}

// Login authenticates a user and creates a session
func (am *AuthManager) Login(username, password string) (string, error) {
	if !am.config.Enabled {
		return "", nil
	}

	if username != am.config.Username || password != am.config.Password {
		return "", http.ErrAbortHandler
	}

	// Generate session token
	token := am.generateToken()

	am.mu.Lock()
	am.sessions[token] = &Session{
		Token:     token,
		Username:  username,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	am.mu.Unlock()

	return token, nil
}

// Logout removes a session
func (am *AuthManager) Logout(token string) {
	am.mu.Lock()
	delete(am.sessions, token)
	am.mu.Unlock()
}

// ValidateSession checks if a session token is valid
func (am *AuthManager) ValidateSession(token string) bool {
	if !am.config.Enabled {
		return true
	}

	am.mu.RLock()
	session, exists := am.sessions[token]
	am.mu.RUnlock()

	if !exists {
		return false
	}

	return time.Now().Before(session.ExpiresAt)
}

// AuthMiddleware is a middleware that requires authentication
func (am *AuthManager) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !am.config.Enabled {
			next(w, r)
			return
		}

		// Check for session cookie
		cookie, err := r.Cookie("session_token")
		if err != nil || !am.ValidateSession(cookie.Value) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

// generateToken generates a random session token
func (am *AuthManager) generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// cleanupExpiredSessions periodically removes expired sessions
func (am *AuthManager) cleanupExpiredSessions() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		am.mu.Lock()
		now := time.Now()
		for token, session := range am.sessions {
			if now.After(session.ExpiresAt) {
				delete(am.sessions, token)
			}
		}
		am.mu.Unlock()
	}
}
