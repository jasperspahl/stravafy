package worker

import (
	"encoding/json"
	"time"
)

type Callback struct {
	ObjectType     string            `json:"object_type"`
	ObjectId       int64             `json:"object_id"`
	AspectType     string            `json:"aspect_type"`
	Updates        map[string]string `json:"updates"`
	OwnerId        int64             `json:"owner_id"`
	SubscriptionId int64             `json:"subscription_id"`
	EventTime      int64             `json:"event_time"`
}

type ExternalUrls struct {
	Spotify string `json:"spotify"`
}

type PlayerContext struct {
	Type         string       `json:"type"`
	Href         string       `json:"href"`
	ExternalUrls ExternalUrls `json:"external_urls"`
	Uri          string       `json:"uri"`
}

type PlayerState struct {
	Timestamp            int64           `json:"timestamp"`
	IsPlaying            bool            `json:"is_playing"`
	CurrentlyPlayingType string          `json:"currently_playing_type"`
	Context              *PlayerContext  `json:"context"`
	Item                 json.RawMessage `json:"item"`
}

type Artist struct {
	Href         string       `json:"href"`
	Id           string       `json:"id"`
	Name         string       `json:"name"`
	Type         string       `json:"type"`
	Uri          string       `json:"uri"`
	ExternalUrls ExternalUrls `json:"external_urls"`
}

type AlbumObject struct {
	AlbumType    string       `json:"album_type;required"`
	ExternalUrls ExternalUrls `json:"external_urls"`
	Href         string       `json:"href"`
	Id           string       `json:"id"`
	Name         string       `json:"name"`
	Type         string       `json:"type"`
	Uri          string       `json:"uri"`
	Artists      []Artist     `json:"artists"`
}

type ShowObject struct {
	Description string `json:"description"`
	Href        string `json:"href"`
	Id          string `json:"id"`
	Name        string `json:"name"`
	Uri         string `json:"uri"`
}

type TrackObject struct {
	ItemObject
	Album   AlbumObject `json:"album"`
	Artists []Artist    `json:"artists"`
}

type EpisodeObject struct {
	ItemObject
	Description string     `json:"description"`
	Show        ShowObject `json:"show"`
}

type ItemObject struct {
	ExternalUrls ExternalUrls `json:"external_urls"`
	Href         string       `json:"href"`
	Id           string       `json:"id"`
	Name         string       `json:"name"`
	Type         string       `json:"type"`
	Uri          string       `json:"uri"`
}

type MetaAthlete struct {
	ID int64 `json:"id"`
}

type LatLng [2]float64

type PolylineMap struct {
	ID              string `json:"id"`
	Polyline        string `json:"polyline"`
	SummaryPolyline string `json:"summary_polyline"`
}
type PhotosSummaryPrimary struct {
	ID       int64             `json:"id"`
	Source   int               `json:"source"`
	UniqueID string            `json:"unique_id"`
	Urls     map[string]string `json:"urls"`
}
type PhotosSummary struct {
	Count           int                  `json:"count"`
	Primary         PhotosSummaryPrimary `json:"primary"`
	UsePrimaryPhoto bool                 `json:"use_primary_photo"`
}

type ResourceState int

const (
	Summary ResourceState = iota + 2
	Detail
)

type SummaryGear struct {
	ID            string        `json:"id"`
	ResourceState ResourceState `json:"resource_state"`
	Primary       bool          `json:"primary"`
	Name          string        `json:"name"`
	Distance      float64       `json:"distance"`
}

type Lap struct {
	ID                 int64        `json:"id"`
	Activity           MetaActivity `json:"activity"`
	Athlete            MetaAthlete  `json:"athlete"`
	AverageCadence     float64      `json:"average_cadence"`
	AverageSpeed       float64      `json:"average_speed"`
	Distance           float64      `json:"distance"`
	ElapsedTime        int          `json:"elapsed_time"`
	StartIndex         int          `json:"start_index"`
	EndIndex           int          `json:"end_index"`
	LapIndex           int          `json:"lap_index"`
	MaxSpeed           float64      `json:"max_speed"`
	MovingTime         int          `json:"moving_time"`
	Name               string       `json:"name"`
	PaceZone           int          `json:"pace_zone"`
	Split              int          `json:"split"`
	StartDate          time.Time    `json:"start_date"`
	StartDateLocal     time.Time    `json:"start_date_local"`
	TotalElevationGain float64      `json:"total_elevation_gain"`
}

type MetaActivity struct {
	ID int64 `json:"id"`
}

type DetailedActivity struct {
	ID                   int64         `json:"id"`
	ExternalID           string        `json:"external_id"`
	UploadID             string        `json:"upload_id"`
	Athlete              MetaAthlete   `json:"athlete"`
	Name                 string        `json:"name"`
	Distance             float64       `json:"distance"`
	MovingTime           int           `json:"moving_time"`
	ElapsedTime          int           `json:"elapsed_time"`
	TotalElevationGain   float64       `json:"total_elevation_gain"`
	ElevationHigh        float64       `json:"elev_high"`
	ElevationLow         float64       `json:"elev_low"`
	Type                 string        `json:"type"`
	SportType            string        `json:"sport_type"`
	StartDate            time.Time     `json:"start_date"`
	StartDateLocal       time.Time     `json:"start_date_local"`
	Timezone             string        `json:"timezone"`
	StartLatLng          LatLng        `json:"start_latlng"`
	EndLatLng            LatLng        `json:"end_latlng"`
	AchievementCount     int           `json:"achievement_count"`
	KudosCount           int           `json:"kudos_count"`
	CommentCount         int           `json:"comment_count"`
	AthleteCount         int           `json:"athlete_count"`
	PhotoCount           int           `json:"photo_count"`
	TotalPhotoCount      int           `json:"total_photo_count"`
	Map                  PolylineMap   `json:"map"`
	Trainer              bool          `json:"trainer"`
	Commute              bool          `json:"commute"`
	Manual               bool          `json:"manual"`
	Private              bool          `json:"private"`
	Flagged              bool          `json:"flagged"`
	WorkoutType          int           `json:"workout_type"`
	UploadIdString       string        `json:"upload_id_str"`
	AverageSpeed         float64       `json:"average_speed"`
	MaxSpeed             float64       `json:"max_speed"`
	HasKudos             bool          `json:"has_kudos"`
	HideFromHome         bool          `json:"hide_from_home"`
	GearID               string        `json:"gear_id"`
	Kilojoules           float64       `json:"kilojoules"`
	AverageWatts         float64       `json:"average_watts"`
	DeviceWatts          bool          `json:"device_watts"`
	MaxWatts             int           `json:"max_watts"`
	WeightedAverageWatts int           `json:"weighted_average_watts"`
	Description          string        `json:"description"`
	Photos               PhotosSummary `json:"photos"`
	Gear                 SummaryGear   `json:"gear"`
	DeviceName           string        `json:"device_name"`
	EmbedToken           string        `json:"embed_token"`
	Laps                 []Lap         `json:"laps"`
}
