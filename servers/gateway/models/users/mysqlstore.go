package users

import (
	"database/sql"
	"errors"
	"github.com/bfranzen1/assignments-bfranzen1/servers/gateway/indexes"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	"strings"
)

const get = "select * from users where id = ?"
const getEM = "select * from users where email = ?"
const getUN = "select * from users where user_name = ?"
const ins = "insert into users(email,pass_hash,user_name,first_name,last_name,photo_url) values(?,?,?,?,?,?)"
const upd = "update users set first_name = ?, last_name = ? where id = ?"
const del = "delete from users where id = ?"
const ins_SI = "insert into signin(id, signin_time, ip) values(?,?,?)"
const getUSERS = "select id, user_name, first_name, last_name from users"

type MySqlStore struct {
	db   *sql.DB
	trie *indexes.Trie
}

func NewMySqlStore(db *sql.DB) *MySqlStore {
	if db == nil {
		return nil
	}
	return &MySqlStore{
		db: db,
	}
}

func (mysql *MySqlStore) get(col string, val string) (*User, error) {
	var row *sql.Row
	if col == "id" { // getting by id
		id, _ := strconv.ParseInt(val, 10, 64)
		row = mysql.db.QueryRow(get, id)
	} else if col == "email" {
		row = mysql.db.QueryRow(getEM, val)
	} else {
		row = mysql.db.QueryRow(getUN, val)
	}
	user := User{}

	// scan row values into user struct
	if err := row.Scan(&user.ID, &user.Email, &user.PassHash,
		&user.UserName, &user.FirstName, &user.LastName,
		&user.PhotoURL); err != nil {
		return nil, ErrUserNotFound
	}
	return &user, nil
}

//GetByID returns the User with the given ID
func (mysql *MySqlStore) GetByID(id int64) (*User, error) {
	return mysql.get("id", strconv.FormatInt(id, 10))
}

//GetByEmail returns the User with the given email
func (mysql *MySqlStore) GetByEmail(email string) (*User, error) {
	return mysql.get("email", email)
}

//GetByuser_name returns the User with the given user_name
func (mysql *MySqlStore) GetByUserName(user_name string) (*User, error) {
	return mysql.get("user_name", user_name)
}

//Insert inserts the user into the database, and returns
//the newly-inserted User, complete with the DBMS-assigned ID
func (mysql *MySqlStore) Insert(user *User) (*User, error) {
	tx, err := mysql.db.Begin() // begin transaction
	if err != nil {
		return nil, err
	}

	res, err := tx.Exec(ins, user.Email, user.PassHash, user.UserName, user.FirstName, user.LastName, user.PhotoURL)
	if err != nil {
		tx.Rollback() // rollback transaction
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		tx.Rollback() // rollback transaction
		return nil, err
	}
	user.ID = id
	tx.Commit()
	return user, nil
}

//Update applies UserUpdates to the given user ID
//and returns the newly-updated user
func (mysql *MySqlStore) Update(id int64, updates *Updates) (*User, error) {
	tx, err := mysql.db.Begin() // begin transaction
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(upd, updates.FirstName, updates.LastName, id)
	if err != nil {
		tx.Rollback() // rollback tx
		return nil, err
	}
	// get user
	usr, err := mysql.GetByID(id)
	if err != nil {
		tx.Rollback() // rollback tx
		return nil, err
	}
	if err := usr.ApplyUpdates(updates); err != nil { //apply updates if valid
		tx.Rollback()
		return nil, err
	}
	tx.Commit() // commit transation
	return usr, nil
}

//Delete deletes the user with the given ID
func (mysql *MySqlStore) Delete(id int64) error {
	tx, err := mysql.db.Begin() // begin transaction
	if err != nil {
		return err
	}

	if _, err := tx.Exec(del, id); err != nil {
		tx.Rollback() // rollback anything thats been done
		return ErrUserNotFound
	}
	tx.Commit() // commit transaction
	return nil
}

// InsertSignIn inserts a row into the signin table. An error is returned if the
// insert transaction fails at any point.
func (mysql *MySqlStore) InsertSignIn(id int64, dt string, ip string) error {
	tx, err := mysql.db.Begin() // begin transaction
	if err != nil {
		return err
	}

	_, err = tx.Exec(ins_SI, id, dt, ip)
	if err != nil {
		tx.Rollback() // rollback transaction
		return err
	}
	tx.Commit()
	return nil
}

// BuildTrie builds a new trie struct using all the current
// users in the user store and returns the trie or any errors that occur
func BuildTrie(db *sql.DB) (*indexes.Trie, error) {
	trie := indexes.NewTrie()
	rows, err := db.Query(getUSERS)
	if err != nil {
		//return empty trie
		return trie, errors.New("Unable to query rows")
	}
	defer rows.Close()
	for rows.Next() {
		tmp := User{}
		if err := rows.Scan(&tmp.ID, &tmp.UserName, &tmp.FirstName, &tmp.LastName); err != nil {
			return trie, errors.New("error scanning row")
		}
		trie = AddField(tmp.UserName, trie, tmp.ID)
		trie = AddField(tmp.LastName, trie, tmp.ID)
		trie = AddField(tmp.FirstName, trie, tmp.ID)
	}
	return trie, nil
}

// addField takes in a field string, trie to add to, and the id of the field
// and adds it to the trie. If the field contains multiple words they are added
// as individual fields
func AddField(field string, t *indexes.Trie, id int64) *indexes.Trie {
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
