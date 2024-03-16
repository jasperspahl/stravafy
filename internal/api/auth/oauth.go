package auth

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"net/http"
	"stravafy/internal/config"
	"stravafy/internal/database"
	"stravafy/internal/sessions"
)

type Service struct {
	queries            *database.Queries
	stravaOauthConfig  oauth2.Config
	spotifyOauthConfig oauth2.Config
}

func New(queries *database.Queries) *Service {
	return &Service{
		queries:            queries,
		stravaOauthConfig:  config.GetStravaOauthConfig(),
		spotifyOauthConfig: config.GetSpotifyOauthConfig(),
	}
}

func (s *Service) Mount(group *gin.RouterGroup) {
	group.GET("/login", s.login)
	group.GET("/login/spotify", s.loginSpotify)
	group.GET("/logout", s.logout)
	group.GET("/strava/callback", s.stravaCallback)
	group.GET("/spotify/callback", s.spotifyCallback)
}

func (s *Service) logout(c *gin.Context) {
	ses, err := sessions.GetSession(c)
	if err != nil {
		return
	}
	ses.Logout(c)
	c.Redirect(http.StatusSeeOther, "/")
}
