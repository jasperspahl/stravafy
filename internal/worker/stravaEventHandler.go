package worker

import (
	"context"
	"database/sql"
	"errors"
	"golang.org/x/oauth2"
	"stravafy/internal/clients/spotify"
	"stravafy/internal/clients/strava"
	"stravafy/internal/database"
	cfgManager "stravafy/internal/manager/config"
	"strings"
	"time"
)

const (
	AspectTypeCreate   = "create"
	AspectTypeUpdate   = "update"
	AspectTypeDelete   = "delete"
	ObjectTypeActivity = "activity"
	ObjectTypeAthlete  = "athlete"
)

var (
	ErrAlreadyProcessed     = errors.New("already processed")
	ErrNoProcessingRequired = errors.New("no processing required")
)

func HandleStravaEvent(event Callback) {
	wg.Add(1)
	go handleStravaEvent(event)
}

func TestStravaActivity(uid, activityId int64) (string, error) {
	db, err := database.NewSQLite()
	if err != nil {
		errorf(uid, "nono database: %v", err)
		return "", err
	}
	q := database.New(db.DB)
	return processStravaEvent(0, uid, activityId, q, false)
}

func handleStravaEvent(event Callback) {
	defer wg.Done()
	if event.AspectType != AspectTypeCreate {
		infof(event.EventTime, "skipping event of type %s", event.AspectType)
		return
	}
	if event.ObjectType != ObjectTypeActivity {
		infof(event.EventTime, "skipping event of type %s", event.ObjectType)
		return
	}
	infof(event.EventTime, "start processing...")
	infof(event.EventTime, "\tactivity: %d", event.ObjectId)
	db, err := database.NewSQLite()
	if err != nil {
		errorf(event.EventTime, "nono database: %v", err)
		return
	}
	q := database.New(db.DB)
	user, err := q.GetUserByStravaId(context.Background(), event.OwnerId)
	if err != nil {
		errorf(event.EventTime, "error getting user from db: %v", err)
		return
	}
	infof(event.EventTime, "\tstrava user: \"%s %s\"", user.FirstName, user.LastName)

	_, err = processStravaEvent(event.EventTime, user.ID, event.ObjectId, q, true)
	if errors.Is(err, ErrNoProcessingRequired) || errors.Is(err, ErrAlreadyProcessed) {
		infof(event.EventTime, "skipping activity %d", event.ObjectId)
		return
	}
	if err != nil {
		errorf(event.EventTime, "an error has occurred: %v", err)
		return
	}
	infof(event.EventTime, "\tactivity processing: %d", event.ObjectId)
}

func processStravaEvent(pid, uid, activityId int64, q *database.Queries, upload bool) (string, error) {
	cm := cfgManager.New(q)
	playlistConfig := cm.GetUserConfig(uid, cfgManager.Playlist)
	podcastConfig := cm.GetUserConfig(uid, cfgManager.Podcast)
	if !playlistConfig.Enabled && !podcastConfig.Enabled {
		infof(pid, "user disabled all processing")
		return "", ErrNoProcessingRequired
	}
	dbToken, err := q.GetTokenByUserId(context.Background(), uid)
	if err != nil {
		errorf(pid, "error while fetching accesstoken: %v", err)
		return "", err
	}
	token := oauth2.Token{
		AccessToken:  dbToken.AccessToken,
		RefreshToken: dbToken.RefreshToken,
		Expiry:       time.Unix(dbToken.ExpiresAt, 0),
	}
	stravaClient := strava.NewStravaClient(token)

	activity, err := stravaClient.GetActivity(activityId)
	if err != nil {
		errorf(pid, "error while fetching activity: %v", err)
		return "", err
	}

	if upload && strings.Contains(activity.Description, "Stravafy") {
		infof(pid, "already processed")
		infof(pid, "exiting...")
		return "", ErrAlreadyProcessed
	}
	startTime := activity.StartDate
	endTime := activity.StartDate.Add(time.Duration(activity.ElapsedTime) * time.Second)
	histEntries, err := q.GetHistoryEntriesBetween(context.Background(), database.GetHistoryEntriesBetweenParams{
		UserID:      uid,
		Timestamp:   startTime.UTC(),
		Timestamp_2: endTime.UTC(),
	})
	if err != nil {
		errorf(pid, "an error accourd while fetching history: %v", err)
		return "", err
	}
	firstHistoryItem, err := q.GetLastHistoryEntryBefore(context.Background(), database.GetLastHistoryEntryBeforeParams{
		UserID:    uid,
		Timestamp: startTime.UTC(),
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		errorf(pid, "an error accourd while fetching history: %v", err)
		return "", err
	} else if errors.Is(err, sql.ErrNoRows) {
		infof(pid, "no history entries found before activity start")
	} else if firstHistoryItem.IsPlaying {
		entry := database.GetHistoryEntriesBetweenRow{
			ID:                     firstHistoryItem.ID,
			Timestamp:              firstHistoryItem.Timestamp,
			IsPlaying:              true,
			CtxType:                firstHistoryItem.CtxType.String,
			CtxHref:                firstHistoryItem.CtxHref.String,
			CtxExternalUrl:         firstHistoryItem.CtxExternalUrl.String,
			CtxUri:                 firstHistoryItem.CtxUri.String,
			ItemType:               firstHistoryItem.ItemType.String,
			ItemHref:               firstHistoryItem.ItemHref.String,
			ItemExternalUrl:        firstHistoryItem.ItemExternalUrl.String,
			ItemUri:                firstHistoryItem.ItemUri.String,
			Name:                   firstHistoryItem.Name.String,
			Artists:                firstHistoryItem.Artists,
			Album:                  firstHistoryItem.Album,
			AlbumUri:               firstHistoryItem.AlbumUri,
			EpisodeDescription:     firstHistoryItem.EpisodeDescription,
			EpisodeShowName:        firstHistoryItem.EpisodeShowName,
			EpisodeShowDescription: firstHistoryItem.EpisodeShowDescription,
			EpisodeShowUri:         firstHistoryItem.EpisodeShowUri,
		}
		histEntries = append([]database.GetHistoryEntriesBetweenRow{entry}, histEntries...)
	}

	spotifyDbToken, err := q.GetSpotifyAccessToken(context.Background(), uid)
	if err != nil {
		errorf(pid, "error while fetching spotify accesstoken: %v", err)
		return "", err
	}
	spotifyToken := oauth2.Token{
		TokenType:    spotifyDbToken.TokenType,
		AccessToken:  spotifyDbToken.AccessToken,
		RefreshToken: spotifyDbToken.RefreshToken,
		Expiry:       time.Unix(spotifyDbToken.ExpiresAt, 0),
	}
	spotifyClient := spotify.NewSpotifyClient(spotifyToken)

	newDescription := generateNewDescription(pid, histEntries, playlistConfig, podcastConfig, spotifyClient.GetPlaylist)
	if newDescription == "" {
		infof(pid, "done")
		return "", nil
	}

	newDescription += "\n-- by Stravafy"
	if upload {
		updatedDescription := ""
		newestActivity, err := stravaClient.GetActivity(activityId)
		if err != nil {
			updatedDescription = activity.Description
		} else {
			updatedDescription = newestActivity.Description
		}
		updatedDescription += newDescription
		infof(pid, "updating description:\n%s", updatedDescription)

		err = stravaClient.UpdateActivityDescription(pid, updatedDescription)
		if err != nil {
			errorf(pid, "%v", err)
			return "", nil
		}
	}
	return newDescription, nil
}

func generateNewDescription(taskId int64, histEntries []database.GetHistoryEntriesBetweenRow, playlistConfig, podcastConfig cfgManager.Config, getPlaylist func(string) (*spotify.MinimalPlaylist, error)) string {

	playlists := make(map[string]string)
	podcastEpisodes := make([]int, 0)

	for i, entry := range histEntries {
		if entry.IsPlaying {
			if entry.CtxType == "playlist" {
				playlists[entry.CtxHref] = entry.CtxExternalUrl
			} else if entry.ItemType == "episode" {
				podcastEpisodes = append(podcastEpisodes, i)
			}
		}
	}
	newDescription := ""
	if len(playlists) > 0 && playlistConfig.Enabled {
		for href, _ := range playlists {
			pl, err := getPlaylist(href)
			if err != nil {
				errorf(taskId, "an error acourd while getting context playlist: %v", err)
				continue
			}
			data := map[string]string{
				"Name":  pl.Name,
				"Owner": pl.Owner.DisplayName,
			}
			desc, err := playlistConfig.Process(data)
			if err != nil {
				errorf(taskId, "an error acourd while processing playlist: %v", err)
				continue
			}

			newDescription += "\n" + desc
		}
	}
	if len(podcastEpisodes) > 0 && podcastConfig.Enabled {
		for _, index := range podcastEpisodes {
			data := map[string]string{
				"Show":     histEntries[index].EpisodeShowName.String,
				"ShowDesc": histEntries[index].EpisodeShowDescription.String,
				"Name":     histEntries[index].Name,
			}
			desc, err := podcastConfig.Process(data)
			if err != nil {
				errorf(taskId, "an error acourd while processing podcast: %v", err)
				continue
			}
			newDescription += "\n" + desc
		}
	}

	return newDescription

}
