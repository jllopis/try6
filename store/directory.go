package store

import (
	"time"

	"github.com/jllopis/try6"
	"github.com/jllopis/try6/log"
)

// SaveDirectory persist the directory data to the database
func (d *DefaultStore) SaveDirectory(t *try6.Directory) error {
	log.LogD("Saving Directory", "pkg", "store", "func", "SaveDirectory(*try6.Directory)", "data", t)
	now := time.Now().UTC()
	t.Updated = now
	if t.ID == "" {
		// New Directory
		t.Created = now
		return d.C.InsertInto("directories").Blacklist("id", "deleted").Record(t).Returning("id").QueryScalar(&t.ID)
	}
	return d.C.Update("directories").SetBlacklist(t, "id", "tenant_uid", "created").Where("id=$1", t.ID).Returning("*").QueryStruct(t)
}
