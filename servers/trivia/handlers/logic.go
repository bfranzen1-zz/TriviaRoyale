package handlers

import (
	"fmt"
	t "github.com/TriviaRoulette/servers/trivia/models/trivia"
	"github.com/TriviaRoulette/servers/trivia/models/users"
	"gopkg.in/mgo.v2/bson"
	"time"
)

// GameState contains information for a specific game
type GameState struct {
	Players   []int64              `json:"players" bson:"players"`
	Questions []t.Question         `json:"questions" bson:"questions"`
	Answers   map[int64][]t.Answer `json:"answers" bson:"answers"`
}

// Lobby contains information for a specific lobby instance
type Lobby struct {
	LobbyID    bson.ObjectId `json:"lobbyId" bson:"_id"`
	State      *GameState    `json:"state" bson:"state"`
	Options    *Options      `json:"options" bson:"options"`
	Creator    *users.User   `json:"creator" bson:"creator"`
	Over       bool          `json:"over" bson:"over"`
	InProgress bool          `json:"inProgress" bson:"in_progress"`
}

// Options stores information that defines how a game will be structured
type Options struct {
	NumQuestions int64  `json:"numQuestions" bson:"numQuestions"`
	Category     int64  `json:"category" bson:"category"`
	Difficulty   string `json:"difficulty" bson:"difficulty"`
	MaxPlayers   int64  `json:"maxPlayers" bson:"maxPlayers"`
}

// UserStatistic contains information about a players performance in a game
type UserStatistic struct {
	MongoID bson.ObjectId `json:"mongoId" bson:"_id"`
	GameID  bson.ObjectId `json:"gameId" bson:"gameId"`
	UserID  int64         `json:"userId" bson:"userId"`
	Correct int64         `json:"correct" bson:"correct"`
	Won     bool          `json:"won" bson:"won"`
	Points  int64         `json:"points" bson:"points"`
}

/*
- when lobby owner starts game we get game-start message
- start looping over questions
			- publish question to queue
			- wait x seconds
			- read off queue and remove people with wrong answers
- after loop
	- save results
	- type game-end
		- results
	- remove lobby from context
*/

func (ctx *TriviaContext) StartGame(lob *Lobby) {
	// TODO: start go routine for specific lobby questions
	lob.InProgress = true
	start := TriviaMessage{
		Type:    "game-start",
		Lobby:   lob,
		UserIDs: lob.State.Players,
	}
	ctx.PublishData(start)
	for i, q := range lob.State.Questions {
		q.SentAt = time.Now().UTC()
		ctx.PublishData(q)
		switch lob.Options.Difficulty {
		case "easy":
			time.Sleep(31 * time.Second)
		case "medium":
			time.Sleep(21 * time.Second)
		case "hard":
			time.Sleep(11 * time.Second)
		}
		if i != len(lob.State.Questions)-1 {
			ctx.handleAnswers(&q, false, lob)
		} else {
			ctx.handleAnswers(&q, true, lob)
		}

	}
}

// SaveResults saves the results of a game for a specific user specified by id
// it will generate a record in the mongo store that saves the points they got for the current game
// as well as number of correct questions and other information
func (ctx *TriviaContext) SaveResults(id int64, over bool, lob *Lobby, currQ int64) {
	// TODO: save results of game in mongo store
	// remove lobby from context
	diff := diffAsInt(lob.Options.Difficulty)

	stat := &UserStatistic{
		MongoID: bson.NewObjectId(),
		GameID:  lob.LobbyID,
		UserID:  id,
		Correct: currQ - 1,
		Won:     over,
	}
	if over {
		stat.Points = 10 + ((currQ - 1) * (diff / int64(len(lob.State.Players))))
	} else {
		stat.Points = 1 + ((currQ - 1) * (diff / int64(len(lob.State.Players))))
	}
	if err := ctx.Mongo.Insert(stat, "user_stats"); err != nil {
		fmt.Printf("error inserting record, %v", err)
	}
}

func (ctx *TriviaContext) handleAnswers(q *t.Question, end bool, lob *Lobby) {
	ans := lob.State.Answers[q.QuestionID]
	if len(ans) != len(lob.State.Players) { // check if all players responded in time
		for _, p := range lob.State.Players {
			if ok := checkPlayer(ans, p); !ok {
				ctx.SaveResults(p, false, lob, q.QuestionID)
				ctx.kickUser(p, false, lob)
			}
		}
	}
	for _, a := range ans {
		if ok := checkAnswer(&a, q); !ok {
			ctx.SaveResults(a.UserID, false, lob, q.QuestionID)
			ctx.kickUser(a.UserID, false, lob)
		}
	}

	if end {
		for _, p := range lob.State.Players {
			ctx.SaveResults(p, true, lob, q.QuestionID)
			ctx.kickUser(p, true, lob)
		}
	}
}

// remove user from lobby, send message, save results
func (ctx *TriviaContext) kickUser(id int64, won bool, lob *Lobby) {
	lob.State.Players = remove(lob.State.Players, id)
	var t string
	if won {
		t = "game-won"
	} else {
		t = "game-over"
	}
	end := TriviaMessage{
		Type:    t,
		UserIDs: []int64{id},
		Lobby:   lob,
	}
	ctx.PublishData(end)
	if id == lob.Creator.ID { // remove lobby after game over
		delete(ctx.Lobbies, lob.LobbyID)
	}
}

// checkPlayer checks if the player identified by id submitted an answer
// to the slice
func checkPlayer(slice []t.Answer, id int64) bool {
	for _, t := range slice {
		if t.UserID == id {
			return true
		}
	}
	return false
}

func checkAnswer(ans *t.Answer, q *t.Question) bool {
	if ans.Answer != q.Answer {
		return false
	}
	return true
}

// remove removes the id from the slice of IDs 's' and returns
// the resulting slice
func remove(s []int64, id int64) []int64 {
	for i := 0; i < len(s); i++ {
		if s[i] == id {
			s = append(s[:i], s[i+1:]...)
			i--
		}
	}
	return s
}

func diffAsInt(diff string) int64 {
	switch diff {
	case "easy":
		return 1
	case "medium":
		return 2
	case "hard":
		return 3
	}
	return 0
}
