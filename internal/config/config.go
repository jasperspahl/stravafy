package config

import (
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
	"log"
	"os"
	"strings"
)

type DatabaseConfig struct {
	Source string
}

type StravaConfig struct {
	ClientId       int
	ClientSecret   string
	ApprovalPrompt string
	StateString    string
}

type SpotifyConfig struct {
	ClientID       string
	ClientSecret   string
	UpdateInterval int
	ShowDialog     bool
}

type ListenConfig struct {
	Host string
	Port int
}

type Config struct {
	Listen   ListenConfig
	Strava   StravaConfig
	Spotify  SpotifyConfig
	Database DatabaseConfig
}

type OnConfigChangeFunc func(event fsnotify.Event, config *Config, oldConfig *Config)

var conf *Config
var onConfigChangeFuncs []OnConfigChangeFunc

func DefaultConfig() *Config {
	return &Config{
		Listen: ListenConfig{
			Host: "localhost",
			Port: 80,
		},
		Strava: StravaConfig{
			ClientId:       123456,
			ClientSecret:   "<client-secret>",
			ApprovalPrompt: "auto",
			StateString:    "stravafy",
		},
		Spotify: SpotifyConfig{
			ClientID:       "<client-id>",
			ClientSecret:   "<client-secret>",
			UpdateInterval: 60,
			ShowDialog:     false,
		},
		Database: DatabaseConfig{
			Source: "file:stravafy.db?mode=rwc",
		},
	}
}

func Setup(configPath string) error {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("STRAVAFY")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.SetConfigType("yml")
	viper.SetConfigName("config")
	viper.SetConfigPermissions(0600)
	viper.AddConfigPath(configPath)

	viper.SetDefault("listen", DefaultConfig().Listen)
	viper.SetDefault("database", DefaultConfig().Database)

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return err
		}

		err := os.MkdirAll(configPath, os.ModePerm)
		if err != nil {
			return err
		}
		viper.SetConfigFile(fmt.Sprintf("%s/config.yml", configPath))
		viper.SetDefault("strava", DefaultConfig().Strava)
		viper.SetDefault("spotify", DefaultConfig().Spotify)
		if err := viper.WriteConfig(); err != nil {
			return err
		}
	}

	viper.OnConfigChange(handleConfigChange)
	viper.WatchConfig()

	return viper.Unmarshal(&conf)
}

func GetConfig() *Config {
	return conf
}

func GetSpotifyOauthConfig() oauth2.Config {
	return oauth2.Config{
		ClientID:     conf.Spotify.ClientID,
		ClientSecret: conf.Spotify.ClientSecret,
		Scopes:       []string{"user-read-currently-playing", "user-read-playback-state"},
		Endpoint:     endpoints.Spotify,
	}
}

func GetStravaOauthConfig() oauth2.Config {
	return oauth2.Config{
		ClientID:     fmt.Sprintf("%d", conf.Strava.ClientId),
		ClientSecret: conf.Strava.ClientSecret,
		Scopes:       []string{"read,activity:read_all,activity:write"},
		Endpoint:     endpoints.Strava,
	}
}

func handleConfigChange(e fsnotify.Event) {
	log.Println("Config changed")
	oldConfig := *conf
	conf = nil
	_ = viper.Unmarshal(&conf)
	for _, callbackFunc := range onConfigChangeFuncs {
		callbackFunc(e, conf, &oldConfig)
	}
}
