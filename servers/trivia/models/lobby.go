package models

import ()

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
