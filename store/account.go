package store

import (
	"time"

	"github.com/jllopis/try6"
	"github.com/jllopis/try6/log"
)

// Accounter interface defines the method to be implemented for account storage managers
type Accounter interface {
	//	LoadAllAccounts() ([]*account.Account, error)
	//	LoadAccount(tuuid string) (*account.Account, error)
	SaveAccount(directory string, a *try6.Account) error
	//	DeleteAccount(uuid string) error
	//	GetAccountByEmail(email string) (*account.Account, error)
	//	ExistAccount(uuid string) bool
}

// SaveAccount persist the account data to the database
func (d *DefaultStore) SaveAccount(directory string, t *try6.Account) error {
	log.LogD("Saving Account", "pkg", "store", "func", "SaveAccount(*try6.Account)", "directory", directory, "data", t)
	now := time.Now().UTC()
	t.Updated = now
	if t.ID == "" {
		// New Account
		t.Created = now
		t.Status = "active"
		if err := d.C.InsertInto("accounts").Blacklist("id", "deleted").Record(t).Returning("*").QueryStruct(t); err != nil {
			log.LogE("error saving account", "pkg", "store", "func", "SaveAccount(*try6.Account)", "error", err.Error())
			return err
		}
	} else {
		if err := d.C.Update("accounts").SetBlacklist(t, "id", "created").Where("id=$1", t.ID).Returning("*").QueryStruct(t); err != nil {
			log.LogE("error updating account", "pkg", "store", "func", "SaveAccount(*try6.Account)", "error", err.Error())
			return err
		}
	}

	log.LogD("account inserted", "pkg", "store", "func", "SaveAccount(*try6.Account)", "data", t)
	// add to directory
	if _, err := d.C.Upsert("directory_account").Columns("directory_id", "account_id", "created", "updated").Record(&try6.DirectoryAccount{
		DirectoryID: directory,
		AccountID:   t.ID,
		Created:     now,
		Updated:     now,
	}).Where("directory_id=$1 AND account_id=$2", directory, t.ID).Exec(); err != nil {
		log.LogE("error updating directory", "pkg", "store", "func", "SaveAccount(*try6.Account)", "error", err.Error())
		return err
	}
	return nil
}
