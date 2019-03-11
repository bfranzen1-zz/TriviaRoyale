package handlers

import (
	"encoding/json"
	"fmt"
	t "github.com/TriviaRoulette/servers/trivia/models/trivia"
	"github.com/TriviaRoulette/servers/trivia/models/users"
	"github.com/mitchellh/mapstructure"
	"github.com/streadway/amqp"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	baseURL      = "https://opentdb.com/api.php?"
	numQuestions = "amount"
	cat          = "category"
	diff         = "difficulty"
	qtype        = "type"
)

// TriviaMessage is a struct that holds
// information about the parts of the trivia service
type TriviaMessage struct {
	Type    string  `json:"type"`
	LobbyID int64   `json:"lobbyID,omitempty"`
	Lobby   Lobby   `json:"lobby,omitempty"`
	Options Options `json:"options,omitempty"`
}

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

/*
POST /v1/trivia/<lobby_id>
	- user joins lobby
	- add user to lobby connections
	- check num users < max
	- type join-lobby
		- lobby struct
*/

// TriviaHandler handles when a client sends messages to the rabbitmq
// that pertains to the trivia microservice
func (ctx *TriviaContext) TriviaHandler(data []byte) {
	event := TriviaMessage{}
	if err := json.Unmarshal(data, &event); err != nil {
		fmt.Printf("error getting message body, %v", err)
	}

	switch event.Type {
	case "lobby-new":
		ctx.UserConnections.WriteToValidConnections([]int64{}, TextMessage, data)
	case "lobby-add":

	case "lobby-start":

	case "game-answer":

	}
}

// LobbyHandler handles when the client creates a new lobby for
// a trivia game
func (ctx *TriviaContext) LobbyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-User") == "" {
		http.Error(w, "Unauthorized Access", 401)
	}

	player := users.User{}
	if err := json.Unmarshal([]byte(r.Header.Get("X-User")), &player); err != nil {
		fmt.Printf("error getting message body, %v", err)
	}

	if r.Method == "POST" {
		j, err := getJSON(r, w)
		if err != nil {
			http.Error(w, "Bad Request", 400)
		}
		var opt Options
		mapstructure.Decode(j["options"], &opt)
		lob := Lobby{
			LobbyId: j["lobbyID"].(int64),
			Options: &opt,
			State:   getData(opt),
		}
		ctx.Lobbies[j["lobbyID"].(int64)] = &lob
		e := TriviaMessage{
			Type:    "lobby-new",
			Lobby:   lob,
			LobbyID: lob.LobbyId,
		}

		ctx.PublishData(e)
		w.Write([]byte("lobby created"))
	} else {
		http.Error(w, "Method Not Allowed", 405)
	}
}

// SpecificLobbyHandler handles when the client sends a request pertaining to a specific lobby
func (ctx *TriviaContext) SpecificLobbyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-User") == "" {
		http.Error(w, "Unauthorized Access", 401)
	}

	player := users.User{}
	if err := json.Unmarshal([]byte(r.Header.Get("X-User")), &player); err != nil {
		fmt.Printf("error getting message body, %v", err)
	}

	if r.Method == "POST" {

	} else {
		http.Error(w, "Method Not Allowed", 405)
	}
}

/*					******MESSAGES******

 -- FROM CLIENT ->
			lobby-new (get unique id from user, lobby options)
				-- add to lobby slice in ctx
				-- get api data from open trivia

			lobby-add
				-- add user to lobby players

			lobby-start
				-- send first question

			game-answer
				-- player answers question

 -- FROM SERVER ->

			game-question


			game-over

			-- player lost, remove connection

			game-end

			-- game over, player made tied/won
				-- send points


*/

// getJSON takes in an http request, destination interface, and response writer
// to unmarshal and store the request body into the destination and write any errors
// to the response writer and will return other errors to the console
func getJSON(r *http.Request, w http.ResponseWriter) (map[string]interface{}, error) {
	var dest map[string]interface{}
	j, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading request body: %v", err)
	}
	if err = json.Unmarshal([]byte(j), &dest); err != nil {
		http.Error(w, "Invalid JSON Syntax", 400)
		return nil, fmt.Errorf("Invalid JSON Syntax")
	}
	return dest, nil
}

// getData queries the open trivia api using the passed options
// and returns a game state object
func getData(opt Options) *GameState {
	u, _ := url.Parse(baseURL)
	q := u.Query()
	if len(opt.Category) > 0 {
		q.Set(cat, opt.Category)
	}
	if len(opt.Difficulty) > 0 {
		q.Set(diff, opt.Difficulty)
	}
	if opt.NumQuestions != 0 {
		q.Set(numQuestions, string(opt.NumQuestions))
	}
	u.RawQuery = q.Encode()

	resp, _ := http.Get(u.String())
	body, _ := ioutil.ReadAll(resp.Body)
	return formatState(body)
}

// formatState takes in a byte array representing data from the
// open trivia api and uses that data to build a GameState object
// and returns that object
func formatState(data []byte) *GameState {
	state := GameState{
		Players: &SocketStore{},
		Answers: map[int64]*t.Answer{},
	}
	var res map[string]interface{}
	if err := json.Unmarshal(data, &res); err != nil {
		fmt.Println("error unmarshaling json")
	}
	arr := res["results"].([]interface{})
	for i, q := range arr {
		nxt := t.Question{}
		row := q.(map[string]interface{})
		nxt.QuestionNum = int64(i)
		nxt.Question = row["question"].(string)
		nxt.Choices = row["incorrect_answers"].([]string)
		nxt.Answer = row["correct_answer"].(string)
		state.Questions = append(state.Questions, nxt)
	}
	return &state
}

func (ctx *TriviaContext) PublishData(data interface{}) {
	body, _ := json.Marshal(data)

	queue, err := ctx.Channel.QueueDeclare(qName, true, false, false, false, nil)
	if err != nil {
		fmt.Errorf("error declaring queue, %v", err)
	}

	err = ctx.Channel.Publish(
		"",
		queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(body),
		})
	if err != nil {
		fmt.Errorf("error publish to queue, %v", err)
	}
}
