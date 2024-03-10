package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
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
	router.Use(sessions.Middleware(queries))
	router.Static("/assets", "./assets")
	pagesService.Mount(router.Group("/"))
	authService.Mount(router.Group("/auth"))
}

func Run() error {
	conf := config.GetConfig()
	return router.Run(fmt.Sprintf("%s:%d", conf.Listen.Host, conf.Listen.Port))
}
