package main

import (
	http "net/http"

	"github.com/julienschmidt/httprouter"
)

func (application *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(application.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(application.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/healthcheck", application.healthcheckHandler)

	// API routes
	router.HandlerFunc(http.MethodGet, "/api/v1/movies", application.listMoviesHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/movies", application.createMovieHandler)
	router.HandlerFunc(http.MethodGet, "/api/v1/movies/:id", application.showMovieHandler)
	router.HandlerFunc(http.MethodPatch, "/api/v1/movies/:id", application.updateMovieHandler)
	router.HandlerFunc(http.MethodDelete, "/api/v1/movies/:id", application.deleteMovieHandler)

	// User Routes
	router.HandlerFunc(http.MethodPost, "/users", application.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/users/activate", application.activateUserHandler)

	return application.recoverPanic(application.rateLimiter(router))
}
