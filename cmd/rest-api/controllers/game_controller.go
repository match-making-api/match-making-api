package controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/leet-gaming/match-making-api/pkg/common"
	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	game_in "github.com/leet-gaming/match-making-api/pkg/domain/game/ports/in"
)

type GameController struct {
	Container container.Container
}

func NewGameController(container container.Container) *GameController {
	return &GameController{Container: container}
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// Get retrieves a game by ID
func (gc *GameController) Get(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		vars := mux.Vars(r)
		gameIDStr, ok := vars["id"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "bad_request",
				Message: "game ID is required",
			})
			return
		}

		gameID, err := uuid.Parse(gameIDStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_id",
				Message: "invalid game ID format",
			})
			return
		}

		var getGameQuery game_in.GetGameByIDQuery
		if err := gc.Container.Resolve(&getGameQuery); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve GetGameByIDQuery", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		game, err := getGameQuery.Execute(r.Context(), gameID)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to get game", "error", err, "game_id", gameID)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "not_found",
				Message: "game not found",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(game)
	}
}

// Create creates a new game
func (gc *GameController) Create(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "method_not_allowed",
				Message: "only POST method is allowed",
			})
			return
		}

		// SECURITY: Admin-only endpoint
		if !common.IsAdmin(r.Context()) {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "forbidden",
				Message: "administrator access required",
			})
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var game game_entities.Game
		if err := json.NewDecoder(r.Body).Decode(&game); err != nil {
			slog.ErrorContext(r.Context(), "failed to decode request body", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_request",
				Message: "invalid request body",
			})
			return
		}

		var createGameCmd game_in.CreateGameCommand
		if err := gc.Container.Resolve(&createGameCmd); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve CreateGameCommand", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		createdGame, err := createGameCmd.Execute(r.Context(), &game)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to create game", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "validation_error",
				Message: "failed to create game",
			})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(createdGame)
	}
}

// Update updates an existing game
func (gc *GameController) Update(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPut && r.Method != http.MethodPatch {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "method_not_allowed",
				Message: "only PUT or PATCH methods are allowed",
			})
			return
		}

		// SECURITY: Admin-only endpoint
		if !common.IsAdmin(r.Context()) {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "forbidden",
				Message: "administrator access required",
			})
			return
		}

		vars := mux.Vars(r)
		gameIDStr, ok := vars["id"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "bad_request",
				Message: "game ID is required",
			})
			return
		}

		gameID, err := uuid.Parse(gameIDStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_id",
				Message: "invalid game ID format",
			})
			return
		}

		var game game_entities.Game
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		if err := json.NewDecoder(r.Body).Decode(&game); err != nil {
			slog.ErrorContext(r.Context(), "failed to decode request body", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_request",
				Message: "invalid request body",
			})
			return
		}

		var updateGameCmd game_in.UpdateGameCommand
		if err := gc.Container.Resolve(&updateGameCmd); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve UpdateGameCommand", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		updatedGame, err := updateGameCmd.Execute(r.Context(), gameID, &game)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to update game", "error", err, "game_id", gameID)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "validation_error",
				Message: "failed to update game",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(updatedGame)
	}
}

// Delete deletes a game (or disables it if enabled)
func (gc *GameController) Delete(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "method_not_allowed",
				Message: "only DELETE method is allowed",
			})
			return
		}

		// SECURITY: Admin-only endpoint
		if !common.IsAdmin(r.Context()) {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "forbidden",
				Message: "administrator access required",
			})
			return
		}

		vars := mux.Vars(r)
		gameIDStr, ok := vars["id"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "bad_request",
				Message: "game ID is required",
			})
			return
		}

		gameID, err := uuid.Parse(gameIDStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_id",
				Message: "invalid game ID format",
			})
			return
		}

		var deleteGameCmd game_in.DeleteGameCommand
		if err := gc.Container.Resolve(&deleteGameCmd); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve DeleteGameCommand", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		if err := deleteGameCmd.Execute(r.Context(), gameID); err != nil {
			slog.ErrorContext(r.Context(), "failed to delete game", "error", err, "game_id", gameID)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "delete_error",
				Message: "failed to delete game",
			})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// List lists all games
func (gc *GameController) List(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var searchGamesQuery game_in.SearchGamesQuery
		if err := gc.Container.Resolve(&searchGamesQuery); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve SearchGamesQuery", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		games, err := searchGamesQuery.Execute(r.Context())
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to list games", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to retrieve games",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(games)
	}
}
