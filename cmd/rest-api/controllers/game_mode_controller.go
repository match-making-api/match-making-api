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

type GameModeController struct {
	Container container.Container
}

func NewGameModeController(container container.Container) *GameModeController {
	return &GameModeController{Container: container}
}

// Get retrieves a game mode by ID
func (gmc *GameModeController) Get(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		vars := mux.Vars(r)
		gameModeIDStr, ok := vars["id"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "bad_request",
				Message: "game mode ID is required",
			})
			return
		}

		gameModeID, err := uuid.Parse(gameModeIDStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_id",
				Message: "invalid game mode ID format",
			})
			return
		}

		var getGameModeQuery game_in.GetGameModeByIDQuery
		if err := gmc.Container.Resolve(&getGameModeQuery); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve GetGameModeByIDQuery", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		gameMode, err := getGameModeQuery.Execute(r.Context(), gameModeID)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to get game mode", "error", err, "game_mode_id", gameModeID)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "not_found",
				Message: "game mode not found",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(gameMode)
	}
}

// Create creates a new game mode
func (gmc *GameModeController) Create(ctx context.Context) http.HandlerFunc {
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
		var gameMode game_entities.GameMode
		if err := json.NewDecoder(r.Body).Decode(&gameMode); err != nil {
			slog.ErrorContext(r.Context(), "failed to decode request body", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_request",
				Message: "invalid request body",
			})
			return
		}

		var createGameModeCmd game_in.CreateGameModeCommand
		if err := gmc.Container.Resolve(&createGameModeCmd); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve CreateGameModeCommand", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		createdGameMode, err := createGameModeCmd.Execute(r.Context(), &gameMode)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to create game mode", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "validation_error",
				Message: "failed to create game mode",
			})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(createdGameMode)
	}
}

// Update updates an existing game mode
func (gmc *GameModeController) Update(ctx context.Context) http.HandlerFunc {
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
		gameModeIDStr, ok := vars["id"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "bad_request",
				Message: "game mode ID is required",
			})
			return
		}

		gameModeID, err := uuid.Parse(gameModeIDStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_id",
				Message: "invalid game mode ID format",
			})
			return
		}

		var gameMode game_entities.GameMode
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		if err := json.NewDecoder(r.Body).Decode(&gameMode); err != nil {
			slog.ErrorContext(r.Context(), "failed to decode request body", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_request",
				Message: "invalid request body",
			})
			return
		}

		var updateGameModeCmd game_in.UpdateGameModeCommand
		if err := gmc.Container.Resolve(&updateGameModeCmd); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve UpdateGameModeCommand", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		updatedGameMode, err := updateGameModeCmd.Execute(r.Context(), gameModeID, &gameMode)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to update game mode", "error", err, "game_mode_id", gameModeID)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "validation_error",
				Message: "failed to update game mode",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(updatedGameMode)
	}
}

// Delete deletes a game mode
func (gmc *GameModeController) Delete(ctx context.Context) http.HandlerFunc {
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
		gameModeIDStr, ok := vars["id"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "bad_request",
				Message: "game mode ID is required",
			})
			return
		}

		gameModeID, err := uuid.Parse(gameModeIDStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_id",
				Message: "invalid game mode ID format",
			})
			return
		}

		var deleteGameModeCmd game_in.DeleteGameModeCommand
		if err := gmc.Container.Resolve(&deleteGameModeCmd); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve DeleteGameModeCommand", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		if err := deleteGameModeCmd.Execute(r.Context(), gameModeID); err != nil {
			slog.ErrorContext(r.Context(), "failed to delete game mode", "error", err, "game_mode_id", gameModeID)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "delete_error",
				Message: "failed to delete game mode",
			})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// List lists all game modes
func (gmc *GameModeController) List(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var searchGameModesQuery game_in.SearchGameModesQuery
		if err := gmc.Container.Resolve(&searchGameModesQuery); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve SearchGameModesQuery", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		gameModes, err := searchGameModesQuery.Execute(r.Context())
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to list game modes", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to retrieve game modes",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(gameModes)
	}
}
