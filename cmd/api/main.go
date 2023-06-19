package main

import (
	"context"
	"database/sql"
	"flag"
	"greenlight.badrchoubai.dev/internal/data"
	"greenlight.badrchoubai.dev/internal/jsonlog"
	"greenlight.badrchoubai.dev/internal/mailer"
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

	smtpOptions struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}

	config struct {
		port    int
		env     string
		db      connectionPoolSettings
		limiter rateLimiterSettings
		smtp    smtpOptions
	}

	application struct {
		config config
		log    *jsonlog.Logger
		models data.Models
		mailer mailer.Mailer
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

	// Setup smtp mail server
	flag.StringVar(&config.smtp.host, "smtp-host", "smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&config.smtp.port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&config.smtp.username, "smtp-username", "4ff9f878683e32", "SMTP username")
	flag.StringVar(&config.smtp.password, "smtp-password", "3b827122aa06cd", "SMTP password")
	flag.StringVar(&config.smtp.sender, "smtp-sender", "Greenlight <no-reply@greenlight.badrchoubai.dev>", "SMTP sender")

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
		mailer: mailer.New(
			config.smtp.host,
			config.smtp.port,
			config.smtp.username,
			config.smtp.password,
			config.smtp.sender,
		),
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
