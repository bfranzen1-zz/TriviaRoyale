package handlers

import (
	"github.com/gorilla/websocket"
	"sync"
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

	UserConnections *SocketStore
}

type SocketStore struct {
	Connections map[int64]*websocket.Conn
	lock        sync.Mutex
}
