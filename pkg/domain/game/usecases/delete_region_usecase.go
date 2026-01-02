package usecases

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/in"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/ports/out"
)

type DeleteRegionUseCase struct {
	RegionWriter out.RegionWriter
	RegionReader out.RegionReader
}

func NewDeleteRegionUseCase(regionWriter out.RegionWriter, regionReader out.RegionReader) in.DeleteRegionCommand {
	return &DeleteRegionUseCase{
		RegionWriter: regionWriter,
		RegionReader: regionReader,
	}
}

func InjectDeleteRegion(c container.Container) error {
	c.Singleton(func(regionWriter out.RegionWriter, regionReader out.RegionReader) (in.DeleteRegionCommand, error) {
		return NewDeleteRegionUseCase(regionWriter, regionReader), nil
	})
	return nil
}

func (usecase *DeleteRegionUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	// Check if the region exists
	existingRegion, err := usecase.RegionReader.GetByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "region not found for deletion", "region_id", id, "error", err)
		return fmt.Errorf("region not found: %w", err)
	}

	// Audit log before deletion
	resourceOwner := common.GetResourceOwner(ctx)
	slog.InfoContext(ctx, "deleting region", "region_id", id, "name", existingRegion.Name, "user_id", resourceOwner.UserID)

	// Delete the region
	err = usecase.RegionWriter.Delete(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to delete region", "error", err, "region_id", id)
		return fmt.Errorf("failed to delete region: %w", err)
	}

	slog.InfoContext(ctx, "region deleted successfully", "region_id", id, "name", existingRegion.Name)

	return nil
}
