package main

import (
	"expvar"
	http "net/http"

	"github.com/julienschmidt/httprouter"
)

func (application *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(application.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(application.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/healthcheck", application.healthcheckHandler)

	// API routes
	router.HandlerFunc(http.MethodGet, "/api/v1/movies", application.requirePermission("movies:read", application.listMoviesHandler))
	router.HandlerFunc(http.MethodPost, "/api/v1/movies", application.requirePermission("movies:write", application.createMovieHandler))
	router.HandlerFunc(http.MethodGet, "/api/v1/movies/:id", application.requirePermission("movies:read", application.showMovieHandler))
	router.HandlerFunc(http.MethodPatch, "/api/v1/movies/:id", application.requirePermission("movies:write", application.updateMovieHandler))
	router.HandlerFunc(http.MethodDelete, "/api/v1/movies/:id", application.requirePermission("movies:write", application.deleteMovieHandler))

	// User Routes
	router.HandlerFunc(http.MethodPost, "/users", application.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/users/activate", application.activateUserHandler)

	// Token Routes
	router.HandlerFunc(http.MethodPost, "/tokens/authentication", application.createAuthenticationTokenHandler)

	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	return application.metrics(application.recoverPanic(application.enableCORS(application.rateLimiter(application.authenticate(router)))))
}
