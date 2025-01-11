package worker

import (
	"database/sql"
	"log"
	"os"
	"stravafy/internal/clients/spotify"
	"stravafy/internal/database"
	"stravafy/internal/manager/config"
	"testing"
)

func TestGenerateNewDescriptionEmptyDescription(t *testing.T) {
	logger = log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	histEntries := make([]database.GetHistoryEntriesBetweenRow, 0)

	mockedGetPlaylist := func(href string) (*spotify.MinimalPlaylist, error) {
		return &spotify.MinimalPlaylist{
			Name: "Test",
			Owner: struct {
				DisplayName  string               `json:"display_name"`
				ExternalUrls spotify.ExternalUrls `json:"external_urls"`
			}{
				DisplayName: "Test",
				ExternalUrls: spotify.ExternalUrls{
					Spotify: "",
				},
			},
		}, nil
	}

	playlistConfig := config.Config{
		Type:     config.Playlist,
		Enabled:  true,
		Template: "{{.Name}} {{.Owner}}",
	}
	podcastConfig := config.Config{
		Type:     config.Podcast,
		Enabled:  true,
		Template: "{{.Show}} {{.Name}}",
	}

	result := generateNewDescription(0, histEntries, playlistConfig, podcastConfig, mockedGetPlaylist)

	if result != "" {
		t.Fatalf("generateNewDescription(0, [], playlistConfig, podcastConfig, mockedGetPlaylist) == \"%s\" != \"\"", result)
	}
}

func TestGenerateNewDescriptionOnlyPlaylist(t *testing.T) {
	logger = log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	histEntries := []database.GetHistoryEntriesBetweenRow{
		{
			ID:              1,
			IsPlaying:       true,
			CtxType:         "playlist",
			CtxHref:         "https://api.spotify.com/v1/playlists/6sp1gCY0lF9G1Wlo983jf0",
			CtxExternalUrl:  "https://open.spotify.com/playlist/6sp1gCY0lF9G1Wlo983jf0",
			CtxUri:          "spotify:playlist:6sp1gCY0lF9G1Wlo983jf0",
			ItemType:        "track",
			ItemHref:        "https://api.spotify.com/v1/tracks/5uu2OCGGrTRS1sIvlMgKwe",
			ItemExternalUrl: "https://open.spotify.com/track/5uu2OCGGrTRS1sIvlMgKwe",
			ItemUri:         "spotify:track:5uu2OCGGrTRS1sIvlMgKwe",
			Name:            "i am not who i was",
		},
		{
			ID:              49,
			IsPlaying:       true,
			CtxType:         "show",
			CtxHref:         "https://api.spotify.com/v1/shows/5vQTGlmbla4NhLUrdmIRph",
			CtxExternalUrl:  "https://open.spotify.com/show/5vQTGlmbla4NhLUrdmIRph",
			CtxUri:          "spotify:show:5vQTGlmbla4NhLUrdmIRph",
			ItemType:        "episode",
			ItemHref:        "https://api.spotify.com/v1/episodes/3D1KaJ4nNDzOol0ek6kaV2",
			ItemExternalUrl: "https://open.spotify.com/episode/3D1KaJ4nNDzOol0ek6kaV2",
			ItemUri:         "spotify:episode:3D1KaJ4nNDzOol0ek6kaV2",
			Name:            "PW No. 64 - Lotterleben",
			EpisodeDescription: sql.NullString{
				String: "Die erste Folge als offizieller Radsport-Rentner Podcast. Aktuell sucht Rick nicht nur einen Job sondern auch sein Fahrrad. Rick und Tanja erörtern außerdem wer der größere Feind des Radsportlers ist: das UCI Regelwerk oder der vermeintliche Spirit of Gravel. Das und vieles mehr in der neuen Folge Parallelwelten!  Partner dieser Folgen:   Athletic Greens - Auf www.drinkag1.com/planz erhältst du bei Abschluss eines monatlichen Abos einen kostenlosen Jahresvorrat Vitamin D3 & K2 und 5 praktische AG1 Travel Packs für unterwegs im Wert von über 40€ gratis dazu.  Buycycle - Für kurze Zeit kannst du jetzt 100 Euro bei deinem nächsten Fahrradkauf sparen, ganz einfach mit dem Code PLANZ. Gib gebraucht eine Chance - und finde dein nächstes Traumbike auf www.buycycle.com",
				Valid:  true,
			},
			EpisodeShowName: sql.NullString{
				String: "Plan Z",
				Valid:  true,
			},
			EpisodeShowDescription: sql.NullString{
				String: "Plan Z und Parallelwelten - der Interview Podcast mit Tanja Erath und Rick Zabel. Tanja und Rick geben einen Blick hinter die Kulissen von (Rad-)Sportler:innen und engen Vertrauten. Sie erzählen aus ihren Leben und locken persönliche und unerwartete Geschichten aus ihren Gesprächspartner:innen.  Realtalk at it’s Best, mit vielfältigen und spannenden Menschen, die vor keiner Frage sicher sind.  Anfragen: planz@rickzabel.de",
				Valid:  true,
			},
			EpisodeShowUri: sql.NullString{
				String: "spotify:show:5vQTGlmbla4NhLUrdmIRph",
				Valid:  true,
			},
		},
		{
			ID:              2,
			IsPlaying:       true,
			CtxType:         "playlist",
			CtxHref:         "https://api.spotify.com/v1/playlists/6sp1gCY0lF9G1Wlo983jf0",
			CtxExternalUrl:  "https://open.spotify.com/playlist/6sp1gCY0lF9G1Wlo983jf0",
			CtxUri:          "spotify:playlist:6sp1gCY0lF9G1Wlo983jf0",
			ItemType:        "track",
			ItemHref:        "https://api.spotify.com/v1/tracks/0wwfjg7kbwGuL3X0JdtHn6",
			ItemExternalUrl: "https://open.spotify.com/track/0wwfjg7kbwGuL3X0JdtHn6",
			ItemUri:         "spotify:track:0wwfjg7kbwGuL3X0JdtHn6",
			Name:            "Hang Me Like Jesus",
		},
	}

	playlistConfig := config.Config{
		Type:     config.Playlist,
		Enabled:  true,
		Template: "{{.Name}} {{.Owner}}",
	}
	podcastConfig := config.Config{
		Type:     config.Podcast,
		Enabled:  false,
		Template: "{{.Show}} {{.Name}}",
	}

	mockedGetPlaylist := func(href string) (*spotify.MinimalPlaylist, error) {
		if href == "https://api.spotify.com/v1/playlists/6sp1gCY0lF9G1Wlo983jf0" {
			return &spotify.MinimalPlaylist{
				Name: "wenn blätter fallen",
				Owner: struct {
					DisplayName  string               `json:"display_name"`
					ExternalUrls spotify.ExternalUrls `json:"external_urls"`
				}{
					DisplayName: "julie",
				},
			}, nil
		}
		return &spotify.MinimalPlaylist{
			Name: "Test",
			Owner: struct {
				DisplayName  string               `json:"display_name"`
				ExternalUrls spotify.ExternalUrls `json:"external_urls"`
			}{
				DisplayName: "Test",
			},
		}, nil
	}

	result := generateNewDescription(0, histEntries, playlistConfig, podcastConfig, mockedGetPlaylist)

	expected := "\nwenn blätter fallen julie https://open.spotify.com/playlist/6sp1gCY0lF9G1Wlo983jf0"

	if result != expected {
		t.Fatalf("%s != %s", result, expected)
	}
}

func TestGenerateNewDescriptionOnlyPodcast(t *testing.T) {
	logger = log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	histEntries := []database.GetHistoryEntriesBetweenRow{
		{
			ID:              1,
			IsPlaying:       true,
			CtxType:         "playlist",
			CtxHref:         "https://api.spotify.com/v1/playlists/6sp1gCY0lF9G1Wlo983jf0",
			CtxExternalUrl:  "https://open.spotify.com/playlist/6sp1gCY0lF9G1Wlo983jf0",
			CtxUri:          "spotify:playlist:6sp1gCY0lF9G1Wlo983jf0",
			ItemType:        "track",
			ItemHref:        "https://api.spotify.com/v1/tracks/5uu2OCGGrTRS1sIvlMgKwe",
			ItemExternalUrl: "https://open.spotify.com/track/5uu2OCGGrTRS1sIvlMgKwe",
			ItemUri:         "spotify:track:5uu2OCGGrTRS1sIvlMgKwe",
			Name:            "i am not who i was",
		},
		{
			ID:              49,
			IsPlaying:       true,
			CtxType:         "show",
			CtxHref:         "https://api.spotify.com/v1/shows/5vQTGlmbla4NhLUrdmIRph",
			CtxExternalUrl:  "https://open.spotify.com/show/5vQTGlmbla4NhLUrdmIRph",
			CtxUri:          "spotify:show:5vQTGlmbla4NhLUrdmIRph",
			ItemType:        "episode",
			ItemHref:        "https://api.spotify.com/v1/episodes/3D1KaJ4nNDzOol0ek6kaV2",
			ItemExternalUrl: "https://open.spotify.com/episode/3D1KaJ4nNDzOol0ek6kaV2",
			ItemUri:         "spotify:episode:3D1KaJ4nNDzOol0ek6kaV2",
			Name:            "PW No. 64 - Lotterleben",
			EpisodeDescription: sql.NullString{
				String: "Die erste Folge als offizieller Radsport-Rentner Podcast. Aktuell sucht Rick nicht nur einen Job sondern auch sein Fahrrad. Rick und Tanja erörtern außerdem wer der größere Feind des Radsportlers ist: das UCI Regelwerk oder der vermeintliche Spirit of Gravel. Das und vieles mehr in der neuen Folge Parallelwelten!  Partner dieser Folgen:   Athletic Greens - Auf www.drinkag1.com/planz erhältst du bei Abschluss eines monatlichen Abos einen kostenlosen Jahresvorrat Vitamin D3 & K2 und 5 praktische AG1 Travel Packs für unterwegs im Wert von über 40€ gratis dazu.  Buycycle - Für kurze Zeit kannst du jetzt 100 Euro bei deinem nächsten Fahrradkauf sparen, ganz einfach mit dem Code PLANZ. Gib gebraucht eine Chance - und finde dein nächstes Traumbike auf www.buycycle.com",
				Valid:  true,
			},
			EpisodeShowName: sql.NullString{
				String: "Plan Z",
				Valid:  true,
			},
			EpisodeShowDescription: sql.NullString{
				String: "Plan Z und Parallelwelten - der Interview Podcast mit Tanja Erath und Rick Zabel. Tanja und Rick geben einen Blick hinter die Kulissen von (Rad-)Sportler:innen und engen Vertrauten. Sie erzählen aus ihren Leben und locken persönliche und unerwartete Geschichten aus ihren Gesprächspartner:innen.  Realtalk at it’s Best, mit vielfältigen und spannenden Menschen, die vor keiner Frage sicher sind.  Anfragen: planz@rickzabel.de",
				Valid:  true,
			},
			EpisodeShowUri: sql.NullString{
				String: "spotify:show:5vQTGlmbla4NhLUrdmIRph",
				Valid:  true,
			},
		},
		{
			ID:              2,
			IsPlaying:       true,
			CtxType:         "playlist",
			CtxHref:         "https://api.spotify.com/v1/playlists/6sp1gCY0lF9G1Wlo983jf0",
			CtxExternalUrl:  "https://open.spotify.com/playlist/6sp1gCY0lF9G1Wlo983jf0",
			CtxUri:          "spotify:playlist:6sp1gCY0lF9G1Wlo983jf0",
			ItemType:        "track",
			ItemHref:        "https://api.spotify.com/v1/tracks/0wwfjg7kbwGuL3X0JdtHn6",
			ItemExternalUrl: "https://open.spotify.com/track/0wwfjg7kbwGuL3X0JdtHn6",
			ItemUri:         "spotify:track:0wwfjg7kbwGuL3X0JdtHn6",
			Name:            "Hang Me Like Jesus",
		},
	}

	playlistConfig := config.Config{
		Type:     config.Playlist,
		Enabled:  false,
		Template: "{{.Name}} {{.Owner}}",
	}
	podcastConfig := config.Config{
		Type:     config.Podcast,
		Enabled:  true,
		Template: "{{.Show}} {{.Name}}",
	}

	mockedGetPlaylist := func(href string) (*spotify.MinimalPlaylist, error) {
		if href == "https://api.spotify.com/v1/playlists/6sp1gCY0lF9G1Wlo983jf0" {
			return &spotify.MinimalPlaylist{
				Name: "wenn blätter fallen",
				Owner: struct {
					DisplayName  string               `json:"display_name"`
					ExternalUrls spotify.ExternalUrls `json:"external_urls"`
				}{
					DisplayName: "julie",
				},
			}, nil
		}
		return &spotify.MinimalPlaylist{
			Name: "Test",
			Owner: struct {
				DisplayName  string               `json:"display_name"`
				ExternalUrls spotify.ExternalUrls `json:"external_urls"`
			}{
				DisplayName: "Test",
			},
		}, nil
	}

	result := generateNewDescription(0, histEntries, playlistConfig, podcastConfig, mockedGetPlaylist)

	expected := "\nPlan Z PW No. 64 - Lotterleben"

	if result != expected {
		t.Fatalf("%s != %s", result, expected)
	}
}

func TestGenerateNewDescription(t *testing.T) {
	logger = log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	histEntries := []database.GetHistoryEntriesBetweenRow{
		{
			ID:              1,
			IsPlaying:       true,
			CtxType:         "playlist",
			CtxHref:         "https://api.spotify.com/v1/playlists/6sp1gCY0lF9G1Wlo983jf0",
			CtxExternalUrl:  "https://open.spotify.com/playlist/6sp1gCY0lF9G1Wlo983jf0",
			CtxUri:          "spotify:playlist:6sp1gCY0lF9G1Wlo983jf0",
			ItemType:        "track",
			ItemHref:        "https://api.spotify.com/v1/tracks/5uu2OCGGrTRS1sIvlMgKwe",
			ItemExternalUrl: "https://open.spotify.com/track/5uu2OCGGrTRS1sIvlMgKwe",
			ItemUri:         "spotify:track:5uu2OCGGrTRS1sIvlMgKwe",
			Name:            "i am not who i was",
		},
		{
			ID:              49,
			IsPlaying:       true,
			CtxType:         "show",
			CtxHref:         "https://api.spotify.com/v1/shows/5vQTGlmbla4NhLUrdmIRph",
			CtxExternalUrl:  "https://open.spotify.com/show/5vQTGlmbla4NhLUrdmIRph",
			CtxUri:          "spotify:show:5vQTGlmbla4NhLUrdmIRph",
			ItemType:        "episode",
			ItemHref:        "https://api.spotify.com/v1/episodes/3D1KaJ4nNDzOol0ek6kaV2",
			ItemExternalUrl: "https://open.spotify.com/episode/3D1KaJ4nNDzOol0ek6kaV2",
			ItemUri:         "spotify:episode:3D1KaJ4nNDzOol0ek6kaV2",
			Name:            "PW No. 64 - Lotterleben",
			EpisodeDescription: sql.NullString{
				String: "Die erste Folge als offizieller Radsport-Rentner Podcast. Aktuell sucht Rick nicht nur einen Job sondern auch sein Fahrrad. Rick und Tanja erörtern außerdem wer der größere Feind des Radsportlers ist: das UCI Regelwerk oder der vermeintliche Spirit of Gravel. Das und vieles mehr in der neuen Folge Parallelwelten!  Partner dieser Folgen:   Athletic Greens - Auf www.drinkag1.com/planz erhältst du bei Abschluss eines monatlichen Abos einen kostenlosen Jahresvorrat Vitamin D3 & K2 und 5 praktische AG1 Travel Packs für unterwegs im Wert von über 40€ gratis dazu.  Buycycle - Für kurze Zeit kannst du jetzt 100 Euro bei deinem nächsten Fahrradkauf sparen, ganz einfach mit dem Code PLANZ. Gib gebraucht eine Chance - und finde dein nächstes Traumbike auf www.buycycle.com",
				Valid:  true,
			},
			EpisodeShowName: sql.NullString{
				String: "Plan Z",
				Valid:  true,
			},
			EpisodeShowDescription: sql.NullString{
				String: "Plan Z und Parallelwelten - der Interview Podcast mit Tanja Erath und Rick Zabel. Tanja und Rick geben einen Blick hinter die Kulissen von (Rad-)Sportler:innen und engen Vertrauten. Sie erzählen aus ihren Leben und locken persönliche und unerwartete Geschichten aus ihren Gesprächspartner:innen.  Realtalk at it’s Best, mit vielfältigen und spannenden Menschen, die vor keiner Frage sicher sind.  Anfragen: planz@rickzabel.de",
				Valid:  true,
			},
			EpisodeShowUri: sql.NullString{
				String: "spotify:show:5vQTGlmbla4NhLUrdmIRph",
				Valid:  true,
			},
		},
		{
			ID:              2,
			IsPlaying:       true,
			CtxType:         "playlist",
			CtxHref:         "https://api.spotify.com/v1/playlists/6sp1gCY0lF9G1Wlo983jf0",
			CtxExternalUrl:  "https://open.spotify.com/playlist/6sp1gCY0lF9G1Wlo983jf0",
			CtxUri:          "spotify:playlist:6sp1gCY0lF9G1Wlo983jf0",
			ItemType:        "track",
			ItemHref:        "https://api.spotify.com/v1/tracks/0wwfjg7kbwGuL3X0JdtHn6",
			ItemExternalUrl: "https://open.spotify.com/track/0wwfjg7kbwGuL3X0JdtHn6",
			ItemUri:         "spotify:track:0wwfjg7kbwGuL3X0JdtHn6",
			Name:            "Hang Me Like Jesus",
		},
	}

	playlistConfig := config.Config{
		Type:     config.Playlist,
		Enabled:  true,
		Template: "{{.Name}} {{.Owner}}",
	}
	podcastConfig := config.Config{
		Type:     config.Podcast,
		Enabled:  true,
		Template: "{{.Show}} {{.Name}}",
	}

	mockedGetPlaylist := func(href string) (*spotify.MinimalPlaylist, error) {
		if href == "https://api.spotify.com/v1/playlists/6sp1gCY0lF9G1Wlo983jf0" {
			return &spotify.MinimalPlaylist{
				Name: "wenn blätter fallen",
				Owner: struct {
					DisplayName  string               `json:"display_name"`
					ExternalUrls spotify.ExternalUrls `json:"external_urls"`
				}{
					DisplayName: "julie",
				},
			}, nil
		}
		return &spotify.MinimalPlaylist{
			Name: "Test",
			Owner: struct {
				DisplayName  string               `json:"display_name"`
				ExternalUrls spotify.ExternalUrls `json:"external_urls"`
			}{
				DisplayName: "Test",
			},
		}, nil
	}

	result := generateNewDescription(0, histEntries, playlistConfig, podcastConfig, mockedGetPlaylist)

	expected := "\nwenn blätter fallen julie\nPlan Z PW No. 64 - Lotterleben"

	if result != expected {
		t.Fatalf("%s != %s", result, expected)
	}
}
