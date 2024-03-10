package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
	"log"
	"net/http"
	"stravafy/internal/api"
	"stravafy/internal/config"
	"stravafy/internal/database"
	"stravafy/internal/sessions"
	"stravafy/internal/templates"
	"strings"
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
		Scopes:       []string{},
		Endpoint:     endpoints.Spotify,
	}}
}

func (s *Service) Mount(group *gin.RouterGroup) {
	group.GET("/login", s.login)
	group.GET("/login/spotify", s.loginSpotify)
	group.GET("/logout", s.logout)
	group.GET("/strava/callback", s.stravaCallback)
}

func (s *Service) login(c *gin.Context) {
	host := c.Request.Host
	conf := config.GetConfig()
	method := "https"
	if parts := strings.Split(host, ":"); len(parts) > 1 {
		method = "http"
	}
	s.stravaOauthConfig.RedirectURL = fmt.Sprintf("%s://%s/auth/strava/callback", method, host)
	url := s.stravaOauthConfig.AuthCodeURL(conf.Strava.StateString, oauth2.SetAuthURLParam("approval_prompt", conf.Strava.ApprovalPrompt))
	c.Redirect(http.StatusSeeOther, url)
}

func (s *Service) loginSpotify(c *gin.Context) {
	host := c.Request.Host
	conf := config.GetConfig()
	method := "https"
	if parts := strings.Split(host, ":"); len(parts) > 1 {
		method = "http"
	}
	s.spotifyOauthConfig.RedirectURL = fmt.Sprintf("%s://%s/auth/spotify/callback", method, host)
	//s.spotifyOauthConfig.RedirectURL = "https://adventofcycing.ddns.net/auth/spotify/callback"
	url := s.spotifyOauthConfig.AuthCodeURL(conf.Strava.StateString)
	c.Redirect(http.StatusSeeOther, url)
}

type StravaCallback struct {
	Scope string `form:"scope"`
	Code  string `form:"code"`
	State string `form:"state"`
}

func (s *Service) stravaCallback(c *gin.Context) {
	errorString := c.Query("error")
	if errorString != "" {
		api.Error(c, http.StatusUnauthorized, errors.New("not authorized"))
		return
	}
	var attr StravaCallback
	conf := config.GetConfig()
	err := c.Bind(&attr)
	if err != nil {
		api.Error(c, http.StatusBadRequest, err)
		return
	}
	if attr.State != conf.Strava.StateString {
		api.Error(c, http.StatusBadRequest, errors.New("state not set correctly"))
		return
	}
	if !strings.Contains(attr.Scope, "activity:write") || !strings.Contains(attr.Scope, "activity:read") {
		api.Error(c, http.StatusUnauthorized, errors.New("missing activity:write or activity:read scope"))
		return
	}
	log.Print(attr)
	token, err := s.stravaOauthConfig.Exchange(c, attr.Code)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/")
		return
	}
	athlete := token.Extra("athlete").(map[string]interface{})
	stravaID := int64(athlete["id"].(float64))
	userId, err := s.queries.GetUserIdByStravaId(c, stravaID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			userId, err = s.queries.InsertUser(c, database.InsertUserParams{
				StravaID:      stravaID,
				FirstName:     athlete["firstname"].(string),
				LastName:      athlete["lastname"].(string),
				Profile:       athlete["profile"].(string),
				ProfileMedium: athlete["profile_medium"].(string),
			})
			if err != nil {
				api.Error(c, http.StatusInternalServerError, err)
				return
			}
			err = s.queries.InsertStravaAccessToken(c, database.InsertStravaAccessTokenParams{
				UserID:      userId,
				AccessToken: token.AccessToken,
				ExpiresAt:   token.Expiry.Unix(),
			})
			if err != nil {
				api.Error(c, http.StatusInternalServerError, err)
				return
			}
			err = s.queries.InsertStravaRefreshToken(c, database.InsertStravaRefreshTokenParams{
				UserID:       userId,
				RefreshToken: token.RefreshToken,
			})
			if err != nil {
				api.Error(c, http.StatusInternalServerError, err)
				return
			}
		} else {
			api.Error(c, http.StatusInternalServerError, err)
			return
		}
	} else {
		err = s.queries.UpdateStravaAccessToken(c, database.UpdateStravaAccessTokenParams{
			UserID:      userId,
			AccessToken: token.AccessToken,
			ExpiresAt:   token.Expiry.Unix(),
		})
		if err != nil {
			api.Error(c, http.StatusInternalServerError, err)
			return
		}
		err = s.queries.UpdateStravaRefreshToken(c, database.UpdateStravaRefreshTokenParams{
			UserID:       userId,
			RefreshToken: token.RefreshToken,
		})
		if err != nil {
			api.Error(c, http.StatusInternalServerError, err)
			return
		}
	}
	session, err := sessions.GetSession(c)
	if err != nil {
		api.Error(c, http.StatusInternalServerError, err)
	}
	err = session.SetUserId(c, userId)
	if err != nil {
		api.Error(c, http.StatusInternalServerError, err)
	}
	c.HTML(http.StatusOK, "", templates.AfterStrava(&database.User{
		StravaID:      stravaID,
		FirstName:     athlete["firstname"].(string),
		LastName:      athlete["lastname"].(string),
		Profile:       athlete["profile"].(string),
		ProfileMedium: athlete["profile_medium"].(string),
	}))

}

func (s *Service) logout(c *gin.Context) {
	ses, err := sessions.GetSession(c)
	if err != nil {
		return
	}
	ses.Logout(c)
	c.Redirect(http.StatusSeeOther, "/")
}
