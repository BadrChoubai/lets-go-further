package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		application.log.PrintInfo("caught signal", map[string]string{
			"signal": s.String(),
		})

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := server.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}
		application.log.PrintInfo("completing background tasks", map[string]string{
			"addr": server.Addr,
		})

		shutdownError <- nil
	}()

	application.log.PrintInfo("server starting", map[string]string{
		"host":        "127.0.0.1",
		"port":        server.Addr,
		"base_url":    "http://127.0.0.1:4000",
		"environment": application.config.env,
		"healthcheck": "http://127.0.0.1:4000/api/healthcheck",
	})

	err := server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	application.log.PrintInfo("server stopped", map[string]string{
		"host":        "127.0.0.1",
		"port":        server.Addr,
		"environment": application.config.env,
	})

	return nil
}
