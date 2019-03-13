package main

import (
	"database/sql"
	"fmt"
	"github.com/TriviaRoulette/servers/trivia/handlers"
	u "github.com/TriviaRoulette/servers/trivia/models/users"
	m "github.com/TriviaRoulette/servers/trivia/mongo"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net/http"
	"os"
)

//main is the main entry point for the server
func main() {
	addr := os.Getenv("ADDR")
	if len(addr) == 0 {
		addr = ":8000"
	}

	// access to user store
	dsn := os.Getenv("DSN")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("error opening database: %v\n", err)
		os.Exit(1)
	}

	// store game/results
	mg := os.Getenv("MONGO_ADDR")
	sess, err := mgo.Dial(mg)
	if err != nil {
		log.Fatalf("error dialing mongo: %v", err)
	}

	ch, err := handlers.ConnectQueue(os.Getenv("RABBITMQ"))
	if err != nil {
		log.Fatalf("error connecting to queue, %v", err)
	}
	// TODO: initialize lobbies
	// TODO: need context
	ctx := handlers.TriviaContext{
		Mongo:   m.NewMongoStore(sess),
		Users:   u.NewMySqlStore(db),
		Lobbies: map[bson.ObjectId]*handlers.Lobby{},
		Channel: ch,
	}
	// connect/add queue to context
	//ctx.ConnectQueue(os.Getenv("RABBITADDR"))

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/trivia", ctx.LobbyHandler)
	mux.HandleFunc("/v1/trivia/", ctx.SpecificLobbyHandler)

	log.Printf("Server is listening at http:/trivia/%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
