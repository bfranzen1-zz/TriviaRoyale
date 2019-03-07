package main

import (
	"log"
	"net/http"
	"os"

	"github.com/bfranzen1/assignments-bfranzen1/servers/summary/handlers"
)

//main is the main entry point for the server
func main() {
	addr := os.Getenv("ADDR")
	if len(addr) == 0 {
		addr = ":80"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/summary", handlers.SummaryHandler)

	log.Printf("Server is listening at http:/{container_name}/%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
