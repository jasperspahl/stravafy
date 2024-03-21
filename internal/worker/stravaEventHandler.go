package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/oauth2"
	"io"
	"net/http"
	"net/url"
	"stravafy/internal/config"
	"stravafy/internal/database"
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
	dbToken, err := q.GetTokenByUserId(context.Background(), user.ID)
	if err != nil {
		errorf(event.EventTime, "error while fetching accesstoken: %v", err)
	}
	token := oauth2.Token{
		AccessToken:  dbToken.AccessToken,
		RefreshToken: dbToken.RefreshToken,
		Expiry:       time.Unix(dbToken.ExpiresAt, 0),
	}

	oauth2Conf := config.GetStravaOauthConfig()
	client := oauth2Conf.Client(context.Background(), &token)

	resp, err := client.Get(fmt.Sprintf("https://www.strava.com/api/v3/activities/%d", event.ObjectId))
	if err != nil {
		errorf(event.EventTime, "an error accured while fetching activity details: %v", err)
		return
	}
	if resp.StatusCode > 299 {
		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			errorf(event.EventTime, "an error accured while reading activity details: %v", err)
			return
		}
		errorf(event.EventTime, "activity details returned with HTTP %d %s: %s", resp.StatusCode, resp.Status, string(bytes))
		return
	}
	decoder := json.NewDecoder(resp.Body)
	var activity DetailedActivity
	err = decoder.Decode(&activity)
	if err != nil {
		errorf(event.EventTime, "unable to decode activity: %v", err)
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
	infof(event.EventTime, "Found following Spotify Activity:")
	playContexts := make(map[string]struct {
		Type string
		Href string
		Url  string
	})
	for _, entry := range histEntries {
		if entry.IsPlaying {
			infof(event.EventTime, "\t Name: %s", entry.Name)
			infof(event.EventTime, "\t Artists: %s", entry.Artists.String)
			infof(event.EventTime, "")
			playContexts[entry.CtxUri] = struct {
				Type string
				Href string
				Url  string
			}{Type: entry.CtxType, Href: entry.CtxHref, Url: entry.CtxExternalUrl}
		}
	}
	newDescription := activity.Description
	for uri, ctx := range playContexts {
		infof(event.EventTime, "Contexts:")
		infof(event.EventTime, "\t Type: %s", ctx.Type)
		infof(event.EventTime, "\t Uri: %s", uri)
		infof(event.EventTime, "\t Url: %s", ctx.Url)
		infof(event.EventTime, "\t Href: %s", ctx.Href)
		if len(playContexts) == 1 && ctx.Type == "playlist" {
			pl, err := getPlaylist(q, user.ID, ctx.Href)
			if err != nil {
				errorf(event.EventTime, "an error acourd while getting context playlist: %v", err)
				continue
			}
			newDescription += fmt.Sprintf("Playlist: %s\nBy: %s\n%s\n\n--stravafy.servebeer.com", pl.Name, pl.Owner.DisplayName, ctx.Url)
		}
	}
	if newDescription == activity.Description {
		infof(event.EventTime, "done")
		return
	}
	infof(event.EventTime, "updating description:\n%s", newDescription)

	values := make(url.Values)
	values.Add("description", newDescription)
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("https://www.strava.com/api/v3/activities/%d?%s", event.ObjectId, values.Encode()), nil)
	if err != nil {
		errorf(event.EventTime, "%v", err)
		return
	}
	r, err := client.Do(req)
	if err != nil {
		errorf(event.EventTime, "%v", err)
		return
	}
	if r.StatusCode > 299 {
		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			errorf(event.EventTime, "an error accured while updating activity details: %v", err)
			return
		}
		errorf(event.EventTime, "updating activity returned with HTTP %d %s: %s", r.StatusCode, r.Status, string(bytes))
		return
	}
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
