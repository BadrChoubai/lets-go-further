package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func (application *application) serve() error {
	// Declare a HTTP server using the same settings as in our main() function.
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", application.config.port),
		Handler:      application.routes(),
		ErrorLog:     log.New(application.log, "", 0),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	application.log.PrintInfo("server running", map[string]string{
		"host":        "127.0.0.1",
		"port":        server.Addr,
		"base_url":    "http://127.0.0.1:4000",
		"environment": application.config.env,
		"healthcheck": "http://127.0.0.1:4000/api/healthcheck",
	})

	// Start the server as normal, returning any error.
	return server.ListenAndServe()
}
