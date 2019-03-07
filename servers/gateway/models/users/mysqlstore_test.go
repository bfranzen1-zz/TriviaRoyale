package users

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"testing"

	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

const getID = "select * from users where id = ?"
const getEMQUERY = "select * from users where email = ?"
const getUSRQUERY = "select * from users where user_name = ?"
const INS = "insert into users(email,pass_hash,user_name,first_name,last_name,photo_url) values(?,?,?,?,?,?)"
const UPD = "update users set first_name = ?, last_name = ? where id = ?"
const DEL = "delete from users where id = ?"

var insertErr = fmt.Errorf("Error executing INSERT operation")

var COLS = []string{"id", "email", "pass_hash", "user_name", "first_name", "last_name", "photo_url"}

func TestGetFuncs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("an error occured while opening mock connection: %s", err)
	}
	defer db.Close()

	exp := &User{ // user to be returned on valid query
		ID:        1,
		Email:     "test@user.com",
		PassHash:  []byte{1, 12, 15, 20, 20, 20, 16, 10, 4, 3},
		UserName:  "Anonymous",
		FirstName: "Test",
		LastName:  "User",
		PhotoURL:  "https://www.gravatar.com/avatar/0bc83cb571cd1c50ba6f3e8a78ef1346",
	}

	// make store to use
	store := NewMySqlStore(db)

	// insert user to test
	row := sqlmock.NewRows(COLS)
	row.AddRow(exp.ID, exp.Email, exp.PassHash, exp.UserName, exp.FirstName, exp.LastName, exp.PhotoURL)

	//******************************************//
	// *********** TESTING GetByID *********** //
	//*****************************************//

	// EXPECT NO ERRORS
	mock.ExpectQuery(regexp.QuoteMeta(getID)).WithArgs(exp.ID).WillReturnRows(row)

	usr, err := store.GetByID(exp.ID)
	if err != nil {
		t.Errorf("Unexpected error in GetByID: %v", err)
	}

	// make sure the user returned matches the exp user
	if err == nil && !reflect.DeepEqual(usr, exp) {
		t.Errorf("user returned by GetByID doesn't match expected user")
	}

	// EXPECT ERROR
	mock.ExpectQuery(regexp.QuoteMeta(getID)).WithArgs(-1).WillReturnError(sql.ErrNoRows)
	if _, err = store.GetByID(-1); err == nil {
		t.Errorf("expected error in GetByID: %v, but didn't get it", sql.ErrNoRows)
	}

	//********************************************//
	// *********** TESTING GetByEmail *********** //
	//********************************************//

	row.AddRow(exp.ID, exp.Email, exp.PassHash, exp.UserName, exp.FirstName, exp.LastName, exp.PhotoURL)
	mock.ExpectQuery(regexp.QuoteMeta(getEMQUERY)).WithArgs(exp.Email).WillReturnRows(row)

	usr, err = store.GetByEmail(exp.Email)
	if err != nil {
		t.Errorf("Unexpected error in GetByEmail: %v", err)
	}

	// make sure the user returned matches the exp user
	if err == nil && !reflect.DeepEqual(usr, exp) {
		t.Errorf("user returned by GetByEmail doesn't match expected user")
	}

	// EXPECT ERROR - Email doesn't exist
	mock.ExpectQuery(regexp.QuoteMeta(getEMQUERY)).WithArgs("not@real.com").WillReturnError(sql.ErrNoRows)
	if _, err = store.GetByEmail("not@real.com"); err == nil {
		t.Errorf("expected error in GetByEmail: %v, but didn't get it", sql.ErrNoRows)
	}

	//***********************************************//
	// *********** TESTING GetByUserName *********** //
	//***********************************************//

	row.AddRow(exp.ID, exp.Email, exp.PassHash, exp.UserName, exp.FirstName, exp.LastName, exp.PhotoURL)
	mock.ExpectQuery(regexp.QuoteMeta(getUSRQUERY)).WithArgs(exp.UserName).WillReturnRows(row)

	usr, err = store.GetByUserName(exp.UserName)
	if err != nil {
		t.Errorf("Unexpected error in GetByUserName: %v", err)
	}
	// make sure the user returned matches the exp user
	if err == nil && !reflect.DeepEqual(usr, exp) {
		t.Errorf("user returned by GetByUserName doesn't match expected user")
	}

	// EXPECT ERROR - UserName doesn't exist
	mock.ExpectQuery(regexp.QuoteMeta(getUSRQUERY)).WithArgs("invalid").WillReturnError(sql.ErrNoRows)
	if _, err = store.GetByUserName("invalid"); err == nil {
		t.Errorf("expected error in GetByUserName: %v, but didn't get it", sql.ErrNoRows)
	}

	// Make sure all expectations are properly met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet sqlmock expectations: %v", err)
	}
}

func TestInsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("an error occured while opening mock connection: %s", err)
	}
	defer db.Close()

	exp := &User{ // user to be inserted
		Email:     "ins@test.com",
		PassHash:  []byte{1, 12, 15, 20},
		UserName:  "Insert",
		FirstName: "Test",
		LastName:  "User",
		PhotoURL:  "https://www.gravatar.com/avatar/0bc83cb571cd1c50ba6f3e8a78ef1346",
	}

	// make store to use
	store := NewMySqlStore(db)

	// should succeed
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(INS)).
		WithArgs(exp.Email, exp.PassHash, exp.UserName, exp.FirstName, exp.LastName, exp.PhotoURL).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	usr, err := store.Insert(exp)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// make sure returned user matches expected
	if err == nil && !reflect.DeepEqual(usr, exp) {
		t.Error("user returned doesn't match expected user")
	}

	usr2 := &User{
		Email:    "usr2@test.com",
		UserName: "test2",
		PassHash: []byte{1, 2, 3, 4, 5},
	}

	// inserting when records exist
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(INS)).
		WithArgs(usr2.Email, usr2.PassHash, usr2.UserName, usr2.FirstName, usr2.LastName, usr2.PhotoURL).
		WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectCommit()

	nu, err := store.Insert(usr2)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// make sure returned user matches expected
	if err == nil && !reflect.DeepEqual(nu, usr2) {
		t.Error("user returned doesn't match expected user")
	}

	inv := &User{ // invalid user
		ID:       1,
		Email:    "usr2@test.com",
		UserName: "test2",
		PassHash: []byte{1, 2, 3, 4, 5},
	}

	// inserting invalid user
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(INS)).
		WithArgs(inv.Email, inv.PassHash, inv.UserName, inv.FirstName, inv.LastName, inv.PhotoURL).
		WillReturnError(insertErr)
	mock.ExpectRollback()

	if _, err = store.Insert(inv); err == nil {
		t.Errorf("expected error: %v, but didn't get it", insertErr)
	}

	// Make sure all expectations are properly met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet sqlmock expectations: %v", err)
	}
}

func TestUpdate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("an error occured while opening mock connection: %s", err)
	}
	defer db.Close()

	exp := &User{ // user to be updated
		ID:        1,
		Email:     "upd@test.com",
		PassHash:  []byte{1, 12, 15, 20},
		UserName:  "Update",
		FirstName: "Test",
		LastName:  "User",
		PhotoURL:  "https://www.gravatar.com/avatar/0bc83cb571cd1c50ba6f3e8a78ef1346",
	}

	upd := &Updates{
		FirstName: "New",
		LastName:  "Names",
	}

	// make store to use
	store := NewMySqlStore(db)

	// add expected to mock DB
	row := sqlmock.NewRows(COLS)
	row.AddRow(exp.ID, exp.Email, exp.PassHash, exp.UserName, exp.FirstName, exp.LastName, exp.PhotoURL)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(UPD)). // updating row
						WithArgs(upd.FirstName, upd.LastName, exp.ID).
						WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(regexp.QuoteMeta(getID)). // getting row to return
							WithArgs(exp.ID).WillReturnRows(row) // get by id call
	mock.ExpectCommit()

	usr, err := store.Update(exp.ID, upd)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if err == nil && (usr.FirstName != upd.FirstName || usr.LastName != upd.LastName) {
		t.Error("user First and Last name doesn't match updates")
	}

	// non-existent row, expect error
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(UPD)). // updating row
						WithArgs(upd.FirstName, upd.LastName, -1).
						WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	_, err = store.Update(-1, upd)
	if err == nil {
		t.Errorf("Expected error but received none")
	}

	// nil update struct
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(UPD)).
		WithArgs("", "", 1).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	_, err = store.Update(1, &Updates{}) // nil update
	if err == nil {
		t.Errorf("Expected error but received none")
	}

	// Make sure all expectations are properly met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet sqlmock expectations: %v", err)
	}

}

func TestDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("an error occured while opening mock connection: %s", err)
	}
	defer db.Close()

	exp := &User{ // user to be deleted
		ID:        1,
		Email:     "test@user.com",
		PassHash:  []byte{1, 12, 15, 20, 20, 20, 16, 10, 4, 3},
		UserName:  "Anonymous",
		FirstName: "Test",
		LastName:  "User",
		PhotoURL:  "https://www.gravatar.com/avatar/0bc83cb571cd1c50ba6f3e8a78ef1346",
	}

	// make store to use
	store := NewMySqlStore(db)

	// add expected to mock DB
	row := sqlmock.NewRows(COLS)
	row.AddRow(exp.ID, exp.Email, exp.PassHash, exp.UserName, exp.FirstName, exp.LastName, exp.PhotoURL)

	// expect no error
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(DEL)). // updating row
						WithArgs(exp.ID).
						WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = store.Delete(exp.ID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// row doesn't exist, expect error
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(DEL)). // updating row
						WithArgs(-1).
						WillReturnError(sql.ErrConnDone)
	mock.ExpectRollback()

	err = store.Delete(-1)
	if err == nil {
		t.Errorf("expected error: %v, but received none", sql.ErrNoRows)
	}

	// Make sure all expectations are properly met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet sqlmock expectations: %v", err)
	}
}
