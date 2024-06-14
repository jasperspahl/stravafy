package config

import (
	"bytes"
	"context"
	"stravafy/internal/database"
	"text/template"
)

type Type int64

const (
	Podcast Type = iota
	Playlist
)

func (t Type) String() string {
	return [...]string{"Podcast", "Playlist"}[t]
}

func (t Type) Help() map[string]string {
	return [...]map[string]string{PodcastVariables, PlaylistVariables}[t]
}
func (t Type) DefaultConfig() Config {
	return DefaultConfig[t]
}

func (t Type) PreviewData() interface{} {
	return previewData[t]
}

type Config struct {
	Type     Type
	Enabled  bool
	Template string
	Preview  string
}

func (c Config) Process(data map[string]string) (string, error) {
	tmpl, err := template.New("t").Parse(c.Template)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	err = tmpl.Execute(&b, data)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

var DefaultConfig = map[Type]Config{
	Podcast: {
		Type:     Podcast,
		Enabled:  true,
		Template: "Podcast: {{.Show}}\nEpisode: {{.Name}}",
	},
	Playlist: {
		Type:     Playlist,
		Enabled:  true,
		Template: "Playlist: {{.Name}} by {{.Owner}}\n{{.Url}}",
	},
}

type PodcastData struct {
	Name     string
	Url      string
	Show     string
	ShowDesc string
}

var PodcastVariables = map[string]string{
	"Name":    "name of the episode",
	"Url":     "url of the episode",
	"Show":    "name of the podcast",
	"ShowUrl": "url of the podcast",
}

type PlaylistData struct {
	Name  string
	Owner string
	Url   string
}

var PlaylistVariables = map[string]string{
	"Name":  "name of the playlist",
	"Owner": "owner of the playlist",
	"Url":   "url of the playlist",
}

var previewData = map[Type]interface{}{
	Podcast: PodcastData{
		Name:     "PW No. 60 - Der Mullet muss weg",
		Url:      "https://open.spotify.com/episode/0rOub7dlXasbeFP0vPFZ7x?si=KwDt_Q13QuK7KcWVJWoLag",
		Show:     "Plan Z",
		ShowDesc: "Plan Z und Parallelwelten - der Interview Podcast mit Tanja Erath und Rick Zabel.",
	},
	Playlist: PlaylistData{
		Name:  "Corrupted Blood",
		Owner: "HandOfBlood",
		Url:   "https://open.spotify.com/playlist/6WlSmKm4leMtyyhqqaDKAb?si=0bc052c74a694e2a",
	},
}

type Manager struct {
	q *database.Queries
}

func New(q *database.Queries) *Manager {
	return &Manager{q}
}

func (m *Manager) GetUserConfig(uid int64, t Type) Config {
	c, err := m.q.GetUserConfig(context.Background(), database.GetUserConfigParams{
		UserID: uid,
		Type:   int64(t),
	})
	var config Config
	if err != nil {
		config = Config{
			Type:     DefaultConfig[t].Type,
			Enabled:  DefaultConfig[t].Enabled,
			Template: DefaultConfig[t].Template,
		}
		err = nil
	} else {
		config = Config{
			Type:     t,
			Enabled:  c.Enabled == 1,
			Template: c.Template,
		}
	}
	tmpl, err := template.New("t").Parse(config.Template)
	if err != nil {
		config.Preview = ""
		return config
	}
	var b bytes.Buffer
	err = tmpl.Execute(&b, previewData[t])
	if err != nil {
		config.Preview = ""
		return config
	}
	config.Preview = b.String()
	return config
}

func (m *Manager) UpdateUserConfigTemplate(uid int64, t Type, template string) error {
	return m.q.UpdateUserConfigTemplate(context.Background(), database.UpdateUserConfigTemplateParams{
		Template: template,
		UserID:   uid,
		Type:     int64(t),
	})
}

func (m *Manager) UpdateUserConfigEnabled(uid int64, t Type, enabled bool) (err error) {
	if enabled {
		err = m.q.UpdateUserConfigEnabled(context.Background(), database.UpdateUserConfigEnabledParams{
			UserID:  uid,
			Type:    int64(t),
			Enabled: 1,
		})
	} else {
		err = m.q.UpdateUserConfigEnabled(context.Background(), database.UpdateUserConfigEnabledParams{
			UserID:  uid,
			Type:    int64(t),
			Enabled: 0,
		})
	}
	return
}

func (m *Manager) EnsureUserConfigExists(uid int64, t Type) error {
	_, err := m.q.GetUserConfig(context.Background(), database.GetUserConfigParams{
		UserID: uid,
		Type:   int64(t),
	})
	if err == nil {
		return nil
	}
	err = m.q.InsertUserConfig(context.Background(), database.InsertUserConfigParams{
		UserID:   uid,
		Type:     int64(t),
		Enabled:  1,
		Template: t.DefaultConfig().Template,
	})
	return err
}
