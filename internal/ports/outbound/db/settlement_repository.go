package RepositoryPorts

import (
	"context"

	Domain "autobill-service/internal/domain"

	"github.com/google/uuid"
)

type SettlementRepositoryPort interface {
	CreateSettlement(ctx context.Context, settlement *Domain.Settlement) (*Domain.Settlement, error)
	GetSettlementById(ctx context.Context, settlementId uuid.UUID) (*Domain.Settlement, error)
	GetSettlementByIdempotencyKey(ctx context.Context, idempotencyKey string) (*Domain.Settlement, error)
	GetPendingSettlementsByUserId(ctx context.Context, userId uuid.UUID, limit, offset int) ([]Domain.Settlement, int64, error)
	GetSettlementHistoryWithConfirmation(ctx context.Context, userId uuid.UUID, limit, offset int) ([]Domain.Settlement, map[uuid.UUID]bool, int64, error)
	ConfirmSettlement(ctx context.Context, settlementId uuid.UUID) error
	IsSettlementConfirmed(ctx context.Context, settlementId uuid.UUID) (bool, error)
	DeleteSettlement(ctx context.Context, settlementId uuid.UUID) error
}
