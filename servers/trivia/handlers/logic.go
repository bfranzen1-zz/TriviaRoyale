package handlers

import (
	t "github.com/TriviaRoulette/servers/trivia/models/trivia"
)

type GameState struct {
	Players   *SocketStore
	Questions []t.Question
	Answers   map[int64]*t.Answer
}

type Lobby struct {
	LobbyId int64
	State   *GameState
	Options *Options
}

type Options struct {
	NumQuestions int64
	Category     string
	Difficulty   string
	MaxPlayers   int64
}

// TODO: Need queue logic here

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

func StartGame() {
	// TODO: start go routine for specific lobby questions
}

func SaveResults() {
	// TODO: save results of game in mongo store
	// remove lobby from context
}
