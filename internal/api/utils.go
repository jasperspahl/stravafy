package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"runtime/debug"
	"stravafy/internal/sessions"
	"stravafy/internal/templates"
)

var Commit = func() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value
			}
		}
	}

	return ""
}()

func BuildInfo(c *gin.Context) {
	c.String(http.StatusOK, Commit)
}

func Error(c *gin.Context, code int, err error) {
	session, e := sessions.GetSession(c)
	if e != nil {
		c.HTML(code, "", templates.Error(code, err.Error(), true))
		return
	}
	userid, e := session.GetUserId(c)
	if e != nil {
		c.HTML(code, "", templates.Error(code, err.Error(), false))
		return
	}
	c.HTML(code, "", templates.Error(code, err.Error(), userid != 0))
}
