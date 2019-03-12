package handlers

import (
	"encoding/json"
	"fmt"
	t "github.com/TriviaRoulette/servers/trivia/models/trivia"
	"github.com/TriviaRoulette/servers/trivia/models/users"
	"github.com/mitchellh/mapstructure"
	"github.com/streadway/amqp"
	"gopkg.in/mgo.v2/bson"
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
	Type     string     `json:"type"`
	Lobby    *Lobby     `json:"lobby,omitempty"`
	Options  Options    `json:"options,omitempty"`
	Question t.Question `json:"question,omitempty"`
	UserIDs  []int64    `json:"userIDs,omitempty"`
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
		lob := &Lobby{
			MongoId: bson.NewObjectId(),
			LobbyId: j["lobbyID"].(int64),
			Options: &opt,
			State:   getData(opt),
			Creator: &player,
			Over:    false,
		}
		if err := ctx.Mongo.Insert(lob, "game"); err != nil {
			fmt.Println("error inserting record, %v", err)
		}
		ctx.Lobbies[j["lobbyID"].(int64)] = lob
		e := TriviaMessage{
			Type:  "lobby-new",
			Lobby: lob,
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

	if r.Method == "GET" { // start game
		if val, ok := ctx.Lobbies[player.ID]; ok { // creator has lobby and is creator
			go ctx.StartGame(val)
		}

	} else if r.Method == "POST" { // add user
		reqType := r.URL.Query().Get("type")
		j, _ := getJSON(r, w)
		if reqType == "add" { // user asking to join lobby
			lob := ctx.Lobbies[j["lobbyID"].(int64)]
			lob.State.Players = append(lob.State.Players, player.ID)
			if err := ctx.Mongo.Update(lob.MongoId, "game", bson.M{"$set": bson.M{"state": lob.State}}); err != nil {
				fmt.Println("error updating record, %v", err)
			}
			e := TriviaMessage{
				Type:    "lobby-add",
				Lobby:   lob,
				UserIDs: lob.State.Players,
			}
			ctx.PublishData(e)
			w.Write([]byte("user added to lobby"))
		}

		if reqType == "answer" { // client answers question
			j, err := ioutil.ReadAll(r.Body)
			if err != nil {
				fmt.Printf("Error reading request body: %v", err)
			}
			var dest t.Answer
			if err = json.Unmarshal(j, &dest); err != nil {
				http.Error(w, "Invalid JSON Syntax", 400)
				fmt.Println("Invalid JSON Syntax")
			}
			ans := ctx.Lobbies[dest.LobbyID].State.Answers[dest.QuestionID]
			ans = append(ans, dest)
			lob := ctx.Lobbies[dest.LobbyID]
			if err := ctx.Mongo.Update(lob.MongoId, "game", bson.M{"$set": bson.M{"state": lob.State}}); err != nil {
				fmt.Println("error updating record, %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			w.Write([]byte("answer received"))
		}
	} else {
		http.Error(w, "Method Not Allowed", 405)
	}
}

/*
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
		Players:   []int64{},
		Answers:   map[int64][]t.Answer{},
		Questions: []t.Question{},
	}
	var res map[string]interface{}
	if err := json.Unmarshal(data, &res); err != nil {
		fmt.Println("error unmarshaling json")
	}
	arr := res["results"].([]interface{})
	for i, q := range arr {
		nxt := t.Question{}
		row := q.(map[string]interface{})
		nxt.QuestionID = int64(i + 1) // start at 1
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
