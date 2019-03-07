package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	u "github.com/bfranzen1/assignments-bfranzen1/servers/gateway/models/users"
	s "github.com/bfranzen1/assignments-bfranzen1/servers/gateway/sessions"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"
)

const INS = "insert into users(email,pass_hash,user_name,first_name,last_name,photo_url) values(?,?,?,?,?,?)"
const GETID = "select * from users where id = ?"
const GETEMQUERY = "select * from users where email = ?"
const UPD = "update users set first_name = ?, last_name = ? where id = ?"
const gravatarBasePhotoURL = "https://www.gravatar.com/avatar/"

var COLS = []string{"id", "email", "pass_hash", "user_name", "first_name", "last_name", "photo_url"}

// COLS expected for BuildTrie
var trieCOLS = []string{"id", "user_name", "first_name", "last_name"}

// TestUserHandler tests the functionality of the UsersHandler function
func TestUsersHandler(t *testing.T) {
	cases := []struct {
		name     string
		method   string
		params   map[string]string
		headers  map[string]string
		expected int
	}{
		{
			"Wrong Request Method",
			"PATCH",
			nil,
			nil,
			405,
		},
		{
			"Wrong Content-Type",
			"POST",
			nil,
			map[string]string{"Content-Type": "text/html"},
			415,
		},
		{
			"Invalid User Data",
			"POST",
			map[string]string{
				"email":        "123@abc.com",
				"password":     "notthesame",
				"passwordConf": "asthisone",
				"userName":     "test",
				"firstName":    "",
				"lastName":     "",
			},
			map[string]string{"Content-Type": "application/json"},
			200,
		},
		{
			"Valid User",
			"POST",
			map[string]string{
				"email":        "123@abc.com",
				"password":     "testpassword",
				"passwordConf": "testpassword",
				"userName":     "t3st",
				"firstName":    "test",
				"lastName":     "user",
			},
			map[string]string{"Content-Type": "application/json"},
			201,
		},
	}

	for _, c := range cases {
		// create mock db
		db, mock, err := sqlmock.New()
		if err != nil {
			log.Fatalf("an error occurred while opening mock connection: %s", err)
		}
		defer db.Close()

		row := sqlmock.NewRows(trieCOLS)
		row.AddRow(1, "test", "test", "user")
		mock.ExpectQuery(regexp.QuoteMeta("select id, user_name, first_name, last_name from users")).WithArgs().WillReturnRows(row)
		trie, _ := u.BuildTrie(db)

		// make handler context with session store, session key for signing, and user db
		hc := HandlerContext{Key: "test Key", Session: s.NewMemStore(time.Hour, time.Minute), Users: u.NewMySqlStore(db), Trie: trie}

		// marshal message to json to send as buffer
		body, err := json.Marshal(c.params)
		if err != nil {
			log.Fatalln(err)
		}

		// make new request with given params
		req, err := http.NewRequest(c.method, "/v1/users", bytes.NewBuffer(body))
		if err != nil {
			t.Fatal(err)
		}

		if len(c.headers) != 0 {
			for k, v := range c.headers {
				req.Header.Set(k, v)
			}
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(hc.UsersHandler)
		if c.expected == 201 { // valid user
			// expect insert into RDBMS store
			mock.ExpectBegin()
			mock.ExpectExec(regexp.QuoteMeta(INS)).
				WithArgs(c.params["email"], sqlmock.AnyArg(), c.params["userName"], c.params["firstName"], c.params["lastName"], sqlmock.AnyArg()).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectCommit()
			if rr.Body == nil {
				t.Error("no returned json")
			}
			if rr.Body != nil { // check that password and email aren't exported fields from JSON
				m := make(map[string]string)
				j, _ := ioutil.ReadAll(rr.Body)
				json.Unmarshal([]byte(j), &m)
				if m["passHash"] != "" || m["email"] != "" {
					t.Errorf("JSON parameters not supposed to be sent in body: %v", m)
				}
			}
		}
		handler.ServeHTTP(rr, req)
		if status := rr.Code; status != c.expected {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, c.expected)
		}

	}
}

// TestSpecificUserHandler tests the functionality of the SpecificUserHandler function
func TestSpecificUserHandler(t *testing.T) {
	cases := []struct {
		name     string
		hint     string
		method   string
		id       string
		user     u.User
		params   map[string]string
		headers  map[string]string
		expected int
		auth     bool
	}{
		{
			"Wrong Request Method",
			"Make sure the Request method is only GET and PATCH",
			"POST",
			"",
			u.User{},
			nil,
			nil,
			405,
			false,
		},
		{
			"No Session",
			"Make sure that the user has been authenticated",
			"GET",
			"",
			u.User{},
			nil,
			nil,
			401,
			false,
		},
		{
			"Valid user with me as ID param",
			"Make sure that the authenticated user can use me instead of an ID",
			"PATCH",
			"me",
			u.User{
				ID:        1,
				Email:     "test@user.com",
				PassHash:  []byte{1, 12, 15, 20},
				UserName:  "Anonymous",
				FirstName: "Test",
				LastName:  "User",
				PhotoURL:  "https://www.gravatar.com/avatar/0bc83cb571cd1c50ba6f3e8a78ef1346",
			},
			map[string]string{
				"firstName": "test",
				"lastName":  "update",
			},
			map[string]string{"Content-Type": "application/json"},
			200,
			true,
		},
		{
			"Valid PATCH with ID",
			"Make sure that the authenticated user can update their profile",
			"PATCH",
			"1",
			u.User{
				ID:        1,
				Email:     "test@user.com",
				PassHash:  []byte{1, 12, 15, 20},
				UserName:  "Anonymous",
				FirstName: "Test",
				LastName:  "User",
				PhotoURL:  "https://www.gravatar.com/avatar/0bc83cb571cd1c50ba6f3e8a78ef1346",
			},
			map[string]string{
				"firstName": "test",
				"lastName":  "update",
			},
			map[string]string{"Content-Type": "application/json"},
			200,
			true,
		},
		{
			"Valid PATCH with incorrect ID",
			"Make sure that the authenticated user can only update their account",
			"PATCH",
			"2",
			u.User{
				ID:        1,
				Email:     "test@user.com",
				PassHash:  []byte{1, 12, 15, 20},
				UserName:  "Anonymous",
				FirstName: "Test",
				LastName:  "User",
				PhotoURL:  "https://www.gravatar.com/avatar/0bc83cb571cd1c50ba6f3e8a78ef1346",
			},
			map[string]string{},
			map[string]string{"Content-Type": "application/json"},
			403,
			true,
		},
		{
			"Valid User",
			"Make sure that the proper JSON is returned",
			"GET",
			"1",
			u.User{
				ID:        1,
				Email:     "test@user.com",
				PassHash:  []byte{1, 12, 15, 20},
				UserName:  "Anonymous",
				FirstName: "Test",
				LastName:  "User",
				PhotoURL:  "https://www.gravatar.com/avatar/0bc83cb571cd1c50ba6f3e8a78ef1346",
			},
			map[string]string{},
			map[string]string{"Content-Type": "application/json"},
			200,
			true,
		},
		{
			"Valid Session, Wrong User",
			"Make sure that the authorized user matches the user they are getting",
			"GET",
			"2",
			u.User{
				ID:        1,
				Email:     "test@user.com",
				PassHash:  []byte{1, 12, 15, 20},
				UserName:  "Anonymous",
				FirstName: "Test",
				LastName:  "User",
				PhotoURL:  "https://www.gravatar.com/avatar/0bc83cb571cd1c50ba6f3e8a78ef1346",
			},
			map[string]string{},
			map[string]string{"Content-Type": "application/json"},
			404,
			true,
		},
	}

	for _, c := range cases {
		// create mock db
		db, mock, err := sqlmock.New()
		if err != nil {
			log.Fatalf("an error occurred while opening mock connection: %s", err)
		}
		defer db.Close()

		// for ensuring Trie is built correctly
		row := sqlmock.NewRows(trieCOLS)
		row.AddRow(1, "test", "test", "user")
		mock.ExpectQuery(regexp.QuoteMeta("select id, user_name, first_name, last_name from users")).WithArgs().WillReturnRows(row)
		trie, _ := u.BuildTrie(db)

		// make handler context with session store, session key for signing, and user db
		hc := HandlerContext{Key: "test Key", Session: s.NewMemStore(time.Hour, time.Minute), Users: u.NewMySqlStore(db), Trie: trie}

		// marshal message to json to send as buffer
		body, err := json.Marshal(c.params)
		if err != nil {
			log.Fatalln(err)
		}

		// make new request with given params
		req, err := http.NewRequest(c.method, "/v1/users/"+c.id, bytes.NewBuffer(body))
		if err != nil {
			t.Fatal(err)
		}

		// add headers to request
		if len(c.headers) != 0 {
			for k, v := range c.headers {
				req.Header.Set(k, v)
			}
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(hc.SpecificUserHandler)
		if c.method == "GET" || c.method == "PATCH" { // if the test expects the right method
			// make session with user
			sid, _ := s.BeginSession(hc.Key, hc.Session, SessionState{SessionStart: time.Now(), User: c.user}, rr)
			if c.auth { // test that handler makes sure user is authorized
				req.Header.Set("Authorization", "Bearer "+string(sid)) // add auth header
			}
			// add expected to mock DB
			row := sqlmock.NewRows(COLS)
			row.AddRow(c.user.ID, c.user.Email, c.user.PassHash, c.user.UserName, c.user.FirstName, c.user.LastName, c.user.PhotoURL)
			if c.method == "GET" {
				mock.ExpectQuery(regexp.QuoteMeta(GETID)).WithArgs(c.user.ID).WillReturnRows(row)
			}

			if c.method == "PATCH" {
				// expect update to user store
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(UPD)). // updating row
									WithArgs(c.params["firstName"], c.params["lastName"], c.user.ID).
									WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectQuery(regexp.QuoteMeta(GETID)). // getting row to return
										WithArgs(c.user.ID).WillReturnRows(row) // get by id call
				mock.ExpectCommit()
			}
		}

		handler.ServeHTTP(rr, req)
		if status := rr.Code; status != c.expected { // make sure the right HTTP header is returned
			t.Errorf("%v: handler returned wrong status code: got %v want %v. \n HINT: %v",
				c.name, status, c.expected, c.hint)
		}
		// check that handler returns proper JSON
		if c.method == "PATCH" {
			if rr.Body == nil {
				t.Errorf("%v: no returned json", c.name)
			}
			if rr.Body != nil { // check that password and email aren't exported fields from JSON
				m := make(map[string]string)
				j, _ := ioutil.ReadAll(rr.Body)
				json.Unmarshal([]byte(j), &m)
				if m["firstName"] != c.params["firstName"] || m["lastName"] != c.params["lastName"] {
					t.Errorf("%v: User not updated properly:\n Wanted: %v", c.name, m)
				}
			}
		}
	}
}

// TestSessionsHandler tests the functionality of the SessionsHandler function
func TestSessionsHandler(t *testing.T) {
	cases := []struct {
		name     string
		method   string
		params   map[string]string
		headers  map[string]string
		expected int // expected http code to be returned
		needUser bool
	}{
		{
			"Wrong Request Method",
			"GET",
			nil,
			nil,
			405,
			false,
		},
		{
			"Wrong Content-Type",
			"POST",
			nil,
			map[string]string{"Content-Type": "text/html"},
			415,
			false,
		},
		{
			"No existing user",
			"POST",
			map[string]string{
				"email":    "wrong@email.com",
				"password": "wrongpass",
			},
			map[string]string{"Content-Type": "application/json"},
			401,
			true,
		},
		{
			"Invalid Password",
			"POST",
			map[string]string{
				"email":    "upd@test.com",
				"password": "wrongpass",
			},
			map[string]string{"Content-Type": "application/json"},
			401,
			true,
		},
		{
			"Authorized User",
			"POST",
			map[string]string{
				"email":    "upd@test.com",
				"password": "testpassword",
			},
			map[string]string{"Content-Type": "application/json"},
			201,
			true,
		},
	}

	for _, c := range cases {
		// create mock db
		db, mock, err := sqlmock.New()
		if err != nil {
			log.Fatalf("an error occurred while opening mock connection: %s", err)
		}
		defer db.Close()

		// make handler context with session store, session key for signing, and user db
		hc := HandlerContext{Key: "test Key", Session: s.NewMemStore(time.Hour, time.Minute), Users: u.NewMySqlStore(db)}

		// marshal message to json to send as buffer
		body, err := json.Marshal(c.params)
		if err != nil {
			log.Fatalln(err)
		}

		// make new request with given params
		req, err := http.NewRequest(c.method, "/v1/sessions/", bytes.NewBuffer(body))
		if err != nil {
			t.Fatal(err)
		}

		// add headers to request
		if len(c.headers) != 0 {
			for k, v := range c.headers {
				req.Header.Set(k, v)
			}
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(hc.SessionsHandler)
		var user u.User
		if c.method == "POST" && c.needUser {
			user = u.User{
				ID:    1,
				Email: "upd@test.com",
				PassHash: []byte{36, 50, 97, 36, 49, 51, 36, 57, 57, 116, 84, 122, 51, 103, 72, 56, 114, 68, 100, 56, 105, 108, 74, 116, 50, 86, 87, 77, 117, 110,
					74, 119, 83, 110, 114, 77, 106, 79, 71, 98, 103, 115, 68, 50, 99, 112, 104, 56, 55, 107, 114, 115, 107, 66, 70, 109, 84, 80, 48, 75},
				UserName:  "TestUSER",
				FirstName: "Test",
				LastName:  "User",
				PhotoURL:  "https://www.gravatar.com/avatar/0bc83cb571cd1c50ba6f3e8a78ef1346",
			}
			row := sqlmock.NewRows(COLS)
			row.AddRow(user.ID, user.Email, user.PassHash, user.UserName, user.FirstName, user.LastName, user.PhotoURL)
			if c.params["email"] != user.Email { // test that user passes right email cred
				mock.ExpectQuery(regexp.QuoteMeta(GETEMQUERY)).WithArgs(c.params["email"]).WillReturnError(sql.ErrNoRows)
			} else {
				mock.ExpectQuery(regexp.QuoteMeta(GETEMQUERY)).WithArgs(c.params["email"]).WillReturnRows(row)
			}
		}
		handler.ServeHTTP(rr, req)
		if status := rr.Code; status != c.expected { // make sure the right HTTP header is returned
			t.Errorf("%v: handler returned wrong status code: got %v want %v.\n",
				c.name, status, c.expected)
		}
		// is authorized user
		if c.params["email"] == user.Email && c.params["password"] == "testpassword" {
			// ensure proper json is returned
			var tmp u.User
			j, _ := ioutil.ReadAll(rr.Body)
			json.Unmarshal([]byte(j), &tmp)
			if tmp.ID != user.ID || tmp.UserName != user.UserName || tmp.FirstName != user.FirstName || tmp.LastName != user.LastName || tmp.PhotoURL != user.PhotoURL {
				t.Errorf("%v: handler returned wrong user: got %v want %v.\n",
					c.name, tmp, user)
			}
		}
	}
}

// TestSpecificSessionHandler tests the functionality of the SpecificSessionHandler function
func TestSpecificSessionHandler(t *testing.T) {
	cases := []struct {
		name     string
		method   string
		expected int // expected http code to be returned
		path     string
		needSess bool
	}{
		{
			"Wrong Request Method",
			"GET",
			405,
			"",
			false,
		},
		{
			"Path not '/mine'",
			"DELETE",
			403,
			"notmine",
			false,
		},
		{
			"No current session",
			"DELETE",
			200,
			"mine",
			false,
		},
		{
			"Valid input",
			"DELETE",
			200,
			"mine",
			true,
		},
	}

	for _, c := range cases {
		// create mock db
		db, mock, err := sqlmock.New()
		if err != nil {
			log.Fatalf("an error occurred while opening mock connection: %s", err)
		}
		defer db.Close()

		row := sqlmock.NewRows(trieCOLS)
		row.AddRow(1, "test", "test", "user")
		mock.ExpectQuery(regexp.QuoteMeta("select id, user_name, first_name, last_name from users")).WithArgs().WillReturnRows(row)
		trie, _ := u.BuildTrie(db)

		// make handler context with session store, session key for signing, and user db
		hc := HandlerContext{Key: "test Key", Session: s.NewMemStore(time.Hour, time.Minute), Users: u.NewMySqlStore(db), Trie: trie}

		// make new request with given params
		req, err := http.NewRequest(c.method, "/v1/sessions/"+c.path, nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(hc.SpecificSessionHandler)
		if c.method == "DELETE" && c.needSess {
			sid, _ := s.BeginSession(hc.Key, hc.Session, SessionState{SessionStart: time.Now(), User: u.User{}}, rr)
			req.Header.Add("Authorization", "Bearer "+string(sid)) // add auth header
		}

		handler.ServeHTTP(rr, req)
		if status := rr.Code; status != c.expected { // make sure the right HTTP header is returned
			t.Errorf("on %v: handler returned wrong status code: got %v want %v.\n",
				c.name, status, c.expected)
		}
		if c.name == "Valid input" {
			if _, err := s.GetState(req, hc.Key, hc.Session, SessionState{}); err == nil {
				t.Errorf("on %v: handler didn't end session\n",
					c.name)
			}
			res, _ := ioutil.ReadAll(rr.Body)
			if string(res) != "signed out" {
				t.Errorf("on %v: handler didn't return proper message: need 'signed out' message\n",
					c.name)
			}
		}

	}
}
