package store

import (
	"time"

	"github.com/jllopis/try6"
	"github.com/jllopis/try6/log"
)

// SaveScope persist the scope data to the database
func (d *DefaultStore) SaveScope(t *try6.Scope) error {
	log.LogD("Saving Scope", "pkg", "store", "func", "SaveScope(*try6.Scope)", "data", t)
	now := time.Now().UTC()
	t.Updated = now
	if t.ID == "" {
		// New Scope
		t.Created = now
		return d.C.InsertInto("scopes").Blacklist("id", "deleted").Record(t).Returning("id").QueryScalar(&t.ID)
	}
	return d.C.Update("scopes").SetBlacklist(t, "id", "tenant_uid", "created").Where("id=$1", t.ID).Returning("*").QueryStruct(t)
}
