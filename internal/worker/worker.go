package worker

import (
	"context"
	"log"
	"os"
	"stravafy/internal/config"
	"stravafy/internal/database"
	"sync"
	"time"
)

var (
	logger     *log.Logger
	shutdownCh chan struct{}
	wg         sync.WaitGroup
)

func init() {
	logfile, err := os.OpenFile("worker.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("error opening worker.log: %v", err)
	}
	logger = log.New(logfile, "", log.LstdFlags)

	shutdownCh = make(chan struct{})
}

func Start() {
	db, err := database.NewSQLite()
	if err != nil {
		logger.Fatalf("nono database: %v", err)
	}
	queries := database.New(db.DB)
	userIds, err := queries.GetUserIdsWithActiveSpotify(context.Background())
	if err != nil {
		logger.Printf("worker error: %v", err)
		return
	}
	wg.Add(len(userIds))
	for _, id := range userIds {
		go worker(id, shutdownCh, &wg)
	}

}

func LaunchSyncForUser(userID int64) {
	wg.Add(1)
	go worker(userID, shutdownCh, &wg)
}

func worker(id int64, shutdown <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	logger.Printf("started worker for %d", id)

	conf := config.GetConfig()

	ticker := time.Tick(time.Duration(conf.Spotify.UpdateInterval) * time.Second)

	for {
		select {
		case <-ticker:
			logger.Printf("updating Spotify player state for %d", id)
		case <-shutdown:
			logger.Printf("shutting down worker for %d", id)
			return
		}
	}

}

func Shutdown() {
	close(shutdownCh)
	wg.Wait()
}
