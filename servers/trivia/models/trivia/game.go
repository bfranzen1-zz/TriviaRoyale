package models

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Question struct {
	QuestionID int64     `json:"questionID" bson:"questionID"`
	Question   string    `json:"question" bson:"question"`
	Choices    []string  `json:"choices" bson:"choices"`
	Answer     string    `json:"-" bson:"answer"` // don't want user to see this
	SentAt     time.Time `json:"sentAt" bson:"SentAt"`
}

type Answer struct {
	LobbyID    bson.ObjectId `json:"lobbyID" bson:"lobbyID"`
	UserID     int64         `json:"userID" bson:"userID"`
	QuestionID int64         `json:"questionID" bson:"questionID"`
	Answer     string        `json:"answer" bson:"answer"`
	// to make sure user answered quick enough
	// ReceivedAt time.Time `json:"receivedAt"` // when client got question
	// SentAt     time.Time `json:"sentAt"`     // when client sent answer

}
