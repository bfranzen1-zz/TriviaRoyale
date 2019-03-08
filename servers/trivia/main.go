package main

import (
	"database/sql"
	"fmt"
	"github.com/TriviaRoulette/servers/trivia/handlers"
	"log"
	"net/http"
	"os"
)

//main is the main entry point for the server
func main() {
	addr := os.Getenv("ADDR")
	if len(addr) == 0 {
		addr = ":80"
	}

	// access to user store
	// TODO: connect to context
	dsn := os.Getenv("DSN")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("error opening database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(db)

	// TODO: initialize mongostore
	// TODO: connect to queue
	// TODO: initialize socketstore
	// TODO: initialize lobbies
	// TODO: need context

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/trivia", handlers.PlayerConnectionHandler)

	log.Printf("Server is listening at http:/{container_name}/%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

// need method to get questions from trivia api
// define num of questions we want to get???
