package handlers

// NotifyContext is a receiver that stores
// information about web sockets
type NotifyContext struct {

	// stores open web socket connections
	Sockets *SocketStore
}
