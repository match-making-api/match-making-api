package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/golobby/container/v3"
)

// HealthCheckResponse represents the response for a health check request.
type HealthCheckResponse struct {
	Status string `json:"status"`
}

// HealthController handles the health check endpoint.
type HealthController struct {
	Container container.Container
}

// NewHealthController creates a new instance of the HealthController.
func NewHealthController(container container.Container) *HealthController {
	return &HealthController{Container: container}
}

// HealthCheck returns an HTTP handler function for performing a health check.
//
// Parameters:
//   - apiContext: A context.Context that carries deadlines, cancellation signals,
//     and other request-scoped values across API boundaries and between processes.
//
// Returns:
//   - An http.HandlerFunc that, when invoked, writes a JSON response with a
//     status of "ok" and sets the HTTP status code to 200 OK.
func (hc *HealthController) HealthCheck(apiContext context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := HealthCheckResponse{
			Status: "ok",
		}

		json.NewEncoder(w).Encode(response)
	}
}
