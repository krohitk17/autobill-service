package RepositoryPorts

import (
	"context"

	Domain "autobill-service/internal/domain"

	"github.com/google/uuid"
)

type BalanceRepositoryPort interface {
	GetUserBalances(ctx context.Context, userId uuid.UUID) ([]Domain.UserBalance, error)
	GetOrCreateUserBalance(ctx context.Context, userId, otherUserId uuid.UUID, currency Domain.Currency) (*Domain.UserBalance, error)
	UpdateUserBalance(ctx context.Context, balance *Domain.UserBalance) error
	UpdateBalancesForSplit(ctx context.Context, split *Domain.Split, participants []Domain.SplitParticipant) error

	GetGroupBalances(ctx context.Context, groupId uuid.UUID) ([]Domain.GroupBalance, error)
	GetOrCreateGroupBalance(ctx context.Context, userId, groupId uuid.UUID, currency Domain.Currency) (*Domain.GroupBalance, error)
	UpdateGroupBalance(ctx context.Context, balance *Domain.GroupBalance) error
	GetSimplifiedDebts(ctx context.Context, groupId uuid.UUID) ([]Domain.SimplifiedDebt, error)

	GetFinalizedSplitsWithParticipants(ctx context.Context, groupId uuid.UUID) ([]Domain.Split, error)
	GetSettlementsForSplits(ctx context.Context, splitIDs []uuid.UUID) ([]Domain.Settlement, error)
	GetSettledParticipants(ctx context.Context, splitId uuid.UUID, userId uuid.UUID) (bool, error)
	ReplaceGroupBalances(ctx context.Context, groupId uuid.UUID, balances []Domain.GroupBalance) ([]Domain.GroupBalance, error)
}
