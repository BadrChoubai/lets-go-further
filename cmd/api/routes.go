package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (application *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(application.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(application.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/api/healthcheck", application.healthcheckHandler)
	router.HandlerFunc(http.MethodGet, "/api/v1/movies", application.listMoviesHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/movies", application.createMovieHandler)
	router.HandlerFunc(http.MethodGet, "/api/v1/movies/:id", application.showMovieHandler)
	router.HandlerFunc(http.MethodPatch, "/api/v1/movies/:id", application.updateMovieHandler)
	router.HandlerFunc(http.MethodDelete, "/api/v1/movies/:id", application.deleteMovieHandler)

	return application.recoverPanic(application.rateLimiter(router))
}
