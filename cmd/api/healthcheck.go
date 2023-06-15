package main

import (
	"net/http"
)

type SystemInfo struct {
	Environment string `json:"environment"`
	Version     string `json:"version"`
}

type HealthCheckResponse struct {
	Status     string     `json:"status"`
	SystemInfo SystemInfo `json:"system_info"`
}

// Declare a handler which writes a plain-text response with information about the
// application status, operating environment and version.
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	response := envelope{"health": &HealthCheckResponse{
		Status: "available",
		SystemInfo: SystemInfo{
			Environment: app.appConfig.env,
			Version:     version,
		},
	}}

	err := app.writeJSON(w, http.StatusOK, response, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
