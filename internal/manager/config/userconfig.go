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

type Config struct {
	Type     Type
	Enabled  bool
	Template string
	Preview  string
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
			Preview:  DefaultConfig[t].Preview,
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
