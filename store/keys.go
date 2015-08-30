package store

import (
	"time"

	"github.com/jllopis/try6"
	"github.com/jllopis/try6/log"
)

// Keyer mandates the methods to implement when dealing with keys
type Keyer interface {
	//	LoadAllKeys() ([]*keys.Key, error)
	//	LoadKey(kid string) (*keys.Key, error)
	SaveKey(key *try6.Key) error
	//	DeleteKey(kid string) error
	//	GetKeyByAccountID(uid string) (*keys.Key, error)
	//	GetKeyByEmail(email string) (*keys.Key, error)
	//	GetKeyByPub(pubkey []byte) (*keys.Key, error)
}

// SaveKey persist the key data to the database
func (d *DefaultStore) SaveKey(key *try6.Key) error {
	log.LogD("Saving Key", "pkg", "store", "func", "SaveKey(*try6.Key)", "data", key)
	now := time.Now().UTC()
	key.Updated = now
	if key.ID == "" {
		// New Key
		key.Created = now
		key.Status = "active"
		if err := d.C.InsertInto("keys").Blacklist("id", "deleted").Record(key).Returning("*").QueryStruct(key); err != nil {
			log.LogE("error saving key", "pkg", "store", "func", "SaveKey(*try6.Key)", "error", err.Error())
			return err
		}
	} else {
		if err := d.C.Update("keys").SetBlacklist(key, "id", "created").Where("id=$1", key.ID).Returning("*").QueryStruct(key); err != nil {
			log.LogE("error updating key", "pkg", "store", "func", "SaveKey(*try6.Key)", "error", err.Error())
			return err
		}
	}

	log.LogD("key inserted", "pkg", "store", "func", "SaveKey(*try6.Key)", "data", key)
	return nil
}
