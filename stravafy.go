package main

import (
	"context"
	_ "embed"
	"log"
	"os"
	"stravafy/internal/config"
	"stravafy/internal/database"
	"stravafy/internal/server"
)

//go:embed sql/schema.sql
var schema string

func main() {
	log.Print("Starting ...")
	configPath, isSet := os.LookupEnv("STRAVAFY_CONFIG_PATH")
	if !isSet {
		configPath = ".stravafy"
	}
	err := config.Setup(configPath)
	if err != nil {
		log.Fatalf("Upsi daisy config not working: %v", err)
	}

	//db, err := database.NewSQLite()
	//if err != nil {
	//	log.Fatalf("nono database: %v", err)
	//}
	//_, err = db.DB.ExecContext(context.Background(), schema)
	db, err := database.NewDebugDB()
	if err != nil {
		log.Fatalf("nono database: %v", err)
	}
	_, err = db.ExecContext(context.Background(), schema)
	if err != nil {
		log.Fatalf("som wrong wis se migration: %v", err)
	}
	queries := database.New(db)
	server.Init(queries)

	err = server.Run()
	if err != nil {
		return
	}
}
