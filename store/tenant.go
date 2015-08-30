package store

import (
	"time"

	"github.com/jllopis/try6"
	"github.com/jllopis/try6/log"
	"github.com/jllopis/try6/tryerr"
)

// Tenanter defines the methods needed to manage Tenants
type Tenanter interface {
	CreateTenant(data *try6.CreateTenantData) error
	SaveTenant(tenant *try6.Tenant) error
}

// SaveTenant persist the tenant data to the database
func (d *DefaultStore) SaveTenant(t *try6.Tenant) error {
	log.LogD("Saving Tenant", "pkg", "store", "func", "SaveTenant(*try6.Tenant)", "data", t)
	now := time.Now().UTC()
	t.Updated = now
	if t.ID == "" {
		// New Tenant
		t.Created = now
		return d.C.InsertInto("tenants").Blacklist("id", "deleted").Record(t).Returning("id").QueryScalar(&t.ID)
	}
	return d.C.Update("tenants").SetBlacklist(t, "id", "label", "created").Where("id=$1", t.ID).Returning("*").QueryStruct(t)
}

// CreateTenant creates a new tenant with the data provided in try6.CreateTenantData.
// The steps are:
//   1. Create a new tenant in the database
//   2. Create the admin directory for the tenant where the tenant admin accounts will live
//   3. If an account is provided (it is created in a previous step), it will be assigned as default admin account
//      If no account is provide, a new one is created and made the default admin account. In this case an RSA Key pair is created for the account
//   4. A default scope is created for admin purposes. As the account, it can be created prior to the call to NewTenant and be used here
//   5. Maps the admin scope to the admin directory so the tenant can modify it
//
// It is responsability of the caller that the account and scope data are valid for the tenant administration. Usually, the account will be created
// as part of a registration process and will be used here leaving the scope and directory data empty so it will be created here.
func (d *DefaultStore) CreateTenant(data *try6.CreateTenantData) error {
	now := time.Now().UTC()
	if data.TData.ID != "" {
		log.LogE("error creating tenant", "pkg", "store", "func", "CreateTenant(*try6.CreateTenantData)", "error", tryerr.ErrIDNotNull.Error())
		return tryerr.ErrIDNotNull
	}

	// 1. Create Tenant
	log.LogD("creating tenant", "pkg", "store", "func", "CreateTenant(*try6.CreateTenantData)", "data", data)
	if err := d.SaveTenant(data.TData); err != nil {
		log.LogE("error creating tenant", "pkg", "store", "func", "CreateTenant(*try6.CreateTenantData)", "error", err.Error())
		return err
	}

	// 2. Create Tenant Admin Directory.
	if data.Dir == nil {
		// directory does not exist. Create!
		data.Dir = &try6.Directory{
			Created: now,
			Updated: now,
		}
		log.LogD("data directory not provided. creating new", "pkg", "store", "func", "CreateTenant(*try6.CreateTenantData)", "directory", data.Dir)
	}

	data.Dir.TenantUID = data.TData.ID

	if data.Dir.Label == "" {
		data.Dir.Label = "Default Admin Directory"
	}
	if data.Dir.Description == "" {
		data.Dir.Description = "Default directory to hold administrative accounts for this tenant"
	}
	if data.Dir.Status == "" {
		data.Dir.Status = "active"
	}

	if err := d.SaveDirectory(data.Dir); err != nil {
		log.LogE("Could not create Directory", "pkg", "store", "func", "CreateTenant(*try6.CreateTenantData)", "error", err)
		return err
	}
	// 3. Add acc to Tenant Admin Directory as default admin account
	if data.Acc == nil {
		log.LogE("nil account", "pkg", "store", "func", "CreateTenant(*try6.CreateTenantData)", "error", "an admin account is needed")
		return tryerr.ErrAccountNotProvided
	}
	if data.Acc.ID == "" {
		// New account
		if err := data.Acc.UpdatePassword(data.Acc.Password); err != nil {
			log.LogW("Could not update password for admin account", "pkg", "store", "func", "CreateTenant(*try6.CreateTenantData)", "error", err)
		}
		if err := d.SaveAccount(data.Dir.ID, data.Acc); err != nil {
			log.LogE("Could not create admin account", "pkg", "store", "func", "CreateTenant(*try6.CreateTenantData)", "error", err)
			return err
		}

		// 4. Create RSA Keys for acc as it is new
		k := try6.NewKey(data.Acc.ID)
		if err := d.SaveKey(k); err != nil {
			log.LogE("Could not create rsa key for admin account", "pkg", "store", "func", "CreateTenant(*try6.CreateTenantData)", "error", err)
			return err
		}
	} else {
		// account exists. Must add it to the admin directory
		if _, err := d.C.Upsert("directory_account").Columns("directory_id", "account_id", "created", "updated").Record(&try6.DirectoryAccount{
			DirectoryID: data.Dir.ID,
			AccountID:   data.Acc.ID,
			Created:     now,
			Updated:     now,
		}).Where("directory_id=$1 AND account_id=$2", data.Dir.ID, data.Acc.ID).Exec(); err != nil {
			log.LogE("error updating directory", "pkg", "store", "func", "CreateTenant(*try6.CreateTenantData)", "error", err.Error())
			return err
		}
	}

	var aki []string
	if err := d.C.Select("id").From("keys").Where("account_id=$1 AND deleted IS NULL", data.Acc.ID).QuerySlice(&aki); err != nil {
		log.LogW("account has no rsa keys", "pkg", "store", "func", "CreateTenant(*try6.CreateTenantData)", "error", err.Error())
	}

	// 5. Create the default admin scope for the tenant to give access to the tenant management)
	if data.Scope == nil {
		// directory does not exist. Create!
		data.Scope = &try6.Scope{
			Created: now,
			Updated: now,
		}
		log.LogD("data scope not provided. creating new", "pkg", "store", "func", "CreateTenant(*try6.CreateTenantData)", "scope", data.Scope)
	}

	data.Scope.TenantID = data.TData.ID
	if data.Scope.Label == "" {
		data.Scope.Label = "Default Admin Scope"
	}
	if data.Scope.Description == "" {
		data.Scope.Description = "Default scope to administer this tenant"
	}
	if data.Scope.Status == "" {
		data.Scope.Status = "active"
	}
	if err := d.SaveScope(data.Scope); err != nil {
		log.LogE("Could not create admin Scope", "pkg", "store", "func", "CreateTenant(*try6.CreateTenantData)", "error", err)
		return err
	}
	// 6. Map the admin app with the admin directory created
	if _, err := d.C.Upsert("directory_scope").Columns("directory_id", "scope_id", "priority", "is_default_account_store", "is_default_group_store", "is_default_rbac_store", "created", "updated").Record(&try6.DirectoryScope{
		DirectoryID:         data.Dir.ID,
		ScopeID:             data.Scope.ID,
		Priority:            1,
		IsDefaultAccStore:   true,
		IsDefaultGroupStore: true,
		IsDefaultRBACStore:  true,
		Created:             now,
		Updated:             now,
	}).Where("directory_id=$1 AND scope_id=$2", data.Dir.ID, data.Scope.ID).Exec(); err != nil {
		log.LogE("error updating directory_scope", "pkg", "store", "func", "CreateTenant(*try6.CreateTenantData)", "error", err.Error())
		return err
	}

	log.LogD("New tenant created", "pkg", "tenant", "func", "CreateTenant(*try6.CreateTenantData)", "tenantID", data.TData.ID, "directoryID", data.Scope.ID, "adminID", data.Acc.ID, "Account Keys", aki, "ScopeID", data.Scope.ID)
	return nil
}
