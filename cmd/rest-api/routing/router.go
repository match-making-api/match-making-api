package routing

import (
	"context"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/gorilla/mux"
	"github.com/leet-gaming/match-making-api/cmd/rest-api/controllers"
	"github.com/leet-gaming/match-making-api/cmd/rest-api/middlewares"
	httpSwagger "github.com/swaggo/http-swagger"
)

const (
	Health string = "/health"
	Search string = "/search/{query:.*}"
)

// NewRouter creates and configures a new HTTP router for the application.
func NewRouter(ctx context.Context, container container.Container) http.Handler {
	r := mux.NewRouter()

	// middleware
	resourceContextMiddleware := middlewares.NewResourceContextMiddleware(&container)

	r.Use(mux.CORSMethodMiddleware(r))
	r.Use(resourceContextMiddleware.Handler)

	// controllers
	healthController := controllers.NewHealthController(container)

	// health
	r.HandleFunc(Health, healthController.HealthCheck(ctx)).Methods("GET")

	// Swagger UI
	r.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/docs/openapi.yaml"),
	))

	// Serve the OpenAPI spec
	r.HandleFunc("/docs/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/openapi.yaml")
	})

	// Add more routes here as needed

	return r
}
