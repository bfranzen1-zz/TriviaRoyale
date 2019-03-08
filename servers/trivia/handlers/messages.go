package handlers

import (
	"github.com/TriviaRoulette/servers/trivia/models"
	"net/http"
)

/*
GET /v1/trivia
	- get all lobbies
	- upgrade user to websocket
	- returns json encoded slice of type lobby
POST /v1/trivia
	- make a new lobby
		- add new map entry to context lobbies (make sure we mutex lock)
		- make sure options are correct
	- type new-lobby queue message
		- lobby struct (contains gamestate, options, id)
	- use passed lobby struct to get questions from api
*/
func (ctx *TriviaContext) LobbyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-User") == "" {
		http.Error(w, "Unauthorized Access", 401)
	}

	if r.Method == "GET" {
	} else if r.Method == "POST" {

	} else {
		http.Error(w, "Method Not Allowed", 405)
	}
}

/*
POST /v1/trivia/<lobby_id>
	- user joins lobby
	- add user to lobby connections
	- check num users < max
	- type join-lobby
		- lobby struct
*/
func (ctx *TriviaContext) SpecificLobbyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-User") == "" {
		http.Error(w, "Unauthorized Access", 401)
	}

	if r.Method == "POST" {

	} else {
		http.Error(w, "Method Not Allowed", 405)
	}
}

func (ctx *TriviaContext) getTriviaQuestions(opt *models.Options) {
	// TODO: get api data using passed in options.
	// update lobby in context with gamestate
}
