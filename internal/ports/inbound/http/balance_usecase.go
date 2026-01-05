package HttpPorts

import (
	Dtos "autobill-service/internal/application/balance/dtos"
	"context"

	"github.com/google/uuid"
)

type BalanceUseCase interface {
	GetMyBalance(ctx context.Context, userId uuid.UUID) (*Dtos.UserBalanceResult, error)

	GetGroupBalance(ctx context.Context, userId, groupId uuid.UUID) (*Dtos.GroupBalanceResult, error)
	RecalculateGroupBalance(ctx context.Context, userId, groupId uuid.UUID) (*Dtos.GroupBalanceResult, error)
	GetSimplifiedDebts(ctx context.Context, userId, groupId uuid.UUID) (*Dtos.SimplifiedDebtsResult, error)
}
