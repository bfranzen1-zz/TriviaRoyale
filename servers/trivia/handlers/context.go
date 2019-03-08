package handlers

import (
	"github.com/TriviaRoulette/servers/gateway/models/users"
	"github.com/TriviaRoulette/servers/trivia/db"
	"github.com/TriviaRoulette/servers/trivia/models"
	//"github.com/gorilla/websocket"
	//"sync"
)

type TriviaContext struct {
	/*
		NEED:
		-	queue to publish/consume from

		-	mongo store

		- 	user db connection

		-	lobbies (each one gets subset of user connections)

		- all user connections
	*/

	Mongo db.MongoStore

	Users users.Store

	Lobbies []*models.Lobby

	UserConnections *SocketStore
}
