package worker

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"os"
	"stravafy/internal/config"
	"stravafy/internal/database"
	"stravafy/internal/clients/spotify"
	"strings"
	"sync"
	"time"
)

var (
	logger     *log.Logger
	shutdownCh chan struct{}
	wg         sync.WaitGroup
)

func init() {
	logfile, err := os.OpenFile("worker.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("error opening worker.log: %v", err)
	}
	logger = log.New(logfile, "", log.LstdFlags)
	shutdownCh = make(chan struct{})
}

func Start() {
	db, err := database.NewSQLite()
	if err != nil {
		logger.Fatalf("nono database: %v", err)
	}
	queries := database.New(db.DB)
	userIds, err := queries.GetUserIdsWithActiveSpotify(context.Background())
	if err != nil {
		logger.Printf("worker error: %v", err)
		return
	}
	wg.Add(len(userIds))
	for _, id := range userIds {
		go worker(id, shutdownCh, &wg)
	}

}

func LaunchSyncForUser(userID int64) {
	wg.Add(1)
	go worker(userID, shutdownCh, &wg)
}

func worker(id int64, shutdown <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	infof(id, "started worker for %d", id)

	db, err := database.NewSQLite()
	if err != nil {
		errorf(id, "nono database: %v", err)
		return
	}
	queries := database.New(db.DB)
	dbToken, err := queries.GetSpotifyAccessToken(context.Background(), id)
	token := oauth2.Token{
		AccessToken:  dbToken.AccessToken,
		TokenType:    dbToken.TokenType,
		RefreshToken: dbToken.RefreshToken,
		Expiry:       time.Unix(dbToken.ExpiresAt, 0),
	}

	client := spotify.NewSpotifyClient(token)

	conf := config.GetConfig()
	ticker := time.Tick(time.Duration(conf.Spotify.UpdateInterval) * time.Second)

	for {
		select {
		case <-ticker:
			statusCode, playerState, item, track, episode, err := client.GetPlayerState()
			if err != nil {
				errorf(id, "%v", err)
				continue
			}
			if statusCode == http.StatusNoContent || playerState != nil && !playerState.IsPlaying {
				err := handlePaused(id, queries)
				if err != nil {
					errorf(id, "%v", err)
				}
				continue
			}
			err = handlePlaying(id, queries, playerState, item, track, episode)
			if err != nil {
				errorf(id, "%v", err)
			}
		case <-shutdown:
			infof(id, "shutting down worker for %d", id)
			return
		}
	}

}

func handlePlaying(id int64, q *database.Queries, playerState *spotify.PlayerState, item *spotify.ItemObject, track *spotify.TrackObject, episode *spotify.EpisodeObject) error {
	lastHistEntry, err := q.GetLastHistoryEntryComplete(context.Background(), id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if errors.Is(err, sql.ErrNoRows) || hasChanged(id, lastHistEntry, playerState, item) {
		return insertPlayingState(id, q, playerState, item, track, episode)
	}
	return nil
}

func hasChanged(id int64, lastEntry database.GetLastHistoryEntryCompleteRow, state *spotify.PlayerState, item *spotify.ItemObject) bool {
	if lastEntry.IsPlaying != state.IsPlaying {
		return true
	}
	if state.Context != nil && lastEntry.CtxUri != state.Context.Uri {
		return true
	}
	if lastEntry.ItemUri != item.Uri {
		return true
	}
	return false
}

func insertPlayingState(id int64, q *database.Queries, playerState *spotify.PlayerState, item *spotify.ItemObject, track *spotify.TrackObject, episode *spotify.EpisodeObject) error {
	infof(id, "inserting new player state")
	histId, err := q.InsertHistory(context.Background(), database.InsertHistoryParams{
		UserID:    id,
		Timestamp: time.UnixMilli(playerState.Timestamp).UTC(),
		IsPlaying: true,
	})
	if err != nil {
		return err
	}
	if playerState.Context != nil {
		err := q.InsertHistoryContext(context.Background(), database.InsertHistoryContextParams{
			HistoryID:   histId,
			Type:        playerState.Context.Type,
			Href:        playerState.Context.Href,
			ExternalUrl: playerState.Context.ExternalUrls.Spotify,
			Uri:         playerState.Context.Uri,
		})
		if err != nil {
			return err
		}
	}
	params := database.InsertHistoryItemParams{
		HistoryID:   histId,
		Type:        item.Type,
		Href:        item.Href,
		ExternalUrl: item.ExternalUrls.Spotify,
		Uri:         item.Uri,
		Name:        item.Name,
	}
	if item.Type == "track" {
		var artists []string
		for _, artist := range track.Artists {
			artists = append(artists, artist.Name)
		}
		params.Artists = sql.NullString{
			String: strings.Join(artists, ", "),
			Valid:  true,
		}
		params.Album = sql.NullString{
			String: track.Album.Name,
			Valid:  true,
		}
		params.AlbumUri = sql.NullString{
			String: track.Album.Uri,
			Valid:  true,
		}
	} else {
		params.EpisodeDescription = sql.NullString{
			String: episode.Description,
			Valid:  true,
		}
		params.EpisodeShowName = sql.NullString{
			String: episode.Show.Name,
			Valid:  true,
		}
		params.EpisodeShowDescription = sql.NullString{
			String: episode.Show.Description,
			Valid:  true,
		}
		params.EpisodeShowUri = sql.NullString{
			String: episode.Show.Uri,
			Valid:  true,
		}
	}
	infof(id, "new history entry id: %d", histId)
	return q.InsertHistoryItem(context.Background(), params)
}

func handlePaused(id int64, q *database.Queries) error {
	lastHistEntry, err := q.GetLastHistoryEntryForUser(context.Background(), id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if errors.Is(err, sql.ErrNoRows) || lastHistEntry.IsPlaying {
		_, err := q.InsertHistory(context.Background(), database.InsertHistoryParams{
			UserID:    id,
			Timestamp: time.Now().UTC(),
			IsPlaying: false,
		})
		if err != nil {
			return err
		}
		infof(id, "[spotify] inserted paused state")
	}
	return nil
}

func Shutdown() {
	close(shutdownCh)
	wg.Wait()
}

func infof(id int64, format string, v ...any) {
	logger.Printf("worker %d [INFO]: %s", id, fmt.Sprintf(format, v...))
}
func errorf(id int64, format string, v ...any) {
	logger.Printf("worker %d [ERROR]: %s", id, fmt.Sprintf(format, v...))
}
