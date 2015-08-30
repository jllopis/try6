package store

import (
	"time"

	"github.com/jllopis/try6"
	"github.com/jllopis/try6/log"
)

// Scoper defines the methods needed to manage Tenants
type Scoper interface {
	SaveScope(s *try6.Scope) error
}

// SaveScope persist the scope data to the database
func (d *DefaultStore) SaveScope(s *try6.Scope) error {
	log.LogD("Saving Scope", "pkg", "store", "func", "SaveScope(*try6.Scope)", "data", s)
	now := time.Now().UTC()
	s.Updated = now
	if s.ID == "" {
		// New Scope
		s.Created = now
		return d.C.InsertInto("scopes").Blacklist("id", "deleted").Record(s).Returning("id").QueryScalar(&s.ID)
	}
	return d.C.Update("scopes").SetBlacklist(s, "id", "tenant_uid", "created").Where("id=$1", s.ID).Returning("*").QueryStruct(s)
}
