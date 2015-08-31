package store

import (
	"time"

	"github.com/jllopis/try6"
	"github.com/jllopis/try6/log"
)

// Scoper defines the methods needed to manage Tenants
type Scoper interface {
	SaveScope(s *try6.Scope) error
	GetScopesByTenantID(id string) ([]*try6.Scope, error)
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

// GetScopesByTenantID returns a list of scopes owned by the tenant or an error
// if something goes wrong
func (d *DefaultStore) GetScopesByTenantID(id string) ([]*try6.Scope, error) {
	log.LogD("Listing Scopes", "pkg", "store", "func", "GetScopesByTenantID(id string)", "tenantID", id)
	var scopes []*try6.Scope
	err := d.C.Select("*").From("scopes").Where("tenant_id=$1 AND deleted IS NULL", id).QueryStructs(&scopes)
	if err != nil {
		return nil, err
	}
	return scopes, nil
}
