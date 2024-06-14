package pages

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"stravafy/internal/database"
	"stravafy/internal/manager/config"
	"stravafy/internal/sessions"
	"stravafy/internal/templates"
)

type Service struct {
	q             *database.Queries
	configManager *config.Manager
}

func New(q *database.Queries) *Service {
	mng := config.New(q)
	return &Service{q, mng}
}

func (s *Service) Mount(group *gin.RouterGroup) {
	group.GET("/", s.index)
}

func (s *Service) index(c *gin.Context) {
	session, err := sessions.GetSession(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	userID, err := session.GetUserId(c)
	if err != nil {
		c.HTML(http.StatusOK, "", templates.Index())
		return
	}
	user, err := s.q.GetUserById(c, userID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	props := templates.IndexAuthenticatedProps{
		StravaID:      user.StravaID,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		StravaProfile: user.Profile,
	}
	spotifyUserInfo, err := s.q.GetSpotifyUserInfo(c, userID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		_ = c.Error(err)
	}
	if errors.Is(err, sql.ErrNoRows) {
		props.SpotifyConnected = false
	} else {
		props.SpotifyConnected = true
		props.SpotifyUserName = spotifyUserInfo.DisplayName
		props.SpotifyID = spotifyUserInfo.SpotifyID
	}
	podcastConf := s.configManager.GetUserConfig(userID, config.Podcast)
	playlistConf := s.configManager.GetUserConfig(userID, config.Playlist)
	c.HTML(http.StatusOK, "", templates.IndexAuthenticated(props, podcastConf, playlistConf))
}
