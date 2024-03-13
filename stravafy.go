package main

import (
	"context"
	_ "embed"
	"log"
	"os"
	"os/signal"
	"stravafy/internal/config"
	"stravafy/internal/database"
	"stravafy/internal/server"
	"stravafy/internal/worker"
	"syscall"
	"time"
)

//go:embed sql/schema.sql
var schema string

func main() {
	log.SetFlags(log.LstdFlags)
	log.Print("Starting ...")
	configPath, isSet := os.LookupEnv("STRAVAFY_CONFIG_PATH")
	if !isSet {
		configPath = ".stravafy"
	}
	err := config.Setup(configPath)
	if err != nil {
		log.Fatalf("Upsi daisy config not working: %v", err)
	}

	db, err := database.NewSQLite()
	if err != nil {
		log.Fatalf("nono database: %v", err)
	}
	_, err = db.DB.ExecContext(context.Background(), schema)
	if err != nil {
		log.Fatalf("nono database: %v", err)
	}
	queries := database.New(db.DB)
	//db, err := database.NewDebugDB()
	//if err != nil {
	//	log.Fatalf("nono database: %v", err)
	//}
	//_, err = db.ExecContext(context.Background(), schema)
	//if err != nil {
	//	log.Fatalf("som wrong wis se migration: %v", err)
	//}
	//queries := database.New(db)
	server.Init(queries)

	go func() {
		if err := server.Run(); err != nil {
			log.Printf("an error accoured: %v\n", err)
		}
	}()

	go worker.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("an error accoured while shuting down the server: %v", err)
	}

	worker.Shutdown()

}
