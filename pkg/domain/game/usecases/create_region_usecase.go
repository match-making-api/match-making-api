package usecases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golobby/container/v3"
	"github.com/leet-gaming/match-making-api/pkg/common"
	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/in"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/out"
)

type CreateRegionUseCase struct {
	RegionWriter out.RegionWriter
	RegionReader out.RegionReader
}

func NewCreateRegionUseCase(regionWriter out.RegionWriter, regionReader out.RegionReader) in.CreateRegionCommand {
	return &CreateRegionUseCase{
		RegionWriter: regionWriter,
		RegionReader: regionReader,
	}
}

func InjectCreateRegion(c container.Container) error {
	c.Singleton(func(regionWriter out.RegionWriter, regionReader out.RegionReader) (in.CreateRegionCommand, error) {
		return NewCreateRegionUseCase(regionWriter, regionReader), nil
	})
	return nil
}

func (usecase *CreateRegionUseCase) Execute(ctx context.Context, region *game_entities.Region) (*game_entities.Region, error) {
	// Validate region data
	if err := validateRegion(region); err != nil {
		slog.ErrorContext(ctx, "region validation failed", "error", err)
		return nil, fmt.Errorf("invalid region data: %w", err)
	}

	// Check if a region with the same name or slug already exists
	existingRegions, err := usecase.RegionReader.Search(ctx, nil)
	if err == nil {
		for _, existing := range existingRegions {
			if existing.Name == region.Name {
				slog.WarnContext(ctx, "region with same name already exists", "name", region.Name)
				return nil, fmt.Errorf("region with name '%s' already exists", region.Name)
			}
			if region.Slug != "" && existing.Slug == region.Slug {
				slog.WarnContext(ctx, "region with same slug already exists", "slug", region.Slug)
				return nil, fmt.Errorf("region with slug '%s' already exists", region.Slug)
			}
		}
	} else {
		// If there's an error in the search, just log but continue
		slog.WarnContext(ctx, "failed to search existing regions", "error", err)
	}

	// Create base entity
	resourceOwner := common.GetResourceOwner(ctx)
	baseEntity := common.NewEntity(resourceOwner)
	region.BaseEntity = baseEntity

	// Audit log
	slog.InfoContext(ctx, "creating new region", "name", region.Name, "slug", region.Slug, "user_id", resourceOwner.UserID)

	// Create the region (repository will create the ID automatically)
	createdRegion, err := usecase.RegionWriter.Create(ctx, region)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create region", "error", err, "name", region.Name)
		return nil, fmt.Errorf("failed to create region: %w", err)
	}

	slog.InfoContext(ctx, "region created successfully", "region_id", createdRegion.ID, "name", createdRegion.Name)

	return createdRegion, nil
}

// validateRegion validates region data before creating or updating
func validateRegion(region *game_entities.Region) error {
	if region.Name == "" {
		return errors.New("region name is required")
	}

	if len(region.Name) > 100 {
		return errors.New("region name must be 100 characters or less")
	}

	if region.Slug != "" && len(region.Slug) > 50 {
		return errors.New("region slug must be 50 characters or less")
	}

	return nil
}
