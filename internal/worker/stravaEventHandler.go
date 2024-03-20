package worker

import (
	"context"
	"fmt"
	"golang.org/x/oauth2"
	"io"
	"stravafy/internal/config"
	"stravafy/internal/database"
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

}
