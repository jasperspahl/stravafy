package auth

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"stravafy/internal/api"
	"stravafy/internal/config"
	"strings"
)

func (s *Service) loginSpotify(c *gin.Context) {
	host := c.Request.Host
	conf := config.GetConfig()
	method := "https"
	if parts := strings.Split(host, ":"); len(parts) > 1 {
		method = "http"
	}
	s.spotifyOauthConfig.RedirectURL = fmt.Sprintf("%s://%s/auth/spotify/callback", method, host)
	url := s.spotifyOauthConfig.AuthCodeURL(conf.Strava.StateString)
	c.Redirect(http.StatusSeeOther, url)
}

func (s *Service) spotifyCallback(c *gin.Context) {
	errorString := c.Query("error")
	if errorString != "" {
		api.Error(c, http.StatusUnauthorized, ErrNotAuthorized)
		return
	}
	var attr OauthCallback
	err := c.Bind(&attr)
	if err != nil {
		api.Error(c, http.StatusBadRequest, err)
		return
	}
	log.Println(attr)
	for _, scope := range s.spotifyOauthConfig.Scopes {
		if !strings.Contains(attr.Scope, scope) {
			api.Error(c, http.StatusUnauthorized, errors.New("missing required scope(s)"))
			return
		}
	}
	token, err := s.stravaOauthConfig.Exchange(c, attr.Code)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/")
		return
	}
	log.Println(token)

}
