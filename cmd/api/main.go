package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"greenlight.badrchoubai.dev/internal/data"
	"greenlight.badrchoubai.dev/internal/jsonlog"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// Declare a string containing the application version number. Later in the book we'll
// generate this automatically at build time, but for now we'll just store the version
// number as a hard-coded global constant.
const version = "1.0.0"

// Define a config struct to hold all the configuration settings for our application.
// For now, the only configuration settings will be the network port that we want the
// server to listen on, and the name of the current operating environment for the
// application (development, staging, production, etc.). We will read in these
// configuration settings from command-line flags when the application starts.
type (
	connectionPoolSettings struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}

	config struct {
		port int
		env  string
		db   connectionPoolSettings
	}

	application struct {
		appConfig config
		log       *jsonlog.Logger
		models    data.Models
	}
)

func main() {
	var config config

	flag.IntVar(&config.port, "port", 4000, "API server port")
	flag.StringVar(&config.env, "env", "development", "Environment (development|staging|production)")

	// Setup DB Connection Pool
	flag.StringVar(&config.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")
	flag.IntVar(&config.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&config.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&config.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	flag.Parse()

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDb(config)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	defer db.Close()

	logger.PrintInfo("database: connection pool established", nil)

	application := &application{
		appConfig: config,
		log:       logger,
		models:    data.NewModels(db),
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.port),
		Handler:      application.routes(),
		ErrorLog:     log.New(logger, "", 0),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.PrintInfo("server running", map[string]string{
		"host":        "127.0.0.1",
		"port":        server.Addr,
		"base_url":    "http://127.0.0.1:4000",
		"environment": config.env,
		"healthcheck": "http://127.0.0.1:4000/api/healthcheck",
	})

	// Start the HTTP server.
	err = server.ListenAndServe()
	logger.PrintFatal(err, nil)
}

func openDb(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
