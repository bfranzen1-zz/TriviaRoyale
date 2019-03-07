package handlers

import (
	u "github.com/TriviaRoulette/servers/gateway/models/users"
	s "github.com/TriviaRoulette/servers/gateway/sessions"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestServeHTTP(t *testing.T) {
	// create mock db
	db, _, err := sqlmock.New()
	if err != nil {
		log.Fatalf("an error occurred while opening mock connection: %s", err)
	}
	defer db.Close()

	// make handler context with session store, session key for signing, and user db
	hc := HandlerContext{Key: "test Key", Session: s.NewMemStore(time.Hour, time.Minute), Users: u.NewMySqlStore(db)}
	req, err := http.NewRequest("GET", "/v1/users", nil) // dummy request
	rr := httptest.NewRecorder()                         // records responses written

	cors := NewCORS(http.HandlerFunc(hc.UsersHandler)) // make CORS middleware
	// generate anonymous function to run CORS middleware
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cors.ServeHTTP(w, r)
	})

	handler.ServeHTTP(rr, req)
	if rr.Header().Get("Access-Control-Allow-Origin") != "*" ||
		rr.Header().Get("Access-Control-Allow-Methods") != "GET, PUT, POST, PATCH, DELETE" ||
		rr.Header().Get("Access-Control-Allow-Headers") != "Content-Type, Authorization" ||
		rr.Header().Get("Access-Control-Expose-Headers") != "Authorization" ||
		rr.Header().Get("Access-Control-Max-Age") != "600" {
		t.Error("Not all pre-flight headers were added, check the cors middleware file")
	}
}
