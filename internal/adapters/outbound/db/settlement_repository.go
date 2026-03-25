package RepositoryAdapters

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"

	Domain "autobill-service/internal/domain"
	DB "autobill-service/internal/infrastructure/db"
	RepositoryPorts "autobill-service/internal/ports/outbound/db"
	Errors "autobill-service/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
		Where("(settlements.payer_id = ? OR settlements.payee_id = ?) AND settlements.confirmed = ?", userId, userId, false)

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	if err := repo.db.DB.WithContext(ctx).Preload("Payer").Preload("Payee").Preload("Split").
		Where("(settlements.payer_id = ? OR settlements.payee_id = ?) AND settlements.confirmed = ?", userId, userId, false).
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
		for _, s := range settlements {
			confirmedMap[s.Id] = s.Confirmed
		}
	}

	return settlements, confirmedMap, total, nil
}

func (repo *SettlementRepository) ConfirmSettlement(ctx context.Context, settlementId uuid.UUID) error {
	tx := repo.db.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var settlement Domain.Settlement
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&settlement, "id = ?", settlementId).Error; err != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusNotFound, Errors.ErrSettlementNotFound)
	}

	if settlement.Confirmed {
		tx.Rollback()
		return fiber.NewError(fiber.StatusBadRequest, Errors.ErrSettlementAlreadyConfirmed)
	}

	var participant Domain.SplitParticipant
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("split_id = ? AND user_id = ?", settlement.SplitID, settlement.PayerID).First(&participant).Error; err != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusNotFound, Errors.ErrParticipantNotFound)
	}

	remaining := participant.ShareAmount - participant.SettledAmount
	if remaining <= 0 {
		tx.Rollback()
		return fiber.NewError(fiber.StatusBadRequest, Errors.ErrSettlementAlreadyConfirmed)
	}
	if settlement.Amount > remaining {
		tx.Rollback()
		return fiber.NewError(fiber.StatusBadRequest, Errors.ErrInvalidSettlementAmount)
	}

	participant.SettledAmount += settlement.Amount
	participant.IsSettled = participant.SettledAmount >= participant.ShareAmount
	if err := tx.Save(&participant).Error; err != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	settlement.Confirmed = true
	if err := tx.Model(&Domain.Settlement{}).Where("id = ?", settlementId).Update("confirmed", true).Error; err != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	if err := repo.applySettlementBalanceUpdatesTx(tx, &settlement); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	return nil
}

func (repo *SettlementRepository) IsSettlementConfirmed(ctx context.Context, settlementId uuid.UUID) (bool, error) {
	var settlement Domain.Settlement
	if err := repo.db.DB.WithContext(ctx).First(&settlement, "id = ?", settlementId).Error; err != nil {
		return false, fiber.NewError(fiber.StatusNotFound, Errors.ErrSettlementNotFound)
	}

	return settlement.Confirmed, nil
}

func (repo *SettlementRepository) applySettlementBalanceUpdatesTx(tx *gorm.DB, settlement *Domain.Settlement) error {
	amount := settlement.Amount
	currency := settlement.Currency

	var payerToPayee Domain.UserBalance
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ? AND other_user_id = ? AND currency = ?", settlement.PayerID, settlement.PayeeID, currency).
		First(&payerToPayee).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			payerToPayee = Domain.UserBalance{
				UserID:      settlement.PayerID,
				OtherUserID: settlement.PayeeID,
				NetAmount:   0,
				Currency:    currency,
			}
			if createErr := tx.Create(&payerToPayee).Error; createErr != nil {
				return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
			}
		} else {
			return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
		}
	}
	payerToPayee.NetAmount += amount
	if err := tx.Save(&payerToPayee).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	var payeeToPayer Domain.UserBalance
	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ? AND other_user_id = ? AND currency = ?", settlement.PayeeID, settlement.PayerID, currency).
		First(&payeeToPayer).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			payeeToPayer = Domain.UserBalance{
				UserID:      settlement.PayeeID,
				OtherUserID: settlement.PayerID,
				NetAmount:   0,
				Currency:    currency,
			}
			if createErr := tx.Create(&payeeToPayer).Error; createErr != nil {
				return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
			}
		} else {
			return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
		}
	}
	payeeToPayer.NetAmount -= amount
	if err := tx.Save(&payeeToPayer).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	var split Domain.Split
	if err := tx.Select("id", "group_id").First(&split, "id = ?", settlement.SplitID).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	if split.GroupID == nil {
		return nil
	}

	groupID := *split.GroupID

	var payerGroupBalance Domain.GroupBalance
	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("group_id = ? AND user_id = ? AND currency = ?", groupID, settlement.PayerID, currency).
		First(&payerGroupBalance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			payerGroupBalance = Domain.GroupBalance{
				GroupID:   groupID,
				UserID:    settlement.PayerID,
				NetAmount: 0,
				Currency:  currency,
			}
			if createErr := tx.Create(&payerGroupBalance).Error; createErr != nil {
				return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
			}
		} else {
			return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
		}
	}
	payerGroupBalance.NetAmount += amount
	if err := tx.Save(&payerGroupBalance).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	var payeeGroupBalance Domain.GroupBalance
	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("group_id = ? AND user_id = ? AND currency = ?", groupID, settlement.PayeeID, currency).
		First(&payeeGroupBalance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			payeeGroupBalance = Domain.GroupBalance{
				GroupID:   groupID,
				UserID:    settlement.PayeeID,
				NetAmount: 0,
				Currency:  currency,
			}
			if createErr := tx.Create(&payeeGroupBalance).Error; createErr != nil {
				return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
			}
		} else {
			return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
		}
	}
	payeeGroupBalance.NetAmount -= amount
	if err := tx.Save(&payeeGroupBalance).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	return nil
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
