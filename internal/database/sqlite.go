package database

import (
	"context"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"stravafy/internal/config"
)

var (
	logger *log.Logger
)

func init() {
	logfile, err := os.OpenFile("db.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("error opening db.log: %v", err)
	}
	logger = log.New(logfile, "", log.LstdFlags)
}

type SQLite struct {
	DB *sql.DB
}

func NewSQLite() (*SQLite, error) {
	conf := config.GetConfig()
	db, err := sql.Open("sqlite3", conf.Database.Source)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &SQLite{DB: db}, nil
}

type DebugDB struct {
	db *sql.DB
}

func (d *DebugDB) ExecContext(c context.Context, query string, data ...interface{}) (sql.Result, error) {
	logger.Println("DB ExecuteContext: ", query, data)
	return d.db.ExecContext(c, query, data...)

}
func (d *DebugDB) PrepareContext(c context.Context, query string) (*sql.Stmt, error) {
	logger.Println("DB PrepareContext: ", query)
	return d.db.PrepareContext(c, query)
}
func (d *DebugDB) QueryContext(c context.Context, query string, data ...interface{}) (*sql.Rows, error) {
	logger.Println("DB QueryContext: ", query, data)
	return d.db.QueryContext(c, query, data...)
}
func (d *DebugDB) QueryRowContext(c context.Context, query string, data ...interface{}) *sql.Row {
	logger.Println("DB QueryRowContext: ", query, data)
	return d.db.QueryRowContext(c, query, data...)
}

func NewDebugDB() (*DebugDB, error) {
	conf := config.GetConfig()
	db, err := sql.Open("sqlite3", conf.Database.Source)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &DebugDB{db: db}, nil
}
