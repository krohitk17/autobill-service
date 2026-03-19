package HttpPorts

import (
	"context"

	Dtos "autobill-service/internal/application/split/dtos"
	Helpers "autobill-service/pkg/helpers"

	"github.com/google/uuid"
)

type SplitUseCase interface {
	CreateSplit(ctx context.Context, userId uuid.UUID, input Dtos.CreateSplitInput) (*Dtos.SplitResult, error)
	GetSplit(ctx context.Context, userId, splitId uuid.UUID) (*Dtos.SplitResult, error)
	GetGroupSplits(ctx context.Context, userId, groupId uuid.UUID, pagination Helpers.PaginationParams) (*Dtos.SplitListResult, error)
	GetMySplits(ctx context.Context, userId uuid.UUID, pagination Helpers.PaginationParams) (*Dtos.SplitListResult, error)
	ReverseSplit(ctx context.Context, userId, splitId uuid.UUID) (*Dtos.SplitResult, error)
}
