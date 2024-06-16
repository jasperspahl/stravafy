package worker

import (
	"context"
	"golang.org/x/oauth2"
	"stravafy/internal/clients/strava"
	"stravafy/internal/clients/spotify"
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

func HandleStravaEvent(event Callback) {
	wg.Add(1)
	go handleStravaEvent(event)
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
	cm := cfgManager.New(q)
	playlistConfig := cm.GetUserConfig(user.ID, cfgManager.Playlist)
	podcastConfig := cm.GetUserConfig(user.ID, cfgManager.Podcast)
	if !playlistConfig.Enabled && !podcastConfig.Enabled {
		infof(event.EventTime, "user disabled all processing")
	}
	dbToken, err := q.GetTokenByUserId(context.Background(), user.ID)
	if err != nil {
		errorf(event.EventTime, "error while fetching accesstoken: %v", err)
	}
	token := oauth2.Token{
		AccessToken:  dbToken.AccessToken,
		RefreshToken: dbToken.RefreshToken,
		Expiry:       time.Unix(dbToken.ExpiresAt, 0),
	}
	stravaClient := strava.NewStravaClient(token)

	activity, err := stravaClient.GetActivity(event.ObjectId)
	if err != nil {
		errorf(event.EventTime, "error while fetching activity: %v", err)
		return
	}

	if strings.Contains(activity.Description, "stravafy.servebeer.com") {
		infof(event.EventTime, "already processed")
		infof(event.EventTime, "exiting...")
		return
	}
	startTime := activity.StartDate
	endTime := activity.StartDate.Add(time.Duration(activity.ElapsedTime) * time.Second)
	histEntries, err := q.GetHistoryEntriesBetween(context.Background(), database.GetHistoryEntriesBetweenParams{
		UserID:      user.ID,
		Timestamp:   startTime.UTC(),
		Timestamp_2: endTime.UTC(),
	})
	if err != nil {
		errorf(event.EventTime, "an error accourd while fetching history: %v", err)
		return
	}
	spotifyDbToken, err := q.GetSpotifyAccessToken(context.Background(), user.ID)
	if err != nil {
		errorf(event.EventTime, "error while fetching spotify accesstoken: %v", err)
		return
	}
	spotifyToken := oauth2.Token{
		TokenType:    spotifyDbToken.TokenType,
		AccessToken:  spotifyDbToken.AccessToken,
		RefreshToken: spotifyDbToken.RefreshToken,
		Expiry:       time.Unix(spotifyDbToken.ExpiresAt, 0),
	}
	spotifyClient := spotify.NewSpotifyClient(spotifyToken)

	newDescription := generateNewDescription(event.EventTime, histEntries, playlistConfig, podcastConfig, spotifyClient.GetPlaylist)
	if newDescription == "" {
		infof(event.EventTime, "done")
		return
	}

	updatedDescription := ""
	newestActivity, err := stravaClient.GetActivity(event.ObjectId)
	if err != nil {
		updatedDescription = activity.Description
	} else {
		updatedDescription = newestActivity.Description
	}
	updatedDescription += newDescription + "\n-- stravafy.servebeer.com"

	infof(event.EventTime, "updating description:\n%s", updatedDescription)

	err = stravaClient.UpdateActivityDescription(event.ObjectId, updatedDescription)
	if err != nil {
		errorf(event.EventTime, "%v", err)
		return
	}
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
		for href, url := range playlists {
			pl, err := getPlaylist(href)
			if err != nil {
				errorf(taskId, "an error acourd while getting context playlist: %v", err)
				continue
			}
			data := map[string]string{
				"Name":  pl.Name,
				"Owner": pl.Owner.DisplayName,
				"Url":   url,
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
				"Show":    histEntries[index].EpisodeShowName.String,
				"ShowUrl": histEntries[index].CtxExternalUrl,
				"Name":    histEntries[index].Name,
				"Url":     histEntries[index].ItemExternalUrl,
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

