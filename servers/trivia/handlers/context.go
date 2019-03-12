package handlers

import (
	"github.com/TriviaRoulette/servers/trivia/models/users"
	mongo "github.com/TriviaRoulette/servers/trivia/mongo"
	"github.com/streadway/amqp"
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

	Channel *amqp.Channel

	Mongo *mongo.MongoStore

	Users users.Store

	Lobbies map[int64]*Lobby

	UserConnections *SocketStore
}
