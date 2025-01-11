package playlist

import (
	"context"
	"fmt"
	"sort"
	"stravafy/internal/clients/spotify"
	"stravafy/internal/database"
	"time"

	"golang.org/x/oauth2"
)

type Manager interface {
	GetUserPlaylists(userID int64, limit, offset int64) ([]Playlist, error)
}

type Playlist struct {
	Name  string
	Owner struct {
		Name string
		Url  string
	}
	Url   string
	Image string
}

type ErrPlaylistFailedToFetch int

func (e *ErrPlaylistFailedToFetch) Error() string {
	return fmt.Sprintf("failed to fetch %d playlists", *e)
}

type manager struct {
	q     *database.Queries
	cache map[string]cachedPlaylist
}

type cachedPlaylist struct {
	Playlist
	ExpiresAt time.Time
}

func New(q *database.Queries) Manager {
	return &manager{q: q, cache: make(map[string]cachedPlaylist)}
}

func (m *manager) GetUserPlaylists(userID int64, limit, offset int64) ([]Playlist, error) {
	playlists, err := m.q.GetUserPlaylists(context.Background(), database.GetUserPlaylistsParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	if len(playlists) == 0 {
		return make([]Playlist, 0), nil
	}
	dbToken, err := m.q.GetSpotifyAccessToken(context.Background(), userID)
	if err != nil {
		return nil, err
	}
	token := oauth2.Token{
		AccessToken:  dbToken.AccessToken,
		TokenType:    dbToken.TokenType,
		RefreshToken: dbToken.RefreshToken,
		Expiry:       time.Unix(dbToken.ExpiresAt, 0),
	}
	spotifyClient := spotify.NewSpotifyClient(token, func() {
		// TODO: Implement cleanup
	})
	var result []Playlist
	var failedToFetch ErrPlaylistFailedToFetch = 0
	for _, p := range playlists {
		cached, ok := m.cache[p.Href]
		if ok && cached.ExpiresAt.After(time.Now()) {
			result = append(result, cached.Playlist)
			continue
		}
		pl, err := spotifyClient.GetPlaylistWithImages(p.Href)
		if err != nil {
			failedToFetch++
			continue
		}
		image := ""
		if len(pl.Images) == 0 {
			image = "No image available"
		} else if len(pl.Images) == 1 {
			image = pl.Images[0].Url
		} else {
			sort.Slice(pl.Images, func(i, j int) bool {
				if pl.Images[i].Height == nil {
					return false
				}
				if pl.Images[j].Height == nil {
					return true
				}
				return *pl.Images[i].Height > *pl.Images[j].Height
			})
			image = pl.Images[0].Url
		}
		playlist := Playlist{
			Name: pl.Name,
			Owner: struct {
				Name string
				Url  string
			}{Name: pl.Owner.DisplayName, Url: pl.Owner.ExternalUrls.Spotify},
			Url:   pl.ExternalUrls.Spotify,
			Image: image,
		}
		m.cache[p.Href] = cachedPlaylist{
			Playlist:  playlist,
			ExpiresAt: time.Now().Add(6 * time.Hour),
		}
		result = append(result, playlist)
	}
	if failedToFetch > 0 {
		return result, &failedToFetch
	}
	return result, nil
}
