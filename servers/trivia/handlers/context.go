package handlers

import (
	"github.com/TriviaRoulette/servers/trivia/models/users"
	mongo "github.com/TriviaRoulette/servers/trivia/mongo"
	"github.com/streadway/amqp"
	//"github.com/gorilla/websocket"
	//"sync"
)

type TriviaContext struct {
	// channel for publishing messages
	Channel *amqp.Channel

	// mongo store for saving game/user stats
	Mongo *mongo.MongoStore

	// access to methods from gateway for basic
	Users users.Store

	// map from creator id to lobby struct for game logic/state
	Lobbies map[int64]*Lobby
}
