package RepositoryPorts

import (
	"context"

	Domain "autobill-service/internal/domain"

	"github.com/google/uuid"
)

type BalanceRepositoryPort interface {
	GetUserBalances(ctx context.Context, userId uuid.UUID) ([]Domain.UserBalance, error)
	GetUserBalancesWithOtherUser(ctx context.Context, userId, otherUserId uuid.UUID) ([]Domain.UserBalance, error)
	UpdateBalancesForSplit(ctx context.Context, split *Domain.Split, participants []Domain.SplitParticipant) error

	GetGroupBalances(ctx context.Context, groupId uuid.UUID) ([]Domain.GroupBalance, error)
	GetSimplifiedDebts(ctx context.Context, groupId uuid.UUID) ([]Domain.SimplifiedDebt, error)

	GetSplitsWithParticipants(ctx context.Context, groupId uuid.UUID) ([]Domain.Split, error)
	GetSettlementsForSplits(ctx context.Context, splitIDs []uuid.UUID) ([]Domain.Settlement, error)
	GetSettledParticipants(ctx context.Context, splitId uuid.UUID, userId uuid.UUID) (bool, error)
	ReplaceGroupBalances(ctx context.Context, groupId uuid.UUID, balances []Domain.GroupBalance) ([]Domain.GroupBalance, error)
}
