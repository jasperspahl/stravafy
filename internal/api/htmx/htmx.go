package htmx

import (
	"fmt"
	"net/http"
	"regexp"
	"database/sql"
	"stravafy/internal/database"
	"stravafy/internal/manager/config"
	"stravafy/internal/manager/playlist"
	"stravafy/internal/sessions"
	"stravafy/internal/templates"
	"strconv"
	"strings"

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
	group.Use(ensureAuthenticatedMiddleware)
	group.PUT("config/disable", s.ToggleUserConfig(false))
	group.PUT("config/enable", s.ToggleUserConfig(true))
	group.PUT("config", s.EditUserConfig)
	group.GET("config/view", s.GetUserConfig)
	group.GET("playlist/view", s.GetPlaylistView)
	group.GET("playlist/cards", s.GetPlaylistCards)
	group.GET("items/view", s.GetItemsView)
	group.GET("items", s.GetItems)
}

func ensureAuthenticatedMiddleware(c *gin.Context) {
	session, err := sessions.GetSession(c)
	currentPath := c.GetHeader("HX-Current-URL")
	requestPath := c.Request.URL.Path
	var viewRequestRegex = regexp.MustCompile(`^/htmx/([a-z].*)/view$`)
	if viewRequestRegex.MatchString(requestPath) {
		var host = c.Request.Host
		method := "https"
		if parts := strings.Split(host, ":"); len(parts) > 1 {
			method = "http"
		}
		matches := viewRequestRegex.FindStringSubmatch(requestPath)
		currentPath = fmt.Sprintf("%s://%s/?page=%s", method, host, matches[1])
	}
	redirectURL := "/auth/login"
	if currentPath != "" {
		redirectURL += "?redirect=" + currentPath
	}
	if err != nil {
		c.Header("HX-Redirect", redirectURL)
		c.Status(http.StatusUnauthorized)
		return
	}
	_, err = session.GetUserId(c)
	if err == sessions.ErrNotLoggedIn {
		c.Header("HX-Redirect", redirectURL)
		c.Status(http.StatusUnauthorized)
		return
	}
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.Next()
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

func (s *Service) GetItemsView(c *gin.Context) {
	c.HTML(http.StatusOK, "", templates.ItemsView())
}

func (s *Service) GetItems(c *gin.Context) {
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
	limitStr := c.DefaultQuery("limit", "100")
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

	items, err := s.q.GetUserHistoryItems(c, database.GetUserHistoryItemsParams{
		UserID: uid,
		Limit:  limit,
		Offset: offset,
	})
	result := make([]templates.HistoryItem, 0, len(items))
	if err == sql.ErrNoRows {} else if err != nil {
		_ = c.Error(err)
		return
	}
	for _, item := range items {
		histItem := templates.HistoryItem{
			Time:   item.Timestamp.String(),
			Href:   item.ExternalUrl,
			Track:  item.Name,
			Artist: "",
		}
		if item.Artists.Valid {
			histItem.Artist = item.Artists.String
		}
		if item.Album.Valid {
			histItem.Album = item.Album.String
		}
		if item.EpisodeShowName.Valid {
			histItem.Album = item.EpisodeShowName.String
		}
		result = append(result, histItem)
	}
	if len(result) < int(limit) {
		c.HTML(http.StatusOK, "", templates.HistoryItems(result))
		return
	}
	c.HTML(http.StatusOK, "", templates.HistoryItemsWithLoadMore(result, limit, offset+limit))
}
