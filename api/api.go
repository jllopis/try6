package api

import (
	"net/http"
	"runtime"
	"strconv"
	"time"

	"bitbucket.org/jllopis/getconf"
	"github.com/labstack/echo"
)

type logMessage struct {
	Status string `json:"status"`
	Action string `json:"action"`
	Info   string `json:"info,omitempty"`
	Table  string `json:"table,omitempty"`
	Code   string `json:"code,omitempty"`
	UID    string `json:"id,omitempty"`
}

// Time is a default service to give server time
func Time(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(strconv.FormatInt(time.Now().UTC().UnixNano(), 10)))
}

// Info provide information of the server API and status
func Info(i string) echo.HandlerFunc {
	return func(ctx *echo.Context) error {
		info := map[string]interface{}{
			"API Server":      i,
			"GetConf Version": getconf.Version(),
			"Go Version":      runtime.Version(),
			"Server Time":     time.Now().UTC().UnixNano(),
		}
		return ctx.JSON(http.StatusOK, info)
	}
}
