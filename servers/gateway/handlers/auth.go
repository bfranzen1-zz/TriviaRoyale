package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/bfranzen1/assignments-bfranzen1/servers/gateway/indexes"
	u "github.com/bfranzen1/assignments-bfranzen1/servers/gateway/models/users"
	s "github.com/bfranzen1/assignments-bfranzen1/servers/gateway/sessions"
	"github.com/mitchellh/mapstructure"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// UsersHandler handles requests for the user resource and allows POST requests for creating
// a new user account
func (hc *HandlerContext) UsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" { // POST
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") { // not application/json content type
			http.Error(w, "Request Body must be JSON", 415)
			return
		}

		// get JSON body from request
		tmp, err := getJSON(r, w)
		if err != nil {
			fmt.Print(err)
			return
		}
		var nu u.NewUser
		mapstructure.Decode(tmp, &nu) // put json body into new user struct

		usr, err := nu.ToUser() // convert to user
		if err != nil {         // invalid user struct
			fmt.Printf("error converting to User: %v", err)
			w.WriteHeader(400)
			return
		}

		if _, err := hc.Users.Insert(usr); err != nil { // insert new user into RDBMS
			fmt.Printf("error inserting user: %v", err)
			w.WriteHeader(400)
			return
		}
		// add user to Trie
		hc.Trie = addField(usr.UserName, hc.Trie, usr.ID)
		hc.Trie = addField(usr.FirstName, hc.Trie, usr.ID)
		hc.Trie = addField(usr.LastName, hc.Trie, usr.ID)

		// begin a new session with user
		_, err = s.BeginSession(hc.Key, hc.Session, SessionState{SessionStart: time.Now(), User: *usr}, w)
		if err != nil {
			fmt.Printf("Error starting session: %v", err)
			return
		}
		// set proper headers
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		enc := json.NewEncoder(w) // send response to client
		if err = enc.Encode(*usr); err != nil {
			fmt.Printf("Error encoding to JSON: %v", err)
			return
		}

	} else if r.Method == "GET" {
		hc.SearchHandler(w, r)
	} else {
		http.Error(w, "Method Not Allowed", 405)
	}
}

// SpecificUserHandler handles requests for a specific users including getting a User account
// and updating a user account
func (hc *HandlerContext) SpecificUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "PATCH" { // only supported request methods
		http.Error(w, "Method Not Allowed", 405)
		return
	}

	var sess SessionState // gets/updates session for user
	if _, err := s.GetState(r, hc.Key, hc.Session, &sess); err != nil {
		http.Error(w, "Not authorized to access resource", 401)
		return
	}

	id := r.URL.Path[10:] // id of user can be "me" or number
	var uid int64         // user id
	if id == "me" {
		uid = sess.User.ID
	} else {
		uid, _ = strconv.ParseInt(id, 10, 64)
	}
	if r.Method == "GET" {
		usr, err := hc.Users.GetByID(uid) // user associated with parsed id
		if err != nil {                   // invalid id
			http.Error(w, "Resource not found", 404)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		// return user info
		enc := json.NewEncoder(w)
		if err = enc.Encode(usr); err != nil {
			fmt.Printf("Error encoding to JSON: %v", err)
			return
		}

	} else if r.Method == "PATCH" {
		if id != "me" && (uid != sess.User.ID) { // not authorized to update/access user
			http.Error(w, "Unauthorized access to resource", 403)
			return
		}

		// not application/json content type
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			http.Error(w, "Request Body must be JSON", 415)
			return
		}

		tmp, err := getJSON(r, w)
		if err != nil {
			fmt.Print(err)
			return
		}
		var upd u.Updates
		mapstructure.Decode(tmp, &upd)

		// remove old first/last name
		hc.Trie.Remove(strings.ToLower(sess.User.FirstName), sess.User.ID)
		hc.Trie.Remove(strings.ToLower(sess.User.LastName), sess.User.ID)

		if err = sess.User.ApplyUpdates(&upd); err != nil {
			fmt.Printf("Error applying updates: %v", err)
		}

		// add new first/last name, if error occurs original first/last is added back
		hc.Trie = addField(upd.FirstName, hc.Trie, sess.User.ID)
		hc.Trie = addField(upd.LastName, hc.Trie, sess.User.ID)

		if _, err = hc.Users.Update(sess.User.ID, &upd); err != nil {
			fmt.Printf("Error updating user in User Store: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		enc := json.NewEncoder(w)
		if err = enc.Encode(sess.User); err != nil {
			fmt.Printf("Error encoding to JSON: %v", err)
			return
		}
	}
}

// SessionsHandler handles generating new sessions when valid credentials are passed
// errors are returned if the user passes incorrect credentials
func (hc *HandlerContext) SessionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// not application/json content type
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			http.Error(w, "Request Body must be JSON", 415)
			return
		}

		tmp, err := getJSON(r, w)
		if err != nil {
			fmt.Print(err)
			return
		}
		var cred u.Credentials
		mapstructure.Decode(tmp, &cred)

		usr, err := hc.Users.GetByEmail(cred.Email)
		if err != nil {
			time.Sleep(800 * time.Millisecond)
			http.Error(w, "Invalid Credentials", 401)
			return
		}

		if err = usr.Authenticate(cred.Password); err != nil {
			time.Sleep(800 * time.Millisecond)
			http.Error(w, "Invalid Credentials", 401)
			return
		}

		_ = logSession(usr, r, hc)

		_, err = s.BeginSession(hc.Key, hc.Session, SessionState{SessionStart: time.Now(), User: *usr}, w)
		if err != nil {
			fmt.Printf("Error starting session: %v", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		enc := json.NewEncoder(w)
		if err = enc.Encode(usr); err != nil {
			fmt.Printf("Error encoding to JSON: %v", err)
			return
		}
	} else {
		http.Error(w, "Method Not Allowed", 405)
	}
}

// SpecificSessionHandler handles when an authenticated user wants to end a session
func (hc *HandlerContext) SpecificSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "DELETE" {
		lastPath := r.URL.Path[13:]
		if lastPath != "mine" {
			http.Error(w, "Invalid request", 403)
		}
		// **TODO**: DELETE USER FROM CONNECTIONS
		tmp := SessionState{}
		_, _ = s.GetState(r, hc.Key, hc.Session, &tmp)
		if tmp.User.ID != 0 {
			delete(hc.Sockets.Connections, tmp.User.ID)
		}
		if _, err := s.EndSession(r, hc.Key, hc.Session); err != nil {
			fmt.Printf("Error ending session: %v", err)
			return
		}
		w.Write([]byte("signed out"))
	} else {
		http.Error(w, "Method Not Allowed", 405)
	}
}

// SearchHandler handles when a user is searching for other users
func (hc *HandlerContext) SearchHandler(w http.ResponseWriter, r *http.Request) {
	var sess SessionState // gets/updates session for user
	if _, err := s.GetState(r, hc.Key, hc.Session, &sess); err != nil {
		http.Error(w, "Not authorized to access resource", 401)
		return
	}

	q := r.URL.RawQuery
	if string(q[0]) != "q" {
		http.Error(w, "Bad Request", 400)
	}

	ids := hc.Trie.Find(strings.ToLower(string(q[2:])), 20)
	var users []*u.User
	for _, id := range ids {
		tmp, err := hc.Users.GetByID(id)
		if err != nil {
			fmt.Printf("Error getting user, %v", err)
			return
		}
		users = append(users, tmp)
	}
	sort.Slice(users, func(i, j int) bool {
		return users[i].UserName < users[j].UserName
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	enc := json.NewEncoder(w)
	if err := enc.Encode(users); err != nil {
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

// logSession adds a record to the RDBMS signin table to keep track of sign
// in attempts corresponding to the user requesting access. It takes in the user,
// the request they sent, and a HandlerContext to insert into the appropriate database.
// An error is returned if the insert fails.
func logSession(usr *u.User, r *http.Request, hc *HandlerContext) error {
	ips := r.Header.Get("X-Forwarded-For")
	var ip string
	if len(ips) != 0 {
		ip = strings.Split(ips, ", ")[0]
	} else {
		ip = r.RemoteAddr
	}

	if err := hc.Users.InsertSignIn(usr.ID, time.Now().Format(time.RFC3339), ip); err != nil {
		return err
	}
	return nil
}

// addField takes in a field string, trie to add to, and the id of the field
// and adds it to the trie. If the field contains multiple words they are added
// as individual fields
func addField(field string, t *indexes.Trie, id int64) *indexes.Trie {
	vals := strings.Split(field, " ")
	if len(vals) > 1 {
		for _, s := range vals {
			t.Add(strings.ToLower(s), id)
		}
	} else {
		t.Add(strings.ToLower(field), id)
	}
	return t
}
