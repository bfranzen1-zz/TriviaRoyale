package data

// User Statistics
// user_stat id
// game id
// user id
// num correct
// won: bool
// points -- TODO determine awards

// Game
// id
// players (array of user IDs)
// questions (array of questions)
// 		question
// 			question answer options
// 			right answer
// winnerID
//

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type MongoStore struct {
	ses *mgo.Session
}

func NewMongoStore(ses *mgo.Session) *MongoStore {
	return &MongoStore{
		ses: ses,
	}
}

// GetByID takes an bson id, collection string and destination interface to put the resulting
// mongo query into.
func (ms *MongoStore) GetByID(id bson.ObjectId, coll string, dest interface{}) interface{} {
	return ms.ses.DB("game").C(coll).Find(bson.M{"_id": string(id)}).One(&dest)
}

// Insert inserts the object into the collection
// and returns an error if any occur
func (ms *MongoStore) Insert(obj interface{}, coll string) error {
	c := ms.ses.DB("game").C(coll)

	if err := c.Insert(obj); err != nil {
		return err
	}

	return nil
}

// Update updates the record at id in the coll collection using the updates interface
// passed in. Errors are returned if the update fails
func (ms *MongoStore) Update(id bson.ObjectId, coll string, updates interface{}) error {
	c := ms.ses.DB("game").C(coll)

	if err := c.UpdateId(id, updates); err != nil {
		return err
	}

	return nil
}

// Delete deletes the record at id from the passed coll collection
// an error is returned if the delete fails
func (ms *MongoStore) Delete(id bson.ObjectId, coll string) error {
	c := ms.ses.DB("game").C(coll)

	if err := c.RemoveId(id); err != nil {
		return err
	}

	return nil
}
