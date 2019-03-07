package models

import (
	"github.com/gorilla/websocket"
	"time"
)

type Question struct {
	Question string   `json:"question"`
	Choices  []string `json:"choices"`
	Answer   string   `json:"-"` // don't want user to see this
}

type Answer struct {
	UserID int64  `json:"userID"`
	Answer string `json:"answer"`
	// to make sure user answered quick enough
	SentAt time.Time `json:"sentAt"`
}

type GameState struct {
	Players   map[int64]*websocket.Conn
	Questions []*Question
}
