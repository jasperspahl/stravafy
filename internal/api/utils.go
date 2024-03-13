package api

import (
	"github.com/gin-gonic/gin"
	"stravafy/internal/sessions"
	"stravafy/internal/templates"
)

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
