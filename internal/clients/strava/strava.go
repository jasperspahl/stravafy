package strava

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"io"
	"net/http"
	"net/url"
	"stravafy/internal/config"
)

type Client struct {
	httpClient *http.Client
}

func NewStravaClient(token oauth2.Token) *Client {
	oauth2Conf := config.GetStravaOauthConfig()
	return &Client{
		httpClient: oauth2Conf.Client(context.Background(), &token),
	}
}

func (c *Client) GetActivity(activityID int64) (*DetailedActivity, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("https://www.strava.com/api/v3/activities/%d", activityID))
	if err != nil {
		return nil, fmt.Errorf("an error accured while fetching activity details: %v", err)
	}
	if resp.StatusCode > 299 {
		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("an error accured while reading activity details: %v", err)
		}
		return nil, fmt.Errorf("activity details returned with HTTP %d %s: %s", resp.StatusCode, resp.Status, string(bytes))
	}
	decoder := json.NewDecoder(resp.Body)
	var activity DetailedActivity
	err = decoder.Decode(&activity)
	if err != nil {
		return nil, fmt.Errorf("unable to decode activity: %v", err)
	}
	return &activity, nil
}

func (c *Client) UpdateActivityDescription(activityID int64, description string) error {
	values := make(url.Values)
	values.Add("description", description)
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("https://www.strava.com/api/v3/activities/%d?%s", activityID, values.Encode()), nil)
	if err != nil {
		return err
	}
	r, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	if r.StatusCode > 299 {
		bytes, err := io.ReadAll(r.Body)
		if err != nil {
			return fmt.Errorf("an error accured while updating activity details: %v", err)
		}
		return fmt.Errorf("updating activity returned with HTTP %d %s: %s", r.StatusCode, r.Status, string(bytes))
	}
	return nil
}
