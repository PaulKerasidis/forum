package models

import (
	"database/sql"
	"errors"
	"time"

	"forum/utils"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")
)

// SessionRepository handles session-related database operations
type SessionRepository struct {
	DB *sql.DB
}

// NewSessionRepository creates a new SessionRepository
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{DB: db}
}

// Create creates a new session for a user
func (r *SessionRepository) Create(userID, ipAddress string) (*Session, error) {
	// First, delete any existing sessions for this user
	_, err := r.DB.Exec("DELETE FROM sessions WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}

	// Generate a new session ID
	sessionID := utils.GenerateSessionToken()
	expiresAt := utils.CalculateSessionExpiry()

	// Insert the new session
	_, err = r.DB.Exec(
		"INSERT INTO sessions (user_id, session_id, ip_address, created_at, expires_at) VALUES (?, ?, ?, ?, ?)",
		userID, sessionID, ipAddress, time.Now(), expiresAt,
	)
	if err != nil {
		return nil, err
	}

	// Return the session
	session := &Session{
		UserID:    userID,
		SessionID: sessionID,
		IPAddress: ipAddress,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
	}

	return session, nil
}

// GetBySessionID retrieves a session by its ID
func (r *SessionRepository) GetBySessionID(sessionID string) (*Session, error) {
	var session Session
	var createdStr, expiresStr string

	err := r.DB.QueryRow(
		"SELECT user_id, session_id, ip_address, created_at, expires_at FROM sessions WHERE session_id = ?",
		sessionID,
	).Scan(&session.UserID, &session.SessionID, &session.IPAddress, &createdStr, &expiresStr)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	// Parse timestamps
	session.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdStr)
	if err != nil {
		return nil, err
	}

	session.ExpiresAt, err = time.Parse("2006-01-02 15:04:05", expiresStr)
	if err != nil {
		return nil, err
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		// Delete the expired session
		_, _ = r.DB.Exec("DELETE FROM sessions WHERE session_id = ?", sessionID)
		return nil, ErrSessionExpired
	}

	return &session, nil
}

// Delete removes a session
func (r *SessionRepository) Delete(sessionID string) error {
	_, err := r.DB.Exec("DELETE FROM sessions WHERE session_id = ?", sessionID)
	return err
}
