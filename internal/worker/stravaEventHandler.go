package worker

import (
	"context"
	"encoding/json"
	"errors"
	"golang.org/x/oauth2"
	"net/http"
	"stravafy/internal/clients/strava"
	"stravafy/internal/config"
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
	podcastConfig := cm.GetUserConfig(user.ID, cfgManager.Playlist)
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
	getPlaylist := func(href string) (*MinimalPlaylist, error) {
		return getPlaylist(q, user.ID, href)
	}

	newDescription := generateNewDescription(event.EventTime, histEntries, playlistConfig, podcastConfig, getPlaylist)
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

func generateNewDescription(taskId int64, histEntries []database.GetHistoryEntriesBetweenRow, playlistConfig, podcastConfig cfgManager.Config, getPlaylist func(string) (*MinimalPlaylist, error)) string {

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

type MinimalPlaylist struct {
	Name  string `json:"name"`
	Owner struct {
		DisplayName string `json:"display_name"`
	} `json:"owner"`
}

func getPlaylist(q *database.Queries, userId int64, playlistHref string) (*MinimalPlaylist, error) {
	oauth2config := config.GetSpotifyOauthConfig()
	dbToken, err := q.GetSpotifyAccessToken(context.Background(), userId)
	if err != nil {
		return nil, err
	}
	token := oauth2.Token{
		TokenType:    dbToken.TokenType,
		AccessToken:  dbToken.AccessToken,
		RefreshToken: dbToken.RefreshToken,
		Expiry:       time.Unix(dbToken.ExpiresAt, 0),
	}
	client := oauth2config.Client(context.Background(), &token)
	resp, err := client.Get(playlistHref + "?fields=name,owner.display_name")
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
