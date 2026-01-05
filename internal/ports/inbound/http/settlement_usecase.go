package HttpPorts

import (
	"context"

	Dtos "autobill-service/internal/application/settlement/dtos"
	Helpers "autobill-service/pkg/helpers"

	"github.com/google/uuid"
)

type SettlementUseCase interface {
	CreateSettlement(ctx context.Context, userId uuid.UUID, input Dtos.CreateSettlementInput) (*Dtos.SettlementResult, error)
	GetPendingSettlements(ctx context.Context, userId uuid.UUID, pagination Helpers.PaginationParams) (*Dtos.SettlementListResult, error)
	GetSettlementHistory(ctx context.Context, userId uuid.UUID, pagination Helpers.PaginationParams) (*Dtos.SettlementListResult, error)
	ConfirmSettlement(ctx context.Context, userId, settlementId uuid.UUID) error
	DeleteSettlement(ctx context.Context, userId, settlementId uuid.UUID) error
}
