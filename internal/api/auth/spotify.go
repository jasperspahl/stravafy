package auth

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"stravafy/internal/config"
	"stravafy/internal/database"
	"stravafy/internal/sessions"
	"stravafy/internal/worker"
	"strconv"
	"strings"
)

func (s *Service) loginSpotify(c *gin.Context) {
	session, err := sessions.GetSession(c)
	if err != nil {
		_ = c.Error(err)
	}
	host := c.Request.Host
	method := "https"
	if parts := strings.Split(host, ":"); len(parts) > 1 {
		method = "http"
	}
	s.spotifyOauthConfig.RedirectURL = fmt.Sprintf("%s://%s/auth/spotify/callback", method, host)
	conf := config.GetConfig()
	url := s.spotifyOauthConfig.AuthCodeURL(session.GetSessionID(), oauth2.SetAuthURLParam("show_dialog", strconv.FormatBool(conf.Spotify.ShowDialog)))
	c.Redirect(http.StatusSeeOther, url)
}

type SpotifyImageObject struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}
type SpotifyMe struct {
	ID          string               `json:"id"`
	DisplayName string               `json:"display_name"`
	Images      []SpotifyImageObject `json:"images"`
}

func (s *Service) spotifyCallback(c *gin.Context) {
	errorString := c.Query("error")
	if errorString != "" {
		_ = c.Error(ErrNotAuthorized)
		return
	}
	session, err := sessions.GetSession(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	var attr OauthCallback
	err = c.Bind(&attr)
	if err != nil {
		_ = c.Error(ErrBindingOauth2Callback)
		return
	}
	if attr.State != session.GetSessionID() {
		_ = c.Error(ErrStateNotSetCorrectly)
		return
	}
	token, err := s.spotifyOauthConfig.Exchange(c, attr.Code)
	if err != nil {
		_ = c.Error(ErrTokenExchangeFailed)
		return
	}
	scopes := token.Extra("scope")
	for _, scope := range s.spotifyOauthConfig.Scopes {
		log.Printf("Checking for scope \"%s\" in \"%s\"", scope, scopes)
		if !strings.Contains(scopes.(string), scope) {
			_ = c.Error(ErrMissingRequiredScopes)
			return
		}
	}
	userId, err := session.GetUserId(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	_, err = s.queries.GetSpotifyAccessToken(c, userId)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		_ = c.Error(err)
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		err = s.queries.InsertSpotifyAccessToken(c, database.InsertSpotifyAccessTokenParams{
			UserID:      userId,
			AccessToken: token.AccessToken,
			TokenType:   token.TokenType,
			ExpiresAt:   token.Expiry.Unix(),
		})
		if err != nil {
			_ = c.Error(err)
			return
		}
		err = s.queries.InsertSpotifyRefreshToken(c, database.InsertSpotifyRefreshTokenParams{
			UserID:       userId,
			RefreshToken: token.RefreshToken,
		})
		if err != nil {
			_ = c.Error(err)
			return
		}
	} else {
		err = s.queries.UpdateSpotifyAccessToken(c, database.UpdateSpotifyAccessTokenParams{
			AccessToken: token.AccessToken,
			ExpiresAt:   token.Expiry.Unix(),
			UserID:      userId,
		})
		if err != nil {
			_ = c.Error(err)
			return
		}
		err = s.queries.UpdateSpotifyRefreshToken(c, database.UpdateSpotifyRefreshTokenParams{
			RefreshToken: token.RefreshToken,
			UserID:       userId,
		})
		if err != nil {
			_ = c.Error(err)
			return
		}
	}
	client := s.spotifyOauthConfig.Client(c, token)
	resp, err := client.Get("https://api.spotify.com/v1/me")
	if err != nil {
		_ = c.Error(err)
		return
	}
	d := json.NewDecoder(resp.Body)
	var spotifyInfo SpotifyMe
	err = d.Decode(&spotifyInfo)
	if err != nil {
		_ = c.Error(err)
		return
	}
	err = s.queries.InsertSpotifyUserInfo(c, database.InsertSpotifyUserInfoParams{
		UserID:      userId,
		SpotifyID:   spotifyInfo.ID,
		DisplayName: spotifyInfo.DisplayName,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}
	for _, img := range spotifyInfo.Images {
		err := s.queries.InsertSpotifyUserImage(c, database.InsertSpotifyUserImageParams{
			UserID: userId,
			Url:    img.URL,
			Width:  int64(img.Width),
			Height: int64(img.Height),
		})
		if err != nil {
			_ = c.Error(err)
			return
		}
	}
	worker.LaunchSyncForUser(userId)
	c.Redirect(http.StatusSeeOther, "/")
}
