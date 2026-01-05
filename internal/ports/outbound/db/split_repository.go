package RepositoryPorts

import (
	"context"

	Domain "autobill-service/internal/domain"

	"github.com/google/uuid"
)

type SplitRepositoryPort interface {
	CreateSplitWithParticipants(ctx context.Context, split *Domain.Split, participants []Domain.SplitParticipant) (*Domain.Split, []Domain.SplitParticipant, error)
	GetSplitById(ctx context.Context, splitId uuid.UUID) (*Domain.Split, error)
	GetSplitWithParticipants(ctx context.Context, splitId uuid.UUID) (*Domain.Split, error)
	GetSplitsByGroupId(ctx context.Context, groupId uuid.UUID, limit, offset int) ([]Domain.Split, int64, error)
	GetSplitsByUserId(ctx context.Context, userId uuid.UUID, limit, offset int) ([]Domain.Split, int64, error)
	DeleteSplit(ctx context.Context, splitId uuid.UUID) error
	IsSplitReversed(ctx context.Context, splitId uuid.UUID) (bool, error)

	AddParticipant(ctx context.Context, participant *Domain.SplitParticipant) (*Domain.SplitParticipant, error)
	GetParticipant(ctx context.Context, splitId, userId uuid.UUID) (*Domain.SplitParticipant, error)
	UpdateParticipant(ctx context.Context, participant *Domain.SplitParticipant) error

	FinalizeSplit(ctx context.Context, splitId uuid.UUID) error
	CreateReversalSplitWithParticipants(ctx context.Context, originalSplitId uuid.UUID, reversalSplit *Domain.Split, participants []Domain.SplitParticipant) (*Domain.Split, []Domain.SplitParticipant, error)
	HasPendingSplitsInGroup(ctx context.Context, userId, groupId uuid.UUID) (bool, error)
}
