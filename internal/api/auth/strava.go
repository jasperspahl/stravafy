package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"net/http"
	"stravafy/internal/config"
	"stravafy/internal/database"
	"stravafy/internal/sessions"
	"strings"
)

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

type OauthCallback struct {
	Scope string `form:"scope"`
	Code  string `form:"code"`
	State string `form:"state"`
}

func (s *Service) stravaCallback(c *gin.Context) {
	errorString := c.Query("error")
	if errorString != "" {
		_ = c.Error(ErrNotAuthorized)
		return
	}
	var attr OauthCallback
	conf := config.GetConfig()
	err := c.Bind(&attr)
	if err != nil {
		_ = c.Error(ErrBindingOauth2Callback)
		return
	}
	if attr.State != conf.Strava.StateString {
		_ = c.Error(ErrStateNotSetCorrectly)
		return
	}
	if !strings.Contains(attr.Scope, "activity:write") || !strings.Contains(attr.Scope, "activity:read") {
		_ = c.Error(ErrMissingRequiredScopes)
		return
	}
	token, err := s.stravaOauthConfig.Exchange(c, attr.Code)
	if err != nil {
		_ = c.Error(ErrTokenExchangeFailed)
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
				_ = c.Error(err)
				return
			}
			err = s.queries.InsertStravaAccessToken(c, database.InsertStravaAccessTokenParams{
				UserID:      userId,
				AccessToken: token.AccessToken,
				ExpiresAt:   token.Expiry.Unix(),
			})
			if err != nil {
				_ = c.Error(err)
				return
			}
			err = s.queries.InsertStravaRefreshToken(c, database.InsertStravaRefreshTokenParams{
				UserID:       userId,
				RefreshToken: token.RefreshToken,
			})
			if err != nil {
				_ = c.Error(err)
				return
			}
		} else {
			_ = c.Error(err)
			return
		}
	} else {
		err = s.queries.UpdateStravaAccessToken(c, database.UpdateStravaAccessTokenParams{
			UserID:      userId,
			AccessToken: token.AccessToken,
			ExpiresAt:   token.Expiry.Unix(),
		})
		if err != nil {
			_ = c.Error(err)
			return
		}
		err = s.queries.UpdateStravaRefreshToken(c, database.UpdateStravaRefreshTokenParams{
			UserID:       userId,
			RefreshToken: token.RefreshToken,
		})
		if err != nil {
			_ = c.Error(err)
			return
		}
	}
	session, err := sessions.GetSession(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	err = session.SetUserId(c, userId)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.Redirect(http.StatusSeeOther, "/")
}
