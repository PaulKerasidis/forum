package repository

import (
	"database/sql"
	"errors"
	"net"
	"time"

	"github.com/PaulKerasidis/forum/internal/models"
	"github.com/PaulKerasidis/forum/internal/utils"
)

type SessionRepository struct {
	DB *sql.DB
}

// NewSessionRepository creates a new SessionRepository
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{DB: db}
}

// Update your session_repository.go CreateSession method:

func (r *SessionRepository) CreateSession(userID, ipAddress string) (*models.Session, error) {
	return utils.ExecuteInTransactionWithResult(r.DB, func(tx *sql.Tx) (*models.Session, error) {
		// First, delete any existing sessions for this user
		_, err := tx.Exec("DELETE FROM sessions WHERE user_id = ?", userID)
		if err != nil {
			return nil, err
		}

		// Generate a new session ID and calculate expiry
		sessionID, err := utils.GenerateSessionToken()
		if err != nil {
			return nil, err
		}

		expiresAt := utils.CalculateSessionExpiry()
		now := time.Now()

		// FIXED: Clean the IP address before storing (remove port if present)
		cleanIP := cleanIPAddress(ipAddress)

		// Insert the new session with clean IP
		_, err = tx.Exec(
			"INSERT INTO sessions (user_id, session_id, ip_address, created_at, expires_at) VALUES (?, ?, ?, ?, ?)",
			userID, sessionID, cleanIP, now, expiresAt, // ← Use cleanIP here
		)
		if err != nil {
			return nil, err
		}

		// Return the session with clean IP
		session := &models.Session{
			UserID:    userID,
			SessionID: sessionID,
			IPAddress: cleanIP, // ← Use cleanIP here too
			CreatedAt: now,
			ExpiresAt: expiresAt,
		}

		return session, nil
	})
}

// Add this helper function to session_repository.go:
func cleanIPAddress(ipWithPossiblePort string) string {
	host, _, err := net.SplitHostPort(ipWithPossiblePort)
	if err != nil {
		// If SplitHostPort fails, it's probably just an IP without port
		return ipWithPossiblePort
	}
	return host
}

// GetBySessionID retrieves a session by its ID
func (sr *SessionRepository) GetBySessionID(sessionID string) (*models.Session, error) {
	var session models.Session

	err := sr.DB.QueryRow(
		"SELECT user_id, session_id, ip_address, created_at, expires_at FROM sessions WHERE session_id = ?",
		sessionID,
	).Scan(&session.UserID, &session.SessionID, &session.IPAddress, &session.CreatedAt, &session.ExpiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("session not found")
		}
		return nil, err
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		// Delete the expired session
		_, _ = sr.DB.Exec("DELETE FROM sessions WHERE session_id = ?", sessionID)
		return nil, errors.New("session expired")
	}

	return &session, nil

}

// // DeleteSession deletes a session by its ID
func (sr *SessionRepository) DeleteSession(sessionID string) error {
	return utils.ExecuteInTransaction(sr.DB, func(tx *sql.Tx) error {
		_, err := sr.DB.Exec("DELETE FROM sessions WHERE session_id = ?", sessionID)
		if err != nil {
			return err
		}

		// Commit the transaction
		return nil
	})
}

// Add this method to your session_repository.go:

// UpdateSessionIP updates the IP address for an existing session
func (sr *SessionRepository) UpdateSessionIP(sessionID, newIP string) error {
	return utils.ExecuteInTransaction(sr.DB, func(tx *sql.Tx) error {
		result, err := tx.Exec("UPDATE sessions SET ip_address = ? WHERE session_id = ?", newIP, sessionID)
		if err != nil {
			return err
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if rowsAffected == 0 {
			return errors.New("session not found")
		}

		return nil
	})
}
