package main

import (
	"context"
	"fmt"
	"golang.org/x/oauth2"
	"os"
	"stravafy/internal/clients/strava"
	"stravafy/internal/config"
	"stravafy/internal/database"
	"time"
)

func main() {
	configPath, isSet := os.LookupEnv("STRAVAFY_CONFIG_PATH")
	if !isSet {
		_, _ = fmt.Fprintln(os.Stderr, "STRAVAFY_CONFIG_PATH environment variable not set")
		os.Exit(1)
		return
	}
	fmt.Printf("loading config from %s\n", configPath)
	if err := config.Setup(configPath); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "config couldn't load, %v", err)
		os.Exit(1)
		return
	}
	fmt.Printf("config loaded from %s\n", configPath)
	fmt.Printf("Connecting to database...\n")
	db, err := database.NewSQLite()
	if err != nil {

	}
	q := database.New(db.DB)
	fmt.Printf("Connected to database\n")
	fmt.Printf("Getting Strava Client ...\n")
	dbToken, err := q.GetTokenByUserId(context.Background(), 1)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
		return
	}
	token := oauth2.Token{
		AccessToken:  dbToken.AccessToken,
		RefreshToken: dbToken.RefreshToken,
		Expiry:       time.Unix(dbToken.ExpiresAt, 0),
	}
	stravaClient := strava.NewStravaClient(token)
	fmt.Printf("Retrievied Strava Client\n")
	fmt.Printf("Updating Strava Activity ...")
	err = stravaClient.UpdateActivityDescription(11987119418, "Test")
	if err != nil {
		_, _ = fmt.Println(err.Error())
		os.Exit(1)
		return
	}
	fmt.Printf("updated strava activity description for user %d\n", 1)

}
