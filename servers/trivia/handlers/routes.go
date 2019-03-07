package handlers

/*
GET /v1/trivia
	- get all lobbies
	- upgrade user to websocket
	- returns json encoded slice of type lobby

*/

/*
POST /v1/trivia
	- make a new lobby
		- add new map entry to context lobbies (make sure we mutex lock)
		- make sure options are correct
	- type new-lobby queue message
		- lobby struct (contains gamestate, options, id)
*/

/*
POST /v1/trivia/<lobby_id>
	- user joins lobby
	- add user to lobby connections
	- check num users < max
	- type join-lobby
		- lobby struct
*/
