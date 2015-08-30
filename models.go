package try6

import (
	"time"

	"gopkg.in/mgutz/dat.v1"
)

// Tenant holds the items that compose a tenant in try5
type Tenant struct {
	ID      string       `json:"id" db:"id"`
	Label   string       `json:"label" db:"label"`
	Status  string       `json:"status" db:"status"`
	Created time.Time    `json:"created" db:"created"`
	Updated time.Time    `json:"updated" db:"updated"`
	Deleted dat.NullTime `json:"deleted,omitempty" db:"deleted"`
}

// CreateTenantData holds all the info needed to create a new tenant. Each tenant has a
// default management account (Acc) that is added to the default admin directory (Dir)
// and can not be deleted.
// The account must exist previously to the Tenant creation so if the provided one
// does not, it will be created an assigned as the default admin account for the tenant.
type CreateTenantData struct {
	TData *Tenant    `json:"tenant"`
	Dir   *Directory `json:"directory"`
	Acc   *Account   `json:"account"`
	Scope *Scope     `json:"scope"`
}

// Directory holds the items related to a directory. A directory group auth data together
type Directory struct {
	ID          string       `json:"id" db:"id"`
	TenantUID   string       `json:"tenant_uid" db:"tenant_uid"`
	Label       string       `json:"label" db:"label"`
	Description string       `json:"description" db:"description"`
	Status      string       `json:"status" db:"status"`
	Created     time.Time    `json:"created" db:"created"`
	Updated     time.Time    `json:"updated" db:"updated"`
	Deleted     dat.NullTime `json:"deleted,omitempty" db:"deleted"`
}

// Account hold the information of an Account type. Variables are of type pointer
// to easily identify null variables when persist/read to/from database storage.
type Account struct {
	ID       string       `json:"id" db:"id"`
	Email    string       `json:"email" db:"email"`
	Name     string       `json:"name,omitempty" db:"name"`
	Password string       `json:"password,omitempty" db:"password"`
	Status   string       `json:"status" db:"status"`
	Created  time.Time    `json:"created" db:"created"`
	Updated  time.Time    `json:"updated" db:"updated"`
	Deleted  dat.NullTime `json:"deleted,omitempty" db:"deleted"`
	//	Gravatar *string    `json:"gravatar,omitempty" db:"gravatar"`
}

// DirectoryAccount hold the grouping of accounts into directories
type DirectoryAccount struct {
	DirectoryID string       `json:"directory_id" db:"directory_id"`
	AccountID   string       `json:"account_id" db:"account_id"`
	Created     time.Time    `json:"created" db:"created"`
	Updated     time.Time    `json:"updated" db:"updated"`
	Deleted     dat.NullTime `json:"deleted,omitempty" db:"deleted"`
}

// Key is the default RSA key associated to an account
type Key struct {
	ID        string       `json:"id" db:"id"`
	AccountID string       `json:"account_id" db:"account_id"`
	PubKey    []byte       `json:"pubkey" db:"pub_key"`
	PrivKey   []byte       `json:"privkey" db:"priv_key"`
	Status    string       `json:"status" db:"status"`
	Created   time.Time    `json:"created" db:"created"`
	Updated   time.Time    `json:"updated" db:"updated"`
	Deleted   dat.NullTime `json:"deleted,omitempty" db:"deleted"`
}

// Scope holds the items related to a scope. A scope can be thougth of as an application
type Scope struct {
	ID          string       `json:"id" db:"id"`
	TenantID    string       `json:"tenant_id" db:"tenant_id"`
	Label       string       `json:"label" db:"label"`
	Description string       `json:"description" db:"description"`
	Status      string       `json:"status" db:"status"`
	Created     time.Time    `json:"created" db:"created"`
	Updated     time.Time    `json:"updated" db:"updated"`
	Deleted     dat.NullTime `json:"deleted,omitempty" db:"deleted"`
}

// DirectoryScope hold the grouping of accounts into directories
type DirectoryScope struct {
	DirectoryID         string       `json:"directory_id" db:"directory_id"`
	ScopeID             string       `json:"scope_id" db:"scope_id"`
	Priority            int64        `json:"priority" db:"priority"`
	IsDefaultAccStore   bool         `json:"is_default_account_store" db:"is_default_account_store"`
	IsDefaultGroupStore bool         `json:"is_default_group_store" db:"is_default_group_store"`
	IsDefaultRBACStore  bool         `json:"is_default_rbac_store" db:"is_default_rbac_store"`
	Created             time.Time    `json:"created" db:"created"`
	Updated             time.Time    `json:"updated" db:"updated"`
	Deleted             dat.NullTime `json:"deleted,omitempty" db:"deleted"`
}
