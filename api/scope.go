package api

import (
	"encoding/json"
	"net/http"

	"github.com/jllopis/try6"
	"github.com/jllopis/try6/store"
	"github.com/labstack/echo"
)

//	apisrv.Get("/tenants/:id/scopes", api.GetScopesByTenantID(storeManager))

// CreateScope handler creates a new scope for the tenant specified in the body
func CreateScope(sm store.Storer) echo.HandlerFunc {
	return func(ctx *echo.Context) error {
		var s try6.Scope
		err := json.NewDecoder(ctx.Request().Body).Decode(&s)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, &logMessage{Status: "error", Action: "create", Info: err.Error(), Table: "scopes"})
		}
		if s.TenantID == "" {
			return ctx.JSON(http.StatusBadRequest, &logMessage{Status: "error", Action: "create", Info: "tenant not specified", Table: "scopes"})
		}
		err = sm.SaveScope(&s)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, &logMessage{Status: "error", Action: "create", Info: err.Error(), Table: "scopes"})
		}
		return ctx.JSON(http.StatusCreated, s)
	}
}

// GetScopesByTenantID returns a list of scopes owned by the tenant
func GetScopesByTenantID(sm store.Storer) echo.HandlerFunc {
	return func(ctx *echo.Context) error {
		var tenantID string
		if tenantID = ctx.Param("id"); tenantID == "" {
			return ctx.JSON(http.StatusBadRequest, &logMessage{Status: "error", Action: "GetScopesByTenantID", Info: "tenant id cannot be nil"})
		}
		scopes, err := sm.GetScopesByTenantID(tenantID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, &logMessage{Status: "error", Action: "GetScopesByTenantID", Info: err.Error(), Table: "scopes"})
		}
		return ctx.JSON(http.StatusOK, scopes)
	}
}
