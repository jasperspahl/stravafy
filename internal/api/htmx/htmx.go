package htmx

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"stravafy/internal/database"
	"stravafy/internal/manager/config"
	"stravafy/internal/sessions"
	"stravafy/internal/templates"
)

type Service struct {
	q             *database.Queries
	configManager *config.Manager
}

func New(q *database.Queries, configManager *config.Manager) *Service {
	return &Service{q: q, configManager: configManager}
}

func (s *Service) Mount(group *gin.RouterGroup) {
	group.PUT("config/disable", s.ToggleUserConfig(false))
	group.PUT("config/enable", s.ToggleUserConfig(true))
	group.PUT("config", s.EditUserConfig)
}

type ConfigType struct {
	Type config.Type `form:"type"`
}

func (s *Service) ToggleUserConfig(enable bool) func(*gin.Context) {
	return func(c *gin.Context) {
		var cfg ConfigType
		err := c.Bind(&cfg)
		if err != nil {
			_ = c.Error(err)
			return
		}
		session, err := sessions.GetSession(c)
		if err != nil {
			_ = c.Error(err)
			return
		}
		uid, err := session.GetUserId(c)
		if err != nil {
			_ = c.Error(err)
			return
		}
		err = s.configManager.EnsureUserConfigExists(uid, cfg.Type)
		if err != nil {
			_ = c.Error(err)
			return
		}
		err = s.configManager.UpdateUserConfigEnabled(uid, cfg.Type, enable)
		if err != nil {
			_ = c.Error(err)
			return
		}
		conf := s.configManager.GetUserConfig(uid, cfg.Type)
		c.HTML(http.StatusOK, "", templates.ConfigCard(cfg.Type, conf))
	}
}

type EditConfig struct {
	Type     config.Type `form:"type"`
	Template string      `form:"template"`
}

func (s *Service) EditUserConfig(c *gin.Context) {
	var cfg EditConfig
	err := c.Bind(&cfg)
	if err != nil {
		_ = c.Error(err)
		return
	}
	session, err := sessions.GetSession(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	uid, err := session.GetUserId(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	err = s.configManager.EnsureUserConfigExists(uid, cfg.Type)
	if err != nil {
		_ = c.Error(err)
		return
	}
	err = s.configManager.UpdateUserConfigTemplate(uid, cfg.Type, cfg.Template)
	if err != nil {
		_ = c.Error(err)
		return
	}
	conf := s.configManager.GetUserConfig(uid, cfg.Type)
	c.HTML(http.StatusOK, "", templates.ConfigCard(cfg.Type, conf))
}
