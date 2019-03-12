package models

import (
	"time"
)

type Question struct {
	QuestionNum int64    `json:"question_number"`
	Question    string   `json:"question"`
	Choices     []string `json:"choices"`
	Answer      string   `json:"-"` // don't want user to see this

}

type Answer struct {
	UserID int64  `json:"userID"`
	Answer string `json:"answer"`
	// to make sure user answered quick enough
	SentAt time.Time `json:"sentAt"`
}
