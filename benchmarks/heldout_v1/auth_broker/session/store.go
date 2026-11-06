package session

import (
	"database/sql"
)

type SessionStore struct {
	db *sql.DB
}

func NewSessionStore(db *sql.DB) *SessionStore {
	return &SessionStore{db: db}
}

// Neg_Control_ValidateSessionSafe uses safe parameterized query and error handling
func (s *SessionStore) ValidateSessionSafe(token string) (int, error) {
	const query = "SELECT user_id FROM sessions WHERE token = ?"
	var userID int
	err := s.db.QueryRow(query, token).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

// DEF-HOLDOUT-07: Unhandled error on critical db.Exec
func (s *SessionStore) InvalidateSessionUnchecked(sessionID string) {
	s.db.Exec("UPDATE sessions SET active = false WHERE id = ?", sessionID) // UNHANDLED_ERROR
}

// DEF-HOLDOUT-08: Dynamic SQL string concatenation
func (s *SessionStore) FindSessionDynamic(rawToken string) (*sql.Rows, error) {
	query := "SELECT id, user_id FROM sessions WHERE token = '" + rawToken + "'"
	return s.db.Query(query)
}

