package handlers

import (
	"time"

	"github.com/TriviaRoulette/servers/gateway/models/users"
)

// SessionState stores the time at which a session starts
// and the users.User that started the session
type SessionState struct {
	// time authenticated user initiated session
	SessionStart time.Time `json:"session_start"`

	// authenticated user
	User users.User `json:"user"`
}
