package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
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
type config struct {
	port int
	env  string
}

type application struct {
	appConfig config
	log       *log.Logger
}

func main() {
	var config config

	flag.IntVar(&config.port, "port", 4000, "API server port")
	flag.StringVar(&config.env, "env", "development", "API server environment")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	application := &application{
		appConfig: config,
		log:       logger,
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.port),
		Handler:      application.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start the HTTP server.
	logger.Printf("starting %s server on http://127.0.0.1%s", config.env, server.Addr)
	err := server.ListenAndServe()
	logger.Fatal(err)

}
