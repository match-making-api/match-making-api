package usecases

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/in"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/out"
)

type UpdateRegionUseCase struct {
	RegionWriter out.RegionWriter
	RegionReader out.RegionReader
}

func NewUpdateRegionUseCase(regionWriter out.RegionWriter, regionReader out.RegionReader) in.UpdateRegionCommand {
	return &UpdateRegionUseCase{
		RegionWriter: regionWriter,
		RegionReader: regionReader,
	}
}

func InjectUpdateRegion(c container.Container) error {
	c.Singleton(func(regionWriter out.RegionWriter, regionReader out.RegionReader) (in.UpdateRegionCommand, error) {
		return NewUpdateRegionUseCase(regionWriter, regionReader), nil
	})
	return nil
}

func (usecase *UpdateRegionUseCase) Execute(ctx context.Context, id uuid.UUID, region *game_entities.Region) (*game_entities.Region, error) {
	// Get existing region
	existingRegion, err := usecase.RegionReader.GetByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "region not found", "region_id", id, "error", err)
		return nil, fmt.Errorf("region not found: %w", err)
	}

	// Validate region data
	if err := validateRegion(region); err != nil {
		slog.ErrorContext(ctx, "region validation failed", "error", err, "region_id", id)
		return nil, fmt.Errorf("invalid region data: %w", err)
	}

	// Check if another region with the same name or slug already exists (except the current one)
	existingRegions, err := usecase.RegionReader.Search(ctx, nil)
	if err == nil {
		for _, existing := range existingRegions {
			if existing.ID != id && existing.Name == region.Name {
				slog.WarnContext(ctx, "region with same name already exists", "name", region.Name, "existing_id", existing.ID)
				return nil, fmt.Errorf("region with name '%s' already exists", region.Name)
			}
			if region.Slug != "" && existing.ID != id && existing.Slug == region.Slug {
				slog.WarnContext(ctx, "region with same slug already exists", "slug", region.Slug, "existing_id", existing.ID)
				return nil, fmt.Errorf("region with slug '%s' already exists", region.Slug)
			}
		}
	}

	// Audit log before update
	resourceOwner := common.GetResourceOwner(ctx)
	slog.InfoContext(ctx, "updating region", "region_id", id, "name", region.Name, "user_id", resourceOwner.UserID)

	// Update all fields
	existingRegion.Name = region.Name
	existingRegion.Description = region.Description
	existingRegion.Slug = region.Slug
	existingRegion.UpdatedAt = time.Now()

	// Update the region
	updatedRegion, err := usecase.RegionWriter.Update(ctx, existingRegion)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update region", "error", err, "region_id", id)
		return nil, fmt.Errorf("failed to update region: %w", err)
	}

	slog.InfoContext(ctx, "region updated successfully", "region_id", updatedRegion.ID, "name", updatedRegion.Name)

	return updatedRegion, nil
}
