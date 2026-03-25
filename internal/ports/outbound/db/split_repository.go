package RepositoryPorts

import (
	"context"

	Domain "autobill-service/internal/domain"

	"github.com/google/uuid"
)

type SplitRepositoryPort interface {
	CreateSplitWithParticipants(ctx context.Context, split *Domain.Split, participants []Domain.SplitParticipant) (*Domain.Split, []Domain.SplitParticipant, error)
	GetSplitById(ctx context.Context, splitId uuid.UUID) (*Domain.Split, error)
	GetSplitByIdempotencyKey(ctx context.Context, idempotencyKey string) (*Domain.Split, error)
	GetSplitWithParticipants(ctx context.Context, splitId uuid.UUID) (*Domain.Split, error)
	GetSplitsByGroupId(ctx context.Context, groupId uuid.UUID, limit, offset int) ([]Domain.Split, int64, error)
	GetSplitsByUserId(ctx context.Context, userId uuid.UUID, limit, offset int) ([]Domain.Split, int64, error)
	GetParticipant(ctx context.Context, splitId, userId uuid.UUID) (*Domain.SplitParticipant, error)
	GetPendingSettlementCountBySplitId(ctx context.Context, splitId uuid.UUID) (int64, error)
	GetConfirmedSettlementTotalsByPayer(ctx context.Context, splitId uuid.UUID) (map[uuid.UUID]int64, error)
	DeleteSplitWithBalanceRollback(ctx context.Context, split *Domain.Split, participants []Domain.SplitParticipant) error
	HasPendingSplitsInGroup(ctx context.Context, userId, groupId uuid.UUID) (bool, error)
}
