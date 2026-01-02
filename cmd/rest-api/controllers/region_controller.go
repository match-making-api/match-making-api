package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	game_in "github.com/leet-gaming/match-making-api/pkg/domain/game/ports/in"
)

type RegionController struct {
	Container container.Container
}

func NewRegionController(container container.Container) *RegionController {
	return &RegionController{Container: container}
}

// Get retrieves a region by ID
func (rc *RegionController) Get(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		vars := mux.Vars(r)
		regionIDStr, ok := vars["id"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "bad_request",
				Message: "region ID is required",
			})
			return
		}

		regionID, err := uuid.Parse(regionIDStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_id",
				Message: "invalid region ID format",
			})
			return
		}

		var getRegionQuery game_in.GetRegionByIDQuery
		if err := rc.Container.Resolve(&getRegionQuery); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve GetRegionByIDQuery", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		region, err := getRegionQuery.Execute(r.Context(), regionID)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to get region", "error", err, "region_id", regionID)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "not_found",
				Message: fmt.Sprintf("region not found: %v", err),
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(region)
	}
}

// Create creates a new region
func (rc *RegionController) Create(ctx context.Context) http.HandlerFunc {
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

		var region game_entities.Region
		if err := json.NewDecoder(r.Body).Decode(&region); err != nil {
			slog.ErrorContext(r.Context(), "failed to decode request body", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_request",
				Message: fmt.Sprintf("invalid JSON: %v", err),
			})
			return
		}

		var createRegionCmd game_in.CreateRegionCommand
		if err := rc.Container.Resolve(&createRegionCmd); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve CreateRegionCommand", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		createdRegion, err := createRegionCmd.Execute(r.Context(), &region)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to create region", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "validation_error",
				Message: err.Error(),
			})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(createdRegion)
	}
}

// Update updates an existing region
func (rc *RegionController) Update(ctx context.Context) http.HandlerFunc {
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

		vars := mux.Vars(r)
		regionIDStr, ok := vars["id"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "bad_request",
				Message: "region ID is required",
			})
			return
		}

		regionID, err := uuid.Parse(regionIDStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_id",
				Message: "invalid region ID format",
			})
			return
		}

		var region game_entities.Region
		if err := json.NewDecoder(r.Body).Decode(&region); err != nil {
			slog.ErrorContext(r.Context(), "failed to decode request body", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_request",
				Message: fmt.Sprintf("invalid JSON: %v", err),
			})
			return
		}

		var updateRegionCmd game_in.UpdateRegionCommand
		if err := rc.Container.Resolve(&updateRegionCmd); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve UpdateRegionCommand", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		updatedRegion, err := updateRegionCmd.Execute(r.Context(), regionID, &region)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to update region", "error", err, "region_id", regionID)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "validation_error",
				Message: err.Error(),
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(updatedRegion)
	}
}

// Delete deletes a region
func (rc *RegionController) Delete(ctx context.Context) http.HandlerFunc {
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

		vars := mux.Vars(r)
		regionIDStr, ok := vars["id"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "bad_request",
				Message: "region ID is required",
			})
			return
		}

		regionID, err := uuid.Parse(regionIDStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_id",
				Message: "invalid region ID format",
			})
			return
		}

		var deleteRegionCmd game_in.DeleteRegionCommand
		if err := rc.Container.Resolve(&deleteRegionCmd); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve DeleteRegionCommand", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		if err := deleteRegionCmd.Execute(r.Context(), regionID); err != nil {
			slog.ErrorContext(r.Context(), "failed to delete region", "error", err, "region_id", regionID)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "delete_error",
				Message: err.Error(),
			})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// List lists all regions
func (rc *RegionController) List(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var searchRegionsQuery game_in.SearchRegionsQuery
		if err := rc.Container.Resolve(&searchRegionsQuery); err != nil {
			slog.ErrorContext(r.Context(), "failed to resolve SearchRegionsQuery", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to process request",
			})
			return
		}

		regions, err := searchRegionsQuery.Execute(r.Context())
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to list regions", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "internal_error",
				Message: "failed to retrieve regions",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(regions)
	}
}
