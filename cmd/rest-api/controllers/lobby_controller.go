package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/leet-gaming/match-making-api/pkg/common"
	"github.com/leet-gaming/match-making-api/pkg/domain/lobbies/entities"
	"github.com/leet-gaming/match-making-api/pkg/infra/db/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
)

type LobbyController struct {
	repo *mongodb.LobbyRepository
}

func NewLobbyController(mongoClient *mongo.Client, dbName string) *LobbyController {
	return &LobbyController{
		repo: mongodb.NewLobbyRepository(mongoClient, dbName),
	}
}

// CreateLobbyRequest represents the request to create a lobby
type CreateLobbyRequest struct {
	Name               string              `json:"name"`
	Description        string              `json:"description,omitempty"`
	GameID             string              `json:"game_id"`
	GameMode           string              `json:"game_mode"`
	Region             string              `json:"region"`
	Type               string              `json:"type"`
	Visibility         string              `json:"visibility"`
	MaxPlayers         int                 `json:"max_players"`
	MinPlayers         int                 `json:"min_players,omitempty"`
	RequiresReadyCheck bool                `json:"requires_ready_check,omitempty"`
	AllowSpectators    bool                `json:"allow_spectators,omitempty"`
	AllowCrossPlatform bool                `json:"allow_cross_platform,omitempty"`
	MapPool            []string            `json:"map_pool,omitempty"`
	Tags               []string            `json:"tags,omitempty"`
	SkillRange         *entities.SkillRange `json:"skill_range,omitempty"`
	MaxPing            int                 `json:"max_ping,omitempty"`
	EntryFeeCents      int                 `json:"entry_fee_cents,omitempty"`
	DistributionRule   string              `json:"distribution_rule,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// LobbyResponse wraps a lobby for API response
type LobbyResponse struct {
	Lobby *entities.Lobby `json:"lobby"`
}

// LobbiesResponse wraps multiple lobbies for API response
type LobbiesResponse struct {
	Lobbies []*entities.Lobby `json:"lobbies"`
	Total   int64             `json:"total"`
	HasMore bool              `json:"has_more"`
}

// Create creates a new lobby
// @Summary Create a new lobby
// @Tags lobbies
// @Accept json
// @Produce json
// @Param request body CreateLobbyRequest true "Create lobby request"
// @Success 201 {object} LobbyResponse
// @Router /api/lobbies [post]
func (c *LobbyController) Create(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req CreateLobbyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_request",
				Message: "failed to parse request body",
			})
			return
		}

		// Validate required fields
		if req.Name == "" || req.GameID == "" || req.MaxPlayers <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "validation_error",
				Message: "name, game_id, and max_players are required",
			})
			return
		}

		// SECURITY: Get creator ID from authenticated context, NOT from headers
		creatorID, ok := r.Context().Value(common.UserIDKey).(uuid.UUID)
		if !ok || creatorID == uuid.Nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "unauthorized",
				Message: "authentication required to create a lobby",
			})
			return
		}

		// Set defaults
		if req.MinPlayers <= 0 {
			req.MinPlayers = req.MaxPlayers
		}
		if req.Visibility == "" {
			req.Visibility = string(entities.LobbyVisibilityPublic)
		}
		if req.Type == "" {
			req.Type = string(entities.LobbyTypeCustom)
		}

		// Create lobby entity
		now := time.Now()
		lobby := &entities.Lobby{
			ID:                 uuid.New(),
			TenantID:           uuid.MustParse("00000000-0000-0000-0000-000000000001"), // Default tenant
			ClientID:           uuid.MustParse("00000000-0000-0000-0000-000000000001"), // Default client
			CreatorID:          creatorID,
			GameID:             req.GameID,
			GameMode:           req.GameMode,
			Region:             req.Region,
			Name:               req.Name,
			Description:        req.Description,
			Type:               entities.LobbyType(req.Type),
			Visibility:         entities.LobbyVisibility(req.Visibility),
			MaxPlayers:         req.MaxPlayers,
			MinPlayers:         req.MinPlayers,
			RequiresReadyCheck: req.RequiresReadyCheck,
			AllowSpectators:    req.AllowSpectators,
			AllowCrossPlatform: req.AllowCrossPlatform,
			MapPool:            req.MapPool,
			Tags:               req.Tags,
			SkillRange:         req.SkillRange,
			MaxPing:            req.MaxPing,
			Status:             entities.LobbyStatusOpen,
			CreatedAt:          now,
			UpdatedAt:          now,
			ExpiresAt:          now.Add(2 * time.Hour), // Default 2 hour expiry
			PlayerSlots:        make([]entities.PlayerSlot, 0),
			Metadata:           req.Metadata,
		}

		// Add prize pool if specified
		if req.EntryFeeCents > 0 {
			lobby.PrizePool = &entities.PrizePoolConfig{
				EntryFeeCents:    req.EntryFeeCents,
				DistributionRule: req.DistributionRule,
			}
		}

		// Add creator as first player
		lobby.PlayerSlots = append(lobby.PlayerSlots, entities.PlayerSlot{
			SlotNumber: 1,
			PlayerID:   &creatorID,
			IsReady:    false,
			JoinedAt:   now,
			Team:       1,
		})

		if err := c.repo.Create(r.Context(), lobby); err != nil {
			slog.ErrorContext(r.Context(), "failed to create lobby", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "create_failed",
				Message: "failed to create lobby",
			})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(LobbyResponse{Lobby: lobby})
	}
}

// Get retrieves a lobby by ID
// @Summary Get lobby by ID
// @Tags lobbies
// @Produce json
// @Param id path string true "Lobby ID"
// @Success 200 {object} LobbyResponse
// @Router /api/lobbies/{id} [get]
func (c *LobbyController) Get(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		vars := mux.Vars(r)
		lobbyID, err := uuid.Parse(vars["id"])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_id",
				Message: "invalid lobby ID format",
			})
			return
		}

		lobby, err := c.repo.GetByID(r.Context(), lobbyID)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to fetch lobby", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "fetch_failed",
				Message: "failed to fetch lobby",
			})
			return
		}
		if lobby == nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "not_found",
				Message: "lobby not found",
			})
			return
		}

		// Return public view for matchmaking lobbies
		if lobby.Visibility == entities.LobbyVisibilityMatchmaking {
			lobby = lobby.GetPublicView()
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(LobbyResponse{Lobby: lobby})
	}
}

// List retrieves lobbies with filters
// @Summary List lobbies
// @Tags lobbies
// @Produce json
// @Param game_id query string false "Filter by game ID"
// @Param game_mode query string false "Filter by game mode"
// @Param region query string false "Filter by region"
// @Param status query string false "Filter by status"
// @Param visibility query string false "Filter by visibility"
// @Param type query string false "Filter by type"
// @Param featured query bool false "Filter featured only"
// @Param q query string false "Text search"
// @Param skip query int false "Skip count"
// @Param limit query int false "Limit count"
// @Success 200 {object} LobbiesResponse
// @Router /api/lobbies [get]
func (c *LobbyController) List(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		query := r.URL.Query()
		params := mongodb.LobbySearchParams{
			GameID:     query.Get("game_id"),
			GameMode:   query.Get("game_mode"),
			Region:     query.Get("region"),
			Status:     query.Get("status"),
			Visibility: query.Get("visibility"),
			Type:       query.Get("type"),
			TextSearch: query.Get("q"),
		}

		if featured := query.Get("featured"); featured == "true" {
			f := true
			params.Featured = &f
		}

		if skip, err := strconv.Atoi(query.Get("skip")); err == nil {
			params.Skip = skip
		}
		if limit, err := strconv.Atoi(query.Get("limit")); err == nil {
			params.Limit = limit
		}

		lobbies, total, err := c.repo.Search(r.Context(), params)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to search lobbies", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "search_failed",
				Message: "failed to search lobbies",
			})
			return
		}

		// Sanitize matchmaking lobbies
		for i, lobby := range lobbies {
			if lobby.Visibility == entities.LobbyVisibilityMatchmaking {
				lobbies[i] = lobby.GetPublicView()
			}
		}

		hasMore := int64(params.Skip+len(lobbies)) < total

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(LobbiesResponse{
			Lobbies: lobbies,
			Total:   total,
			HasMore: hasMore,
		})
	}
}

// GetFeatured retrieves featured lobbies for homepage
// @Summary Get featured lobbies
// @Tags lobbies
// @Produce json
// @Param game_id query string false "Filter by game ID"
// @Param limit query int false "Limit count"
// @Success 200 {object} LobbiesResponse
// @Router /api/lobbies/featured [get]
func (c *LobbyController) GetFeatured(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		query := r.URL.Query()
		gameID := query.Get("game_id")
		limit := 8
		if l, err := strconv.Atoi(query.Get("limit")); err == nil && l > 0 {
			limit = l
		}

		lobbies, err := c.repo.GetFeaturedLobbies(r.Context(), gameID, limit)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to fetch featured lobbies", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "fetch_failed",
				Message: "failed to fetch featured lobbies",
			})
			return
		}

		// Sanitize matchmaking lobbies
		for i, lobby := range lobbies {
			if lobby.Visibility == entities.LobbyVisibilityMatchmaking {
				lobbies[i] = lobby.GetPublicView()
			}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(LobbiesResponse{
			Lobbies: lobbies,
			Total:   int64(len(lobbies)),
			HasMore: false,
		})
	}
}

// JoinLobbyRequest represents the request to join a lobby
type JoinLobbyRequest struct {
	PlayerID   string `json:"player_id,omitempty"`
	PlayerMMR  int    `json:"player_mmr,omitempty"`
	PlayerRank string `json:"player_rank,omitempty"`
}

// JoinLobbyResponse represents the response after joining
type JoinLobbyResponse struct {
	Lobby        *entities.Lobby     `json:"lobby"`
	AssignedSlot *entities.PlayerSlot `json:"assigned_slot"`
}

// Join allows a player to join a lobby
// @Summary Join a lobby
// @Tags lobbies
// @Accept json
// @Produce json
// @Param id path string true "Lobby ID"
// @Param request body JoinLobbyRequest true "Join request"
// @Success 200 {object} JoinLobbyResponse
// @Router /api/lobbies/{id}/join [post]
func (c *LobbyController) Join(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		vars := mux.Vars(r)
		lobbyID, err := uuid.Parse(vars["id"])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_id",
				Message: "invalid lobby ID format",
			})
			return
		}

		var req JoinLobbyRequest
		json.NewDecoder(r.Body).Decode(&req)

		// SECURITY: Get player ID from authenticated context, NOT from headers or body
		playerID, ok := r.Context().Value(common.UserIDKey).(uuid.UUID)
		if !ok || playerID == uuid.Nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "unauthorized",
				Message: "authentication required to join a lobby",
			})
			return
		}

		lobby, err := c.repo.GetByID(r.Context(), lobbyID)
		if err != nil || lobby == nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "not_found",
				Message: "lobby not found",
			})
			return
		}

		// Check if lobby is open
		if lobby.Status != entities.LobbyStatusOpen {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "lobby_not_open",
				Message: "lobby is not accepting players",
			})
			return
		}

		// Check if lobby is full
		if lobby.IsFull() {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "lobby_full",
				Message: "lobby is full",
			})
			return
		}

		// Check if player already in lobby
		if lobby.HasPlayer(playerID) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "already_joined",
				Message: "player already in lobby",
			})
			return
		}

		// Add player to lobby
		slotNumber := len(lobby.PlayerSlots) + 1
		newSlot := entities.PlayerSlot{
			SlotNumber: slotNumber,
			PlayerID:   &playerID,
			IsReady:    false,
			JoinedAt:   time.Now(),
			MMR:        req.PlayerMMR,
			Rank:       req.PlayerRank,
			Team:       (slotNumber % 2) + 1, // Simple team assignment
		}
		lobby.PlayerSlots = append(lobby.PlayerSlots, newSlot)

		if err := c.repo.Update(r.Context(), lobby); err != nil {
			slog.ErrorContext(r.Context(), "failed to join lobby", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "join_failed",
				Message: "failed to join lobby",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(JoinLobbyResponse{
			Lobby:        lobby,
			AssignedSlot: &newSlot,
		})
	}
}

// GetStats retrieves lobby statistics
// @Summary Get lobby statistics
// @Tags lobbies
// @Produce json
// @Param game_id query string false "Filter by game ID"
// @Success 200 {object} common.LobbyStats
// @Router /api/lobbies/stats [get]
func (c *LobbyController) GetStats(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		gameID := r.URL.Query().Get("game_id")

		stats, err := c.repo.GetLobbyStats(r.Context(), gameID)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to get lobby stats", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "stats_failed",
				Message: "failed to retrieve lobby statistics",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(stats)
	}
}

// Delete cancels/deletes a lobby
// @Summary Delete a lobby
// @Tags lobbies
// @Param id path string true "Lobby ID"
// @Success 204
// @Router /api/lobbies/{id} [delete]
func (c *LobbyController) Delete(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		lobbyID, err := uuid.Parse(vars["id"])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_id",
				Message: "invalid lobby ID format",
			})
			return
		}

		lobby, err := c.repo.GetByID(r.Context(), lobbyID)
		if err != nil || lobby == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// SECURITY: Verify the requester is the lobby creator or admin
		userID, ok := r.Context().Value(common.UserIDKey).(uuid.UUID)
		if !ok || userID == uuid.Nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "unauthorized",
				Message: "authentication required",
			})
			return
		}
		if lobby.CreatorID != userID && !common.IsAdmin(r.Context()) {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "forbidden",
				Message: "only the lobby creator can cancel this lobby",
			})
			return
		}

		// Update status to cancelled
		lobby.Status = entities.LobbyStatusCancelled
		c.repo.Update(r.Context(), lobby)

		w.WriteHeader(http.StatusNoContent)
	}
}

// SeedDemoLobbies creates sample lobbies for demonstration
// SECURITY: Only available in development/staging environments
func (c *LobbyController) SeedDemoLobbies(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// SECURITY: Block in production
		env := os.Getenv("APP_ENV")
		if env == "production" || env == "prod" {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "not_found",
				Message: "endpoint not available",
			})
			return
		}

		demoLobbies := []CreateLobbyRequest{
			{
				Name:        "🏆 Pro League Qualifier",
				Description: "Weekly qualifier for the Pro League. Top 3 advance!",
				GameID:      "cs2",
				GameMode:    "competitive",
				Region:      "na-east",
				Type:        "ranked",
				Visibility:  "public",
				MaxPlayers:  10,
				MinPlayers:  10,
				Tags:        []string{"tournament", "ranked", "pro"},
			},
			{
				Name:        "⚔️ Ranked 5v5 Competitive",
				Description: "Serious competitive matches. Bring your A-game!",
				GameID:      "cs2",
				GameMode:    "competitive",
				Region:      "eu-west",
				Type:        "ranked",
				Visibility:  "public",
				MaxPlayers:  10,
				MinPlayers:  6,
				Tags:        []string{"ranked", "competitive"},
			},
			{
				Name:        "🎮 Casual Practice",
				Description: "Chill practice session. All skill levels welcome!",
				GameID:      "cs2",
				GameMode:    "casual",
				Region:      "na-west",
				Type:        "casual",
				Visibility:  "public",
				MaxPlayers:  10,
				MinPlayers:  4,
				Tags:        []string{"casual", "practice", "friendly"},
			},
			{
				Name:        "💰 High Stakes 1v1",
				Description: "$50 entry fee, winner takes all!",
				GameID:      "cs2",
				GameMode:    "duel",
				Region:      "eu-central",
				Type:        "ranked",
				Visibility:  "public",
				MaxPlayers:  2,
				MinPlayers:  2,
				EntryFeeCents: 5000,
				DistributionRule: "winner_takes_all",
				Tags:        []string{"1v1", "high-stakes", "duel"},
			},
			{
				Name:        "🔥 Quick Match Queue",
				Description: "Auto-match with players of similar skill",
				GameID:      "cs2",
				GameMode:    "competitive",
				Region:      "global",
				Type:        "ranked",
				Visibility:  "matchmaking",
				MaxPlayers:  10,
				MinPlayers:  10,
				Tags:        []string{"quick", "matchmaking"},
			},
			{
				Name:        "🎯 Valorant Ranked",
				Description: "Competitive Valorant matches",
				GameID:      "valorant",
				GameMode:    "competitive",
				Region:      "na-east",
				Type:        "ranked",
				Visibility:  "public",
				MaxPlayers:  10,
				MinPlayers:  10,
				Tags:        []string{"valorant", "ranked"},
			},
		}

		created := make([]*entities.Lobby, 0, len(demoLobbies))
		now := time.Now()

		for i, req := range demoLobbies {
			creatorID := uuid.New()
			lobby := &entities.Lobby{
				ID:                 uuid.New(),
				TenantID:           uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				ClientID:           uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				CreatorID:          creatorID,
				GameID:             req.GameID,
				GameMode:           req.GameMode,
				Region:             req.Region,
				Name:               req.Name,
				Description:        req.Description,
				Type:               entities.LobbyType(req.Type),
				Visibility:         entities.LobbyVisibility(req.Visibility),
				IsFeatured:         i < 3, // First 3 are featured
				MaxPlayers:         req.MaxPlayers,
				MinPlayers:         req.MinPlayers,
				RequiresReadyCheck: true,
				AllowSpectators:    true,
				Tags:               req.Tags,
				Status:             entities.LobbyStatusOpen,
				CreatedAt:          now.Add(-time.Duration(i) * time.Minute),
				UpdatedAt:          now,
				ExpiresAt:          now.Add(2 * time.Hour),
				PlayerSlots:        make([]entities.PlayerSlot, 0),
			}

			if req.EntryFeeCents > 0 {
				lobby.PrizePool = &entities.PrizePoolConfig{
					EntryFeeCents:    req.EntryFeeCents,
					DistributionRule: req.DistributionRule,
				}
			}

			// Add some fake players
			numPlayers := (i % 5) + 2 // 2-6 players
			for j := 0; j < numPlayers && j < req.MaxPlayers; j++ {
				pid := uuid.New()
				lobby.PlayerSlots = append(lobby.PlayerSlots, entities.PlayerSlot{
					SlotNumber: j + 1,
					PlayerID:   &pid,
					PlayerName: fmt.Sprintf("Player%d", j+1),
					IsReady:    j%2 == 0,
					JoinedAt:   now.Add(-time.Duration(j) * time.Minute),
					MMR:        1500 + (j * 100),
					Team:       (j % 2) + 1,
				})
			}

			if err := c.repo.Create(ctx, lobby); err != nil {
				continue
			}
			created = append(created, lobby)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(LobbiesResponse{
			Lobbies: created,
			Total:   int64(len(created)),
		})
	}
}
