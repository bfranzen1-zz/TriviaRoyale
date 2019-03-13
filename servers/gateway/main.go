package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/TriviaRoulette/servers/gateway/handlers"
	"github.com/TriviaRoulette/servers/gateway/models/users"
	s "github.com/TriviaRoulette/servers/gateway/sessions"
	"github.com/go-redis/redis"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

//main is the main entry point for the server
func main() {
	addr := os.Getenv("ADDR")
	tlscert := os.Getenv("TLSCERT")
	tlskey := os.Getenv("TLSKEY")
	seskey := os.Getenv("SESSIONKEY")
	red := os.Getenv("REDISADDR")
	dsn := os.Getenv("DSN")
	rmq := os.Getenv("RABBITMQ")
	msg := strings.Split(os.Getenv("MESSAGESADDR"), ",")
	// for final project
	triv := os.Getenv("TRIVADDR")

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("error opening database: %v\n", err)
		os.Exit(1)
	}

	trie, err := users.BuildTrie(db)
	if err != nil {
		// just want to warn, don't want to quit
		fmt.Printf("error building trie: %v\n", err)
	}

	hc := handlers.HandlerContext{Key: seskey,
		Session: s.NewRedisStore(redis.NewClient(&redis.Options{Addr: red}), time.Hour),
		Users:   users.NewMySqlStore(db),
		Trie:    trie,
		Sockets: handlers.NewSocketStore(),
	}

	// connect to RabbitMQ
	events, err := hc.Sockets.ConnectQueue(rmq)
	if err != nil {
		log.Fatalf("Error connecting to RabbitMQ, %v", err)
	}

	// start go routine to read/send event/message notifications
	// to sockets
	go hc.Sockets.Read(events)

	if len(addr) == 0 {
		addr = ":443"
	}

	if len(tlscert) == 0 {
		log.Fatal("No TLSCERT variable specified, exiting...")
	}
	if len(tlskey) == 0 {
		log.Fatal("No TLSKEY variable specified, exiting...")
	}

	msgProxy := &httputil.ReverseProxy{Director: CustomDirectorRR(msg, &hc)}
	tUrl, _ := url.Parse(triv)
	trivProxy := &httputil.ReverseProxy{Director: CustomDirector(tUrl, &hc)}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/ws", hc.WebSocketConnectionHandler)
	mux.Handle("/v1/channels", msgProxy)
	mux.Handle("/v1/channels/", msgProxy)
	mux.Handle("/v1/messages/", msgProxy)
	mux.Handle("/v1/trivia", trivProxy)
	mux.Handle("/v1/trivia/", trivProxy)
	mux.HandleFunc("/v1/users", hc.UsersHandler)
	mux.HandleFunc("/v1/users/", hc.SpecificUserHandler)
	mux.HandleFunc("/v1/sessions", hc.SessionsHandler)
	mux.HandleFunc("/v1/sessions/", hc.SpecificSessionHandler)

	wrappedMux := handlers.NewCORS(mux)

	log.Printf("Server is listening on port %s...\n", addr)
	log.Fatal(http.ListenAndServeTLS(addr, tlscert, tlskey, wrappedMux))
}

// Director handles the transport of requests to proper endpoints
type Director func(r *http.Request)

// CustomDirectorRR directs requests for services that have
// multiple servers via Round Robin technique
func CustomDirectorRR(targets []string, hc *handlers.HandlerContext) Director {
	if len(targets) == 1 {
		dest, _ := url.Parse(targets[0])
		return CustomDirector(dest, hc)
	}
	var i int32
	i = 0
	url, _ := url.Parse(targets[int(i)%len(targets)])
	atomic.AddInt32(&i, 1)
	dest := url
	return func(r *http.Request) {
		r.Header.Del("X-User") // remove any previous user
		tmp := handlers.SessionState{}
		_, _ = s.GetState(r, hc.Key, hc.Session, &tmp)
		if tmp.User.ID != 0 { // set if user exists
			j, err := json.Marshal(tmp.User)
			if err != nil {
				return
			}
			r.Header.Set("X-User", string(j))
		}
		r.Header.Add("X-Forwarded-Host", r.Host)
		r.URL.Scheme = "http"
		r.URL.Host = dest.String()
		r.Host = dest.String()
	}
}

// CustomDirector directs requests to a specified server and modifies the request
// before being passed along
func CustomDirector(target *url.URL, hc *handlers.HandlerContext) Director {
	return func(r *http.Request) {
		r.Header.Del("X-User") // remove any previous user
		tmp := handlers.SessionState{}
		_, _ = s.GetState(r, hc.Key, hc.Session, &tmp)
		if tmp.User.ID != 0 { // set if user exists
			j, err := json.Marshal(tmp.User)
			if err != nil {
				return
			}
			r.Header.Set("X-User", string(j))
		}
		r.Header.Add("X-Forwarded-Host", r.Host)
		r.URL.Scheme = "http"
		r.URL.Host = target.String()
		r.Host = target.String()
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
