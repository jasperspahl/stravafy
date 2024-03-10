package pages

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"stravafy/internal/api"
	"stravafy/internal/database"
	"stravafy/internal/sessions"
	"stravafy/internal/templates"
)

type Service struct {
	q *database.Queries
}

func New(q *database.Queries) *Service {
	return &Service{q}
}

func (s *Service) Mount(group *gin.RouterGroup) {
	group.GET("/", s.index)
}

func (s *Service) index(c *gin.Context) {
	session, err := sessions.GetSession(c)
	if err != nil {
		api.Error(c, http.StatusInternalServerError, err)
		return
	}

	userID, err := session.GetUserId(c)
	if err == nil {
		user, err := s.q.GetUserById(c, userID)
		if err != nil {
			api.Error(c, http.StatusInternalServerError, err)
			return
		}
		c.HTML(http.StatusOK, "", templates.Index(&user))
		return
	}
	c.HTML(http.StatusOK, "", templates.Index(nil))

}
