package routing

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/gorilla/mux"
	"github.com/leet-gaming/match-making-api/cmd/rest-api/controllers"
	"github.com/leet-gaming/match-making-api/cmd/rest-api/middlewares"
	"github.com/leet-gaming/match-making-api/pkg/infra/config"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	Health string = "/health"
	Search string = "/search/{query:.*}"
)

// NewRouter creates and configures a new HTTP router for the application.
//
// It sets up middleware, defines routes, and configures Swagger documentation.
//
// Parameters:
//   - ctx: A context.Context that can be used for cancellation or passing values.
//   - container: A container.Container instance for dependency injection.
//
// Returns:
//
//	An http.Handler that can be used to serve HTTP requests.
func NewRouter(ctx context.Context, container container.Container) http.Handler {
	r := mux.NewRouter()

	// middleware
	resourceContextMiddleware := middlewares.NewResourceContextMiddleware(&container)
	corsMiddleware := middlewares.NewCORSMiddleware()

	r.Use(corsMiddleware.Handler)
	r.Use(mux.CORSMethodMiddleware(r))
	r.Use(resourceContextMiddleware.Handler)

	// controllers
	healthController := controllers.NewHealthController(container)
	gameController := controllers.NewGameController(container)
	gameModeController := controllers.NewGameModeController(container)
	regionController := controllers.NewRegionController(container)
	invitationController := controllers.NewInvitationController(container)
	externalInvitationController := controllers.NewExternalInvitationController(container)
	notificationController := controllers.NewNotificationController(container)

	// health
	r.HandleFunc(Health, healthController.HealthCheck(ctx)).Methods("GET")
	resourceContextMiddleware.RegisterOperation(Health, "match-making:health:get")

	// games
	r.HandleFunc("/games", gameController.List(ctx)).Methods("GET")
	r.HandleFunc("/games", gameController.Create(ctx)).Methods("POST")
	r.HandleFunc("/games/{id}", gameController.Get(ctx)).Methods("GET")
	r.HandleFunc("/games/{id}", gameController.Update(ctx)).Methods("PUT", "PATCH")
	r.HandleFunc("/games/{id}", gameController.Delete(ctx)).Methods("DELETE")
	resourceContextMiddleware.RegisterOperation("/games", "match-making:games:list")
	resourceContextMiddleware.RegisterOperation("/games", "match-making:games:create")
	resourceContextMiddleware.RegisterOperation("/games/{id}", "match-making:games:get")
	resourceContextMiddleware.RegisterOperation("/games/{id}", "match-making:games:update")
	resourceContextMiddleware.RegisterOperation("/games/{id}", "match-making:games:delete")

	// game modes
	r.HandleFunc("/game-modes", gameModeController.List(ctx)).Methods("GET")
	r.HandleFunc("/game-modes", gameModeController.Create(ctx)).Methods("POST")
	r.HandleFunc("/game-modes/{id}", gameModeController.Get(ctx)).Methods("GET")
	r.HandleFunc("/game-modes/{id}", gameModeController.Update(ctx)).Methods("PUT", "PATCH")
	r.HandleFunc("/game-modes/{id}", gameModeController.Delete(ctx)).Methods("DELETE")
	resourceContextMiddleware.RegisterOperation("/game-modes", "match-making:game-modes:list")
	resourceContextMiddleware.RegisterOperation("/game-modes", "match-making:game-modes:create")
	resourceContextMiddleware.RegisterOperation("/game-modes/{id}", "match-making:game-modes:get")
	resourceContextMiddleware.RegisterOperation("/game-modes/{id}", "match-making:game-modes:update")
	resourceContextMiddleware.RegisterOperation("/game-modes/{id}", "match-making:game-modes:delete")

	// regions
	r.HandleFunc("/regions", regionController.List(ctx)).Methods("GET")
	r.HandleFunc("/regions", regionController.Create(ctx)).Methods("POST")
	r.HandleFunc("/regions/{id}", regionController.Get(ctx)).Methods("GET")
	r.HandleFunc("/regions/{id}", regionController.Update(ctx)).Methods("PUT", "PATCH")
	r.HandleFunc("/regions/{id}", regionController.Delete(ctx)).Methods("DELETE")
	resourceContextMiddleware.RegisterOperation("/regions", "match-making:regions:list")
	resourceContextMiddleware.RegisterOperation("/regions", "match-making:regions:create")
	resourceContextMiddleware.RegisterOperation("/regions/{id}", "match-making:regions:get")
	resourceContextMiddleware.RegisterOperation("/regions/{id}", "match-making:regions:update")
	resourceContextMiddleware.RegisterOperation("/regions/{id}", "match-making:regions:delete")

	// invitations
	r.HandleFunc("/invitations", invitationController.Create(ctx)).Methods("POST")
	r.HandleFunc("/invitations", invitationController.List(ctx)).Methods("GET")
	r.HandleFunc("/invitations/{id}", invitationController.Get(ctx)).Methods("GET")
	r.HandleFunc("/invitations/{id}/accept", invitationController.Accept(ctx)).Methods("POST")
	r.HandleFunc("/invitations/{id}/decline", invitationController.Decline(ctx)).Methods("POST")
	r.HandleFunc("/invitations/{id}", invitationController.Update(ctx)).Methods("PATCH")
	r.HandleFunc("/invitations/{id}", invitationController.Delete(ctx)).Methods("DELETE")
	resourceContextMiddleware.RegisterOperation("/invitations", "match-making:invitations:create")
	resourceContextMiddleware.RegisterOperation("/invitations", "match-making:invitations:list")
	resourceContextMiddleware.RegisterOperation("/invitations/{id}", "match-making:invitations:get")
	resourceContextMiddleware.RegisterOperation("/invitations/{id}/accept", "match-making:invitations:accept")
	resourceContextMiddleware.RegisterOperation("/invitations/{id}/decline", "match-making:invitations:decline")
	resourceContextMiddleware.RegisterOperation("/invitations/{id}", "match-making:invitations:update")
	resourceContextMiddleware.RegisterOperation("/invitations/{id}", "match-making:invitations:delete")

	// external invitations
	r.HandleFunc("/external-invitations", externalInvitationController.Create(ctx)).Methods("POST")
	r.HandleFunc("/external-invitations", externalInvitationController.List(ctx)).Methods("GET")
	r.HandleFunc("/external-invitations/by-token", externalInvitationController.GetByToken(ctx)).Methods("GET")
	r.HandleFunc("/external-invitations/{id}", externalInvitationController.Get(ctx)).Methods("GET")
	r.HandleFunc("/external-invitations/{id}/resend", externalInvitationController.Resend(ctx)).Methods("POST")
	r.HandleFunc("/external-invitations/{id}", externalInvitationController.Delete(ctx)).Methods("DELETE")
	resourceContextMiddleware.RegisterOperation("/external-invitations", "match-making:external-invitations:create")
	resourceContextMiddleware.RegisterOperation("/external-invitations", "match-making:external-invitations:list")
	resourceContextMiddleware.RegisterOperation("/external-invitations/by-token", "match-making:external-invitations:get-by-token")
	resourceContextMiddleware.RegisterOperation("/external-invitations/{id}", "match-making:external-invitations:get")
	resourceContextMiddleware.RegisterOperation("/external-invitations/{id}/resend", "match-making:external-invitations:resend")
	resourceContextMiddleware.RegisterOperation("/external-invitations/{id}", "match-making:external-invitations:delete")

	// notifications
	r.HandleFunc("/notifications", notificationController.Send(ctx)).Methods("POST")
	r.HandleFunc("/notifications/batch", notificationController.SendBatch(ctx)).Methods("POST")
	r.HandleFunc("/notifications/users/{user_id}", notificationController.GetUserNotifications(ctx)).Methods("GET")
	r.HandleFunc("/notifications/{id}/read", notificationController.MarkAsRead(ctx)).Methods("POST")
	r.HandleFunc("/notifications/{id}/retry", notificationController.Retry(ctx)).Methods("POST")
	resourceContextMiddleware.RegisterOperation("/notifications", "match-making:notifications:send")
	resourceContextMiddleware.RegisterOperation("/notifications/batch", "match-making:notifications:send-batch")
	resourceContextMiddleware.RegisterOperation("/notifications/users/{user_id}", "match-making:notifications:get-user")
	resourceContextMiddleware.RegisterOperation("/notifications/{id}/read", "match-making:notifications:mark-read")
	resourceContextMiddleware.RegisterOperation("/notifications/{id}/retry", "match-making:notifications:retry")

	// lobbies - resolve MongoDB client and config for lobby controller
	var mongoClient *mongo.Client
	var cfg config.Config
	if err := container.Resolve(&mongoClient); err != nil {
		slog.Error("Failed to resolve MongoDB client for lobbies", "error", err)
	}
	if err := container.Resolve(&cfg); err != nil {
		slog.Error("Failed to resolve config for lobbies", "error", err)
	}
	
	if mongoClient != nil {
		lobbyController := controllers.NewLobbyController(mongoClient, cfg.MongoDB.DBName)
		
		// Lobby CRUD
		r.HandleFunc("/api/lobbies", lobbyController.List(ctx)).Methods("GET", "OPTIONS")
		r.HandleFunc("/api/lobbies", lobbyController.Create(ctx)).Methods("POST", "OPTIONS")
		r.HandleFunc("/api/lobbies/featured", lobbyController.GetFeatured(ctx)).Methods("GET", "OPTIONS")
		r.HandleFunc("/api/lobbies/stats", lobbyController.GetStats(ctx)).Methods("GET", "OPTIONS")
		r.HandleFunc("/api/lobbies/seed", lobbyController.SeedDemoLobbies(ctx)).Methods("POST", "OPTIONS")
		r.HandleFunc("/api/lobbies/{id}", lobbyController.Get(ctx)).Methods("GET", "OPTIONS")
		r.HandleFunc("/api/lobbies/{id}", lobbyController.Delete(ctx)).Methods("DELETE", "OPTIONS")
		r.HandleFunc("/api/lobbies/{id}/join", lobbyController.Join(ctx)).Methods("POST", "OPTIONS")
		
		resourceContextMiddleware.RegisterOperation("/api/lobbies", "match-making:lobbies:list")
		resourceContextMiddleware.RegisterOperation("/api/lobbies", "match-making:lobbies:create")
		resourceContextMiddleware.RegisterOperation("/api/lobbies/featured", "match-making:lobbies:featured")
		resourceContextMiddleware.RegisterOperation("/api/lobbies/stats", "match-making:lobbies:stats")
		resourceContextMiddleware.RegisterOperation("/api/lobbies/seed", "match-making:lobbies:seed")
		resourceContextMiddleware.RegisterOperation("/api/lobbies/{id}", "match-making:lobbies:get")
		resourceContextMiddleware.RegisterOperation("/api/lobbies/{id}", "match-making:lobbies:delete")
		resourceContextMiddleware.RegisterOperation("/api/lobbies/{id}/join", "match-making:lobbies:join")
		
		slog.Info("Lobby routes registered")
	}

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
