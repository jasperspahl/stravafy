package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"stravafy/internal/api"
	"stravafy/internal/api/auth"
	"stravafy/internal/api/pages"
	"stravafy/internal/config"
	"stravafy/internal/database"
	"stravafy/internal/renderer"
	"stravafy/internal/sessions"
)

var (
	router *gin.Engine
	srv    *http.Server
)

func Init(queries *database.Queries) {
	pagesService := pages.New(queries)
	authService := auth.New(queries)

	router = gin.Default()
	router.HTMLRender = renderer.Default
	router.Use(ErrorHandler())
	router.Use(sessions.Middleware(queries))
	router.Static("/assets", "./assets")
	pagesService.Mount(router.Group("/"))
	authService.Mount(router.Group("/auth"))

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
