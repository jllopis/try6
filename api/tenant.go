package api

import (
	"encoding/json"
	"net/http"

	"github.com/jllopis/try6"
	"github.com/jllopis/try6/store"
	"github.com/labstack/echo"
)

// CreateTenant handler creates a new tenant with the data provided in the request body
func CreateTenant(sm store.Storer) echo.HandlerFunc {
	return func(ctx *echo.Context) error {
		var ctd try6.CreateTenantData
		err := json.NewDecoder(ctx.Request().Body).Decode(&ctd)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, &logMessage{Status: "error", Action: "create", Info: err.Error(), Table: "tenants"})
		}
		err = sm.CreateTenant(&ctd)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, &logMessage{Status: "error", Action: "create", Info: err.Error(), Table: "tenants"})
		}
		return ctx.JSON(http.StatusCreated, ctd)
	}
}
