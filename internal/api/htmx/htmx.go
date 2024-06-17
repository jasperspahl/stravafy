package htmx

import (
	"net/http"
	"stravafy/internal/database"
	"stravafy/internal/manager/config"
	"stravafy/internal/manager/playlist"
	"stravafy/internal/sessions"
	"stravafy/internal/templates"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Service struct {
	q             *database.Queries
	configManager *config.Manager
	playlistManager playlist.Manager
}

func New(q *database.Queries, configManager *config.Manager, playlistManager playlist.Manager) *Service {
	return &Service{q: q, configManager: configManager, playlistManager: playlistManager}
}

func (s *Service) Mount(group *gin.RouterGroup) {
	group.PUT("config/disable", s.ToggleUserConfig(false))
	group.PUT("config/enable", s.ToggleUserConfig(true))
	group.PUT("config", s.EditUserConfig)
	group.GET("config/view", s.GetUserConfig)
	group.GET("playlist/view", s.GetPlaylistView)
	group.GET("playlist/cards", s.GetPlaylistCards)
}

func (s *Service) GetPlaylistCards(c *gin.Context) {
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
	limitStr := c.DefaultQuery("limit", "4")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil {
		_ = c.Error(err)
		return
	}
	offset, err := strconv.ParseInt(offsetStr, 10, 64)
	if err != nil {
		_ = c.Error(err)
		return
	}
	playlists, err := s.playlistManager.GetUserPlaylists(uid, limit, offset)

	failedToFetch := 0
	if err != nil {
		if failed, ok := err.(*playlist.ErrPlaylistFailedToFetch); ok {
			failedToFetch = int(*failed)
		} else {
			_ = c.Error(err)
			return
		}
	}

	if len(playlists) == 0 {
		if offset == 0 {
			c.String(http.StatusOK, "No playlists found")
			return
		}
		c.String(http.StatusOK, "")
		return
	}
	if len(playlists) < int(limit) - failedToFetch {
		c.HTML(http.StatusOK, "", templates.PlaylistCards(playlists))
		return
	}
	c.HTML(http.StatusOK, "", templates.PlaylistCardsWithLoadMore(playlists, limit, offset+limit))

}

func (s *Service) GetPlaylistView(c *gin.Context) {
	c.HTML(http.StatusOK, "", templates.PlaylistView())
}

func (s *Service) GetUserConfig(c *gin.Context) {
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
	playlistConf := s.configManager.GetUserConfig(uid, config.Playlist)
	podcastConf := s.configManager.GetUserConfig(uid, config.Podcast)
	c.HTML(http.StatusOK, "", templates.ConfigView([]config.Config{playlistConf, podcastConf}))
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
		c.HTML(http.StatusOK, "", templates.ConfigCard(conf))
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
	c.HTML(http.StatusOK, "", templates.ConfigCard(conf))
}
