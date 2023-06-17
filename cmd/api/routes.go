package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (application *application) routes() http.Handler {
	// Initialize a new httprouter router instance.
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(application.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(application.methodNotAllowedResponse)

	// Register the relevant methods, URL patterns and handler functions for our
	// endpoints using the HandlerFunc() method. Note that http.MethodGet and
	// http.MethodPost are constants which equate to the strings "GET" and "POST"
	// respectively.
	router.HandlerFunc(http.MethodGet, "/api/healthcheck", application.healthcheckHandler)
	router.HandlerFunc(http.MethodGet, "/api/v1/movies", application.listMoviesHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/movies", application.createMovieHandler)
	router.HandlerFunc(http.MethodGet, "/api/v1/movies/:id", application.showMovieHandler)
	router.HandlerFunc(http.MethodPatch, "/api/v1/movies/:id", application.updateMovieHandler)
	router.HandlerFunc(http.MethodDelete, "/api/v1/movies/:id", application.deleteMovieHandler)

	// Return the httprouter instance.
	return application.recoverPanic(application.rateLimiter(router))
}
