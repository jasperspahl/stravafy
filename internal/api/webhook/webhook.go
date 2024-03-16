package webhook

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"stravafy/internal/config"
	"stravafy/internal/database"
)

var logger *log.Logger

func init() {
	logfile, err := os.OpenFile("webhook.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("error opening worker.log: %v", err)
	}
	logger = log.New(logfile, "", log.LstdFlags)
}

type Service struct {
	queries *database.Queries
}

func New(queries *database.Queries) *Service {
	return &Service{
		queries: queries,
	}
}

func (s *Service) Mount(group *gin.RouterGroup) {
	group.POST("", s.webhookCallback)
	group.GET("", s.webhookValidation)
}

type SubscriptionPayload struct {
	ID int64 `json:"id"`
}

func RegisterWebhook() {
	conf := config.GetConfig()
	data := url.Values{}
	data.Add("client_id", fmt.Sprintf("%d", conf.Strava.ClientId))
	data.Add("client_secret", conf.Strava.ClientSecret)
	data.Add("callback_url", fmt.Sprintf("%s/callback", conf.Strava.WebhookHost))
	data.Add("verify_token", conf.Strava.StateString)

	logger.Printf("[INFO]: starting subscription")
	resp, err := http.PostForm("https://www.strava.com/api/v3/push_subscriptions", data)
	if err != nil {
		logger.Printf("[ERROR]: error while subscribing: %v", err)
		return
	}
	logger.Printf("[INFO]: StatusCode %d: %s", resp.StatusCode, resp.Status)
	if resp.StatusCode != http.StatusOK {
		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Fatalf("[FATAL]: could not read body: %v", err)
		}
		logger.Printf("[ERROR]: could not register webhook: %s", string(bytes))
		return
	}
	var payload SubscriptionPayload
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&payload)
	if err != nil {
		logger.Printf("[ERROR]: error while subscribing: %v", err)
		return
	}
	logger.Printf("[INFO]: subscribed with id: %d", payload.ID)
}

type Callback struct {
	ObjectType     string            `json:"object_type"`
	ObjectId       int64             `json:"object_id"`
	AspectType     string            `json:"aspect_type"`
	Updates        map[string]string `json:"updates"`
	OwnerId        int64             `json:"owner_id"`
	SubscriptionId int64             `json:"subscription_id"`
	EventTime      int64             `json:"event_time"`
}

func (s *Service) webhookCallback(c *gin.Context) {
	var args Callback
	err := c.Bind(&args)
	if err != nil {
		logger.Printf("[ERROR]: could not bind callback args: %v", err)
	}
	logger.Printf("[INFO]: %v", args)
	// TODO: handle the logic
}

type Validation struct {
	Mode        string `form:"hub.mode"`
	Challenge   string `form:"hub.challenge"`
	VerifyToken string `form:"hub.verify_token"`
}

func (s *Service) webhookValidation(c *gin.Context) {
	logger.Printf("[INFO]: reciving validation request")
	var validation Validation
	err := c.Bind(&validation)
	if err != nil {
		logger.Printf("[ERROR]: could not bind validation args: %v", err)
		c.Status(http.StatusForbidden)
		return
	}
	logger.Printf("[INFO]: %v", validation)
	if validation.Mode != "subscribe" {
		logger.Printf("[ERROR]: unexpected validation mode: %s", validation.Mode)
		c.Status(http.StatusForbidden)
		return
	}
	conf := config.GetConfig()
	if validation.VerifyToken != conf.Strava.StateString {
		logger.Printf("[ERROR]: unexpected verify token: %s", validation.VerifyToken)
		c.Status(http.StatusForbidden)
		return
	}
	c.JSON(http.StatusOK, &gin.H{
		"hub.challenge": validation.Challenge,
	})
}
