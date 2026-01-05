package RepositoryAdapters

import (
	"context"

	"github.com/gofiber/fiber/v2"

	Domain "autobill-service/internal/domain"
	DB "autobill-service/internal/infrastructure/db"
	RepositoryPorts "autobill-service/internal/ports/outbound/db"
	Errors "autobill-service/pkg/errors"

	"github.com/google/uuid"
)

type SettlementRepository struct {
	db DB.PostgresDB
}

func CreateSettlementRepository(db DB.PostgresDB) RepositoryPorts.SettlementRepositoryPort {
	return &SettlementRepository{db: db}
}

func (repo *SettlementRepository) CreateSettlement(ctx context.Context, settlement *Domain.Settlement) (*Domain.Settlement, error) {
	if err := repo.db.DB.WithContext(ctx).Create(settlement).Error; err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	if err := repo.db.DB.WithContext(ctx).Preload("Payer").Preload("Payee").First(settlement, "id = ?", settlement.Id).Error; err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	return settlement, nil
}

func (repo *SettlementRepository) GetSettlementById(ctx context.Context, settlementId uuid.UUID) (*Domain.Settlement, error) {
	var settlement Domain.Settlement
	if err := repo.db.DB.WithContext(ctx).First(&settlement, "id = ?", settlementId).Error; err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, Errors.ErrSettlementNotFound)
	}
	return &settlement, nil
}

func (repo *SettlementRepository) GetSettlementByIdempotencyKey(ctx context.Context, idempotencyKey string) (*Domain.Settlement, error) {
	var settlement Domain.Settlement
	if err := repo.db.DB.WithContext(ctx).Preload("Payer").Preload("Payee").First(&settlement, "idempotency_key = ?", idempotencyKey).Error; err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, Errors.ErrSettlementNotFound)
	}
	return &settlement, nil
}

func (repo *SettlementRepository) GetPendingSettlementsByUserId(ctx context.Context, userId uuid.UUID, limit, offset int) ([]Domain.Settlement, int64, error) {
	var settlements []Domain.Settlement
	var total int64

	baseQuery := repo.db.DB.WithContext(ctx).Model(&Domain.Settlement{}).
		Joins("JOIN split_participants ON split_participants.split_id = settlements.split_id AND split_participants.user_id = settlements.payer_id").
		Where("(settlements.payer_id = ? OR settlements.payee_id = ?) AND split_participants.is_settled = ?", userId, userId, false)

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	if err := repo.db.DB.WithContext(ctx).Preload("Payer").Preload("Payee").Preload("Split").
		Joins("JOIN split_participants ON split_participants.split_id = settlements.split_id AND split_participants.user_id = settlements.payer_id").
		Where("(settlements.payer_id = ? OR settlements.payee_id = ?) AND split_participants.is_settled = ?", userId, userId, false).
		Order("date DESC").
		Limit(limit).Offset(offset).
		Find(&settlements).Error; err != nil {
		return nil, 0, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}
	return settlements, total, nil
}

func (repo *SettlementRepository) GetSettlementHistoryWithConfirmation(ctx context.Context, userId uuid.UUID, limit, offset int) ([]Domain.Settlement, map[uuid.UUID]bool, int64, error) {
	var settlements []Domain.Settlement
	var total int64

	if err := repo.db.DB.WithContext(ctx).Model(&Domain.Settlement{}).Where("payer_id = ? OR payee_id = ?", userId, userId).Count(&total).Error; err != nil {
		return nil, nil, 0, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	if err := repo.db.DB.WithContext(ctx).Preload("Payer").Preload("Payee").Preload("Split").
		Where("payer_id = ? OR payee_id = ?", userId, userId).
		Order("date DESC").
		Limit(limit).Offset(offset).
		Find(&settlements).Error; err != nil {
		return nil, nil, 0, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	confirmedMap := make(map[uuid.UUID]bool)
	if len(settlements) > 0 {
		var conditions [][]interface{}
		for _, s := range settlements {
			conditions = append(conditions, []interface{}{s.SplitID, s.PayerID})
		}

		var participants []Domain.SplitParticipant
		query := repo.db.DB.WithContext(ctx).Model(&Domain.SplitParticipant{})
		for i, cond := range conditions {
			if i == 0 {
				query = query.Where("(split_id = ? AND user_id = ?)", cond[0], cond[1])
			} else {
				query = query.Or("(split_id = ? AND user_id = ?)", cond[0], cond[1])
			}
		}
		if err := query.Find(&participants).Error; err != nil {
			return nil, nil, 0, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
		}

		participantMap := make(map[string]bool)
		for _, p := range participants {
			key := p.SplitID.String() + "-" + p.UserID.String()
			participantMap[key] = p.IsSettled
		}

		for _, s := range settlements {
			key := s.SplitID.String() + "-" + s.PayerID.String()
			confirmedMap[s.Id] = participantMap[key]
		}
	}

	return settlements, confirmedMap, total, nil
}

func (repo *SettlementRepository) ConfirmSettlement(ctx context.Context, settlementId uuid.UUID) error {
	var settlement Domain.Settlement
	if err := repo.db.DB.WithContext(ctx).First(&settlement, "id = ?", settlementId).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, Errors.ErrSettlementNotFound)
	}

	result := repo.db.DB.WithContext(ctx).Model(&Domain.SplitParticipant{}).
		Where("split_id = ? AND user_id = ?", settlement.SplitID, settlement.PayerID).
		Update("is_settled", true)
	if result.Error != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	return nil
}

func (repo *SettlementRepository) IsSettlementConfirmed(ctx context.Context, settlementId uuid.UUID) (bool, error) {
	var settlement Domain.Settlement
	if err := repo.db.DB.WithContext(ctx).First(&settlement, "id = ?", settlementId).Error; err != nil {
		return false, fiber.NewError(fiber.StatusNotFound, Errors.ErrSettlementNotFound)
	}

	var participant Domain.SplitParticipant
	if err := repo.db.DB.WithContext(ctx).Where("split_id = ? AND user_id = ?", settlement.SplitID, settlement.PayerID).First(&participant).Error; err != nil {
		return false, nil
	}

	return participant.IsSettled, nil
}

func (repo *SettlementRepository) DeleteSettlement(ctx context.Context, settlementId uuid.UUID) error {
	result := repo.db.DB.WithContext(ctx).Delete(&Domain.Settlement{}, "id = ?", settlementId)
	if result.Error != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}
	if result.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusNotFound, Errors.ErrSettlementNotFound)
	}
	return nil
}
