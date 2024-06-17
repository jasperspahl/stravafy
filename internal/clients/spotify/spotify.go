package spotify

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"stravafy/internal/config"

	"golang.org/x/oauth2"
)


type Client struct {
	httpClient *http.Client
}

func NewSpotifyClient(token oauth2.Token) *Client {
	oauth2Config := config.GetSpotifyOauthConfig()
	return &Client{
		httpClient: oauth2Config.Client(context.Background(), &token),
	}
}

func (c *Client) GetPlaylist(href string) (*MinimalPlaylist, error) {
	resp, err := c.httpClient.Get(href + "?fields=name,owner(display_name,external_urls.spotify)")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("playlist not found")
	}
	decoder := json.NewDecoder(resp.Body)
	var pl MinimalPlaylist
	err = decoder.Decode(&pl)
	if err != nil {
		return nil, err
	}
	return &pl, nil
}

func (c *Client) GetPlaylistWithImages(href string) (*PlaylistWithImages, error) {
	resp, err := c.httpClient.Get(href + "?fields=name,owner(display_name,external_urls.spotify),external_urls.spotify,images")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("playlist not found exited with status " + resp.Status)
	}
	decoder := json.NewDecoder(resp.Body)
	var pl PlaylistWithImages
	err = decoder.Decode(&pl)
	if err != nil {
		return nil, err
	}
	return &pl, nil
}

func (c *Client) GetPlayerState() (int, *PlayerState, *ItemObject, *TrackObject, *EpisodeObject, error) {
	resp, err := c.httpClient.Get("https://api.spotify.com/v1/me/player?additional_types=track,episode")
	if err != nil {
		return -1, nil, nil, nil, nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil, nil, nil, nil, err
	}
	decoder := json.NewDecoder(resp.Body)

	var ps PlayerState

	err = decoder.Decode(&ps)
	if err != nil {
		return resp.StatusCode, nil, nil, nil, nil, err
	}
	if !ps.IsPlaying {
		return resp.StatusCode, &ps, nil, nil, nil, nil
	}
	var item ItemObject
	err = json.Unmarshal(ps.Item, &item)
	if err != nil {
		return resp.StatusCode, &ps, nil, nil, nil, err
	}
	switch item.Type {
	case "track":
		var track TrackObject
		err = json.Unmarshal(ps.Item, &track)
		if err != nil {
			return resp.StatusCode, &ps, nil, nil, nil, err
		}
		return resp.StatusCode, &ps, &item, &track, nil, nil
	case "episode":
		var episode EpisodeObject
		err = json.Unmarshal(ps.Item, &episode)
		if err != nil {
			return resp.StatusCode, &ps, nil, nil, nil, err
		}
		return resp.StatusCode, &ps, &item, nil, &episode, nil
	default:
		return resp.StatusCode, &ps, nil, nil, nil, errors.New("unknown item type")
	}
}
