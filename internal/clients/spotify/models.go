package spotify

import (
	"encoding/json"
)

type MinimalPlaylist struct {
	Name  string `json:"name"`
	Owner struct {
		DisplayName string `json:"display_name"`
	} `json:"owner"`
}

type ExternalUrls struct {
	Spotify string `json:"spotify"`
}

type PlayerContext struct {
	Type         string       `json:"type"`
	Href         string       `json:"href"`
	ExternalUrls ExternalUrls `json:"external_urls"`
	Uri          string       `json:"uri"`
}

type PlayerState struct {
	Timestamp            int64           `json:"timestamp"`
	IsPlaying            bool            `json:"is_playing"`
	CurrentlyPlayingType string          `json:"currently_playing_type"`
	Context              *PlayerContext  `json:"context"`
	Item                 json.RawMessage `json:"item"`
}

type Artist struct {
	Href         string       `json:"href"`
	Id           string       `json:"id"`
	Name         string       `json:"name"`
	Type         string       `json:"type"`
	Uri          string       `json:"uri"`
	ExternalUrls ExternalUrls `json:"external_urls"`
}

type AlbumObject struct {
	AlbumType    string       `json:"album_type;required"`
	ExternalUrls ExternalUrls `json:"external_urls"`
	Href         string       `json:"href"`
	Id           string       `json:"id"`
	Name         string       `json:"name"`
	Type         string       `json:"type"`
	Uri          string       `json:"uri"`
	Artists      []Artist     `json:"artists"`
}

type ShowObject struct {
	Description string `json:"description"`
	Href        string `json:"href"`
	Id          string `json:"id"`
	Name        string `json:"name"`
	Uri         string `json:"uri"`
}

type TrackObject struct {
	ItemObject
	Album   AlbumObject `json:"album"`
	Artists []Artist    `json:"artists"`
}

type EpisodeObject struct {
	ItemObject
	Description string     `json:"description"`
	Show        ShowObject `json:"show"`
}

type ItemObject struct {
	ExternalUrls ExternalUrls `json:"external_urls"`
	Href         string       `json:"href"`
	Id           string       `json:"id"`
	Name         string       `json:"name"`
	Type         string       `json:"type"`
	Uri          string       `json:"uri"`
}
