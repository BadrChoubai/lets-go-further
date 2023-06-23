package main

import (
	"fmt"
	"net/http"
)

func (application *application) authenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	message := "you must be authenticated in order to access this resource"
	application.errorResponse(w, r, http.StatusUnauthorized, message)
}

func (application *application) logError(r *http.Request, err error) {
	application.log.PrintError(err, map[string]string{
		"request_method": r.Method,
		"request_url":    r.URL.String(),
	})
}

func (application *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := envelope{"error": message}

	err := application.writeJSON(w, status, env, nil)
	if err != nil {
		application.logError(r, err)
		w.WriteHeader(500)
	}
}

func (application *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	application.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (application *application) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the requested record due to an edit conflict, please try again"
	application.errorResponse(w, r, http.StatusConflict, message)
}

func (application *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	application.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

func (application *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded"
	application.errorResponse(w, r, http.StatusTooManyRequests, message)
}

func (application *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	application.logError(r, err)

	message := "the server encountered a problem and could not process the request"
	application.errorResponse(w, r, http.StatusInternalServerError, message)
}

func (application *application) inactiveAccountResponse(w http.ResponseWriter, r *http.Request) {
	message := "your user account must be activated to access this resource"
	application.errorResponse(w, r, http.StatusForbidden, message)
}

func (application *application) invalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")

	message := "invalid or missing authentication token"
	application.errorResponse(w, r, http.StatusUnauthorized, message)
}

func (application *application) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	message := "invalid authentication credentials"
	application.errorResponse(w, r, http.StatusUnauthorized, message)
}

func (application *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("Method: '%s', is not supported for this request", r.Method)
	application.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

func (application *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	application.errorResponse(w, r, http.StatusInternalServerError, message)
}
