package main

import (
	"context"
	"database/sql"
	"flag"
	"greenlight.badrchoubai.dev/internal/data"
	"greenlight.badrchoubai.dev/internal/jsonlog"
	"os"
	"time"

	_ "github.com/lib/pq"
)

const version = "1.0.0"

type (
	connectionPoolSettings struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}

	rateLimiterSettings struct {
		rps     float64
		burst   int
		enabled bool
	}

	config struct {
		port    int
		env     string
		db      connectionPoolSettings
		limiter rateLimiterSettings
	}

	application struct {
		config config
		log    *jsonlog.Logger
		models data.Models
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

	// Setup Rate Limiter Settings
	flag.Float64Var(&config.limiter.rps, "limiter-rps", 2, "Rate limiter: maximum requests per second")
	flag.IntVar(&config.limiter.burst, "limiter-burst", 4, "Rate limiter: maximum burst")
	flag.BoolVar(&config.limiter.enabled, "limiter-enabled", true, "Rate limiter: enabled or disabled")

	flag.Parse()

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDB(config)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	defer db.Close()

	logger.PrintInfo("database: connection pool established", nil)

	application := &application{
		config: config,
		log:    logger,
		models: data.NewModels(db),
	}

	// Start the HTTP server.
	err = application.serve()
	if err != nil {
		application.log.PrintFatal(err, nil)
	}
}

func openDB(cfg config) (*sql.DB, error) {
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
