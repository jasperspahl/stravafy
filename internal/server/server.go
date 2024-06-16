package server

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"net/http"
	"stravafy/internal/api"
	"stravafy/internal/api/auth"
	"stravafy/internal/api/htmx"
	"stravafy/internal/api/pages"
	"stravafy/internal/api/webhook"
	"stravafy/internal/config"
	"stravafy/internal/database"
	cfgManager "stravafy/internal/manager/config"
	"stravafy/internal/manager/playlist"
	"stravafy/internal/renderer"
	"stravafy/internal/sessions"

	"github.com/gin-gonic/gin"
)

var (
	router *gin.Engine
	srv    *http.Server
)

//go:embed assets/*
var assets embed.FS

func Init(queries *database.Queries) {
	configManager := cfgManager.New(queries)
	playlistManager := playlist.New(queries)

	pagesService := pages.New(queries)
	authService := auth.New(queries)
	webhookService := webhook.New(queries)
	htmxService := htmx.New(queries, configManager, playlistManager)

	router = gin.Default()
	router.HTMLRender = renderer.Default
	router.Use(ErrorHandler())
	router.Use(sessions.Middleware(queries))
	router.StaticFS("/static", http.FS(assets))
	pagesService.Mount(router.Group("/"))
	authService.Mount(router.Group("/auth"))
	webhookService.Mount(router.Group("/callback"))
	htmxService.Mount(router.Group("/htmx"))

	conf := config.GetConfig()

	srv = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", conf.Listen.Host, conf.Listen.Port),
		Handler: router,
	}
}

func Run() error {
	log.Println("starting server ...")
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func Shutdown(ctx context.Context) error {
	log.Println("shutting down server")
	return srv.Shutdown(ctx)
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		err := c.Errors.Last()
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrNotAuthorized),
				errors.Is(err, auth.ErrMissingRequiredScopes),
				errors.Is(err, auth.ErrTokenExchangeFailed):
				api.Error(c, http.StatusUnauthorized, err)
			case errors.Is(err, auth.ErrBindingOauth2Callback),
				errors.Is(err, auth.ErrStateNotSetCorrectly):
				api.Error(c, http.StatusBadRequest, err)
			default:
				api.Error(c, http.StatusInternalServerError, err)
			}
			return
		}
	}
}
