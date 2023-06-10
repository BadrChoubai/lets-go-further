package main

import (
	"encoding/json"
	"net/http"
)

type HealthCheckResponse struct {
	Status      string `json:"status"`
	Environment string `json:"environment"`
	Version     string `json:"version"`
}

// Declare a handler which writes a plain-text response with information about the
// application status, operating environment and version.
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	response := &HealthCheckResponse{
		Status:      "available",
		Environment: app.appConfig.env,
		Version:     version,
	}

	w.Header().Set("Accept", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

}
