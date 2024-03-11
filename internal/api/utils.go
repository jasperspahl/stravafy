package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"stravafy/internal/sessions"
	"stravafy/internal/templates"
)

func Error(c *gin.Context, code int, err error) {
	session, err := sessions.GetSession(c)
	if err != nil {
		c.HTML(code, "", templates.Error(code, err.Error(), nil))
		return
	}
	user, err := session.GetUser(context.Background())
	if err != nil {
		c.HTML(code, "", templates.Error(code, err.Error(), nil))
	}
	c.HTML(code, "", templates.Error(code, err.Error(), &user))

}
