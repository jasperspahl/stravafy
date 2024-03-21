package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/oauth2"
	"io"
	"log"
	"net/http"
	"os"
	"stravafy/internal/config"
	"stravafy/internal/database"
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

	oauth2Conf := config.GetSpotifyOauthConfig()
	client := oauth2Conf.Client(context.Background(), &token)

	conf := config.GetConfig()
	ticker := time.Tick(time.Duration(conf.Spotify.UpdateInterval) * time.Second)

	for {
		select {
		case <-ticker:
			logger.Printf("worker %d [INFO]: updating player state", id)
			resp, err := client.Get("https://api.spotify.com/v1/me/player?additional_types=track,episode")
			if err != nil {
				errorf(id, "%v", err)
				continue
			}
			infof(id, "[HTTP] GET /me/player %d", resp.StatusCode)
			switch resp.StatusCode {
			case http.StatusNoContent:
				err := handlePaused(id, queries)
				if err != nil {
					errorf(id, "%v", err)
				}
			case http.StatusOK:
				err := handlePlaying(id, queries, resp)
				if err != nil {
					errorf(id, "%v", err)
				}
			}
		case <-shutdown:
			infof(id, "shutting down worker for %d", id)
			return
		}
	}

}

func handlePlaying(id int64, q *database.Queries, resp *http.Response) error {
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		errorf(id, "could not read body: %v", err)
		return err
	}
	var playerState PlayerState
	err = json.Unmarshal(bytes, &playerState)
	if err != nil {
		errorf(id, "could not serialize response: %v", err)
		return err
	}
	if !playerState.IsPlaying {
		return handlePaused(id, q)
	}
	var item ItemObject
	err = json.Unmarshal(playerState.Item, &item)
	if err != nil {
		return fmt.Errorf("could not serialize item: %v", err)
	}
	var track TrackObject
	var episode EpisodeObject
	switch item.Type {
	case "track":
		err := json.Unmarshal(playerState.Item, &track)
		if err != nil {
			return err
		}
	case "episode":
		err := json.Unmarshal(playerState.Item, &episode)
		if err != nil {
			return err
		}
	default:
		infof(id, string(bytes))
		return fmt.Errorf("looking for type \"track\" or \"episode\" found %s", item.Type)
	}
	lastHistEntry, err := q.GetLastHistoryEntryComplete(context.Background(), id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if errors.Is(err, sql.ErrNoRows) || hasChanged(id, lastHistEntry, playerState, item) {
		return insertPlayingState(id, q, playerState, item, &track, &episode)
	}
	return nil
}

func hasChanged(id int64, lastEntry database.GetLastHistoryEntryCompleteRow, state PlayerState, item ItemObject) bool {
	infof(id, "looking for changes")
	if lastEntry.IsPlaying != state.IsPlaying {
		return true
	}
	if lastEntry.CtxUri != state.Context.Uri {
		return true
	}
	if lastEntry.ItemUri != item.Uri {
		return true
	}
	infof(id, "no changes found")
	return false
}

func insertPlayingState(id int64, q *database.Queries, playerState PlayerState, item ItemObject, track *TrackObject, episode *EpisodeObject) error {
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
	infof(id, "currently not playing")
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
		return err
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
