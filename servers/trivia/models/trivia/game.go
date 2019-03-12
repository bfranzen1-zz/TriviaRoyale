package models

import (
	"time"
)

type Question struct {
	QuestionID int64     `json:"questionID"`
	Question   string    `json:"question"`
	Choices    []string  `json:"choices"`
	Answer     string    `json:"-"` // don't want user to see this
	SentAt     time.Time `json:"sentAt"`
}

type Answer struct {
	LobbyID    int64  `json:"lobbyID"`
	UserID     int64  `json:"userID"`
	QuestionID int64  `json:"questionID"`
	Answer     string `json:"answer"`
	// to make sure user answered quick enough
	// ReceivedAt time.Time `json:"receivedAt"` // when client got question
	// SentAt     time.Time `json:"sentAt"`     // when client sent answer

}
