package auth

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
	"net/http"
	"stravafy/internal/config"
	"stravafy/internal/database"
	"stravafy/internal/sessions"
)

type Provider int

type Service struct {
	queries            *database.Queries
	stravaOauthConfig  oauth2.Config
	spotifyOauthConfig oauth2.Config
}

func New(queries *database.Queries) *Service {
	conf := config.GetConfig()
	return &Service{queries, oauth2.Config{
		ClientID:     fmt.Sprintf("%d", conf.Strava.ClientId),
		ClientSecret: conf.Strava.ClientSecret,
		Scopes:       []string{"read,activity:read_all,activity:write"},
		Endpoint:     endpoints.Strava,
	}, oauth2.Config{
		ClientID:     conf.Spotify.ClientID,
		ClientSecret: conf.Spotify.ClientSecret,
		Scopes:       []string{"user-read-currently-playing", "user-read-playback-state"},
		Endpoint:     endpoints.Spotify,
	}}
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
