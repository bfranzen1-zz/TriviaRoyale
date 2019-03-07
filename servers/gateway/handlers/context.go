package handlers

import (
	"github.com/TriviaRoulette/servers/gateway/indexes"
	"github.com/TriviaRoulette/servers/gateway/models/users"
	"github.com/TriviaRoulette/servers/gateway/sessions"
)

// HandlerContext is a receiver that stores
// information to verify and sign SessionIDs
// as well as the location for session storage, user storage,
// user search data, and web sockets
type HandlerContext struct {
	// used for signing/verifying SessionID
	Key string

	// stores information in memory for session
	Session sessions.Store

	// stores information on the server about users
	Users users.Store

	// stores searchable references to users
	Trie *indexes.Trie

	// stores open web socket connections
	Sockets *SocketStore
}
