package api

import (
	"github.com/gin-gonic/gin"
	"log"
	"stravafy/internal/sessions"
	"stravafy/internal/templates"
)

func Error(c *gin.Context, code int, err error) {
	session, err := sessions.GetSession(c)
	if err != nil {
		c.HTML(code, "", templates.Error(code, err.Error(), nil))
		return
	}
	user := session.GetUser(c)
	log.Print(user)
	c.HTML(code, "", templates.Error(code, err.Error(), user))

}
