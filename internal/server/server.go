package server

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
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
}

func Run() error {
	conf := config.GetConfig()
	return router.Run(fmt.Sprintf("%s:%d", conf.Listen.Host, conf.Listen.Port))
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
