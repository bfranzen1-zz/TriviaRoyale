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
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
)

// constants for api request to Open Trivia
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
	Type     string        `json:"type"`
	LobbyID  bson.ObjectId `json:"lobbyID,omitempty"`
	Lobby    *Lobby        `json:"lobby,omitempty"`
	Options  Options       `json:"options,omitempty"`
	Question t.Question    `json:"question,omitempty"`
	UserIDs  []int64       `json:"userIDs,omitempty"`
}

// LobbyHandler handles when the client creates a new lobby for
// a trivia game
func (ctx *TriviaContext) LobbyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-User") == "" {
		http.Error(w, "Unauthorized Access", 401)
		return
	}

	player := users.User{}
	if err := json.Unmarshal([]byte(r.Header.Get("X-User")), &player); err != nil {
		fmt.Printf("error getting message body, %v", err)
		http.Error(w, "Bad Request", 400)
		return
	}

	if r.Method == "GET" { // user goes to lobby page
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		enc := json.NewEncoder(w)
		if err := enc.Encode(ctx.Lobbies); err != nil {
			fmt.Printf("Error encoding to JSON: %v", err)
			return
		}
	} else if r.Method == "POST" { // new lobby
		j, err := getJSON(r, w)
		if err != nil {
			http.Error(w, "Bad Request", 400)
			return
		}
		var opt Options
		mapstructure.Decode(j, &opt)
		lob := &Lobby{
			LobbyID:    bson.NewObjectId(),
			Options:    &opt,
			State:      getData(&opt),
			Creator:    &player,
			Over:       false,
			InProgress: false,
		}
		lob.lock.Lock()
		lob.State.Players = append(lob.State.Players, player.ID)
		lob.lock.Unlock()
		if !checkOptions(lob.Options) { // ensure valid options
			http.Error(w, "Bad Request", 400)
			return
		}

		if err := ctx.Mongo.Insert(lob, "game"); err != nil {
			fmt.Printf("error inserting record, %v", err)
		}
		ctx.Lobbies[lob.LobbyID] = lob
		e := TriviaMessage{
			Type:    "lobby-new",
			Lobby:   lob,
			UserIDs: []int64{},
		}

		ctx.PublishData(e)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		enc := json.NewEncoder(w)
		if err := enc.Encode(lob); err != nil {
			fmt.Printf("Error encoding to JSON: %v", err)
			return
		}
	} else {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
}

// SpecificLobbyHandler handles when the client sends a request pertaining to a specific lobby
func (ctx *TriviaContext) SpecificLobbyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-User") == "" {
		http.Error(w, "Unauthorized Access", 401)
		return
	}

	player := users.User{}
	if err := json.Unmarshal([]byte(r.Header.Get("X-User")), &player); err != nil {
		http.Error(w, "Bad Request", 400)
		return
	}
	//lobby id
	lid := r.URL.Path[11:]

	if r.Method == "GET" { // start game
		if val, ok := ctx.Lobbies[bson.ObjectIdHex(lid)]; ok { // creator has lobby and is creator
			if val.Creator.ID == player.ID {
				go ctx.StartGame(val)
			} else { // not creator, can't start game
				http.Error(w, "Unauthorized Access", 401)
				return
			}
		}
		w.WriteHeader(200)
	} else if r.Method == "POST" { // add user
		reqType := r.URL.Query().Get("type")
		if reqType == "add" { // user asking to join lobby
			lob := ctx.Lobbies[bson.ObjectIdHex(lid)]
			if len(lob.State.Players) < int(lob.Options.MaxPlayers) {
				lob.lock.Lock()
				lob.State.Players = append(lob.State.Players, player.ID)
				lob.lock.Unlock()
				// got all the players we want
				if len(lob.State.Players) == int(lob.Options.MaxPlayers) {
					go ctx.StartGame(lob)
				}
			} else { // max players reached
				http.Error(w, "Bad Request", 400)
				return
			}
			if err := ctx.Mongo.Update(lob.LobbyID, "game", bson.M{"$set": bson.M{"state": lob.State}}); err != nil {
				fmt.Printf("error updating record, %v", err)
			}
			e := TriviaMessage{
				Type:    "lobby-add",
				Lobby:   lob,
				UserIDs: lob.State.Players,
			}
			ctx.PublishData(e)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(201)
			enc := json.NewEncoder(w)
			if err := enc.Encode(lob); err != nil {
				fmt.Printf("Error encoding to JSON: %v", err)
				return
			}
		}

		if reqType == "answer" { // client answers question
			j, err := ioutil.ReadAll(r.Body)
			if err != nil {
				fmt.Printf("Error reading request body: %v", err)
			}
			var dest t.Answer
			if err = json.Unmarshal(j, &dest); err != nil {
				http.Error(w, "Bad Request", 400)
				return
			}
			if bson.ObjectIdHex(lid) != dest.LobbyID {
				fmt.Printf("format error, request id for lobby was %s, answer contained %s", lid, dest.LobbyID)
			}
			if !ctx.Lobbies[dest.LobbyID].InProgress { // game hasn't started yet
				http.Error(w, "Bad Request", 400)
				return
			}
			lob := ctx.Lobbies[dest.LobbyID]
			lob.lock.Lock()
			lob.State.Answers[dest.QuestionID] = append(lob.State.Answers[dest.QuestionID], dest)
			lob.lock.Unlock()
			if err := ctx.Mongo.Update(lob.LobbyID, "game", bson.M{"$set": bson.M{"state": lob.State}}); err != nil {
				fmt.Printf("error updating record, %v", err)
			}
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(200)
			w.Write([]byte("answer received"))
		}
	} else if r.Method == "PATCH" { // update options
		j, err := getJSON(r, w)
		if err != nil {
			http.Error(w, "Bad Request", 400)
			return
		}
		// player is not creator
		if ctx.Lobbies[bson.ObjectIdHex(lid)].Creator.ID != player.ID {
			http.Error(w, "Unauthorized Access", 401)
			return
		}

		var opt Options
		mapstructure.Decode(j, &opt)

		if !checkOptions(&opt) { // ensure valid options
			http.Error(w, "Bad Request", 400)
			return
		}

		ctx.Lobbies[bson.ObjectIdHex(lid)].Options = &opt
		ctx.Lobbies[bson.ObjectIdHex(lid)].State = getData(&opt)
		e := TriviaMessage{
			Type:    "lobby-update",
			Lobby:   ctx.Lobbies[bson.ObjectIdHex(lid)],
			UserIDs: []int64{},
		}
		ctx.PublishData(e)
		// update makes max players == num players in lobby, need to start game
		if int(opt.MaxPlayers) == len(ctx.Lobbies[bson.ObjectIdHex(lid)].State.Players) {
			go ctx.StartGame(ctx.Lobbies[bson.ObjectIdHex(lid)])
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		enc := json.NewEncoder(w)
		if err := enc.Encode(ctx.Lobbies[bson.ObjectIdHex(lid)]); err != nil {
			fmt.Printf("Error encoding to JSON: %v", err)
			return
		}
	} else {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
}

// StatisticsHandler handles when a client requests user statistics
func (ctx *TriviaContext) StatisticsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-User") == "" {
		http.Error(w, "Unauthorized Access", 401)
		return
	}

	player := users.User{}
	if err := json.Unmarshal([]byte(r.Header.Get("X-User")), &player); err != nil {
		http.Error(w, "Bad Request", 400)
		return
	}
	var dest []UserStatistic
	if err := ctx.Mongo.GetUserStat(player.ID, dest); err != nil {
		// no rows in mongo
		http.Error(w, "Bad Request", 400)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	enc := json.NewEncoder(w)
	if err := enc.Encode(dest); err != nil {
		fmt.Printf("Error encoding to JSON: %v", err)
		return
	}
}

// getJSON takes in an http request, destination interface, and response writer
// to unmarshal and store the request body into the destination and write any errors
// to the response writer and will return other errors to the console
func getJSON(r *http.Request, w http.ResponseWriter) (interface{}, error) {
	var dest interface{}
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
func getData(opt *Options) *GameState {
	u, _ := url.Parse(baseURL)
	q := u.Query()
	// don't need to check for valid options here
	q.Set(cat, strconv.FormatInt(opt.Category, 10))
	q.Set(diff, opt.Difficulty)
	q.Set(numQuestions, strconv.FormatInt(opt.NumQuestions, 10))
	u.RawQuery = q.Encode()
	resp, err := http.Get(u.String())
	if err != nil {
		fmt.Printf("Error getting Trivia data, %v", err)
		return nil
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body, %v", err)
		return nil
	}
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
		s := make([]string, len(row["incorrect_answers"].([]interface{})))
		for i, v := range row["incorrect_answers"].([]interface{}) {
			s[i] = fmt.Sprint(v)
		}
		nxt.Choices = s
		nxt.Choices = append(nxt.Choices, row["correct_answer"].(string))
		rand.Shuffle(len(nxt.Choices), func(i, j int) { // shuffle choices, so answer not obvious
			nxt.Choices[i], nxt.Choices[j] = nxt.Choices[j], nxt.Choices[i]
		})
		nxt.Answer = row["correct_answer"].(string)
		state.Questions = append(state.Questions, nxt)
	}
	return &state
}

// PublishData takes the input data and publishes it to rabbitmq
// for consumers to parse and send to clients
func (ctx *TriviaContext) PublishData(data interface{}) {
	body, _ := json.Marshal(data)

	queue, err := ctx.Channel.QueueDeclare(qName, true, false, false, false, nil)
	if err != nil {
		fmt.Printf("error declaring queue, %v", err)
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
		fmt.Printf("error publish to queue, %v", err)
	}
}

// checkOptions ensures the passed opt options are valid and returns a boolean
// representing such
func checkOptions(opt *Options) bool {
	if opt.Category < 9 {
		fmt.Println("invalid category, less than 9")
		return false
	}
	if opt.Difficulty == "" {
		fmt.Println("need category")
		return false
	}
	if opt.MaxPlayers < 2 {
		fmt.Println("need more than 1 player")
		return false
	}
	if opt.NumQuestions < 1 {
		fmt.Println("need at least one question")
		return false
	}
	return true
}
