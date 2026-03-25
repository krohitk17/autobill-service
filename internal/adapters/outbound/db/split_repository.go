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

type SplitRepository struct {
	db DB.PostgresDB
}

func CreateSplitRepository(db DB.PostgresDB) RepositoryPorts.SplitRepositoryPort {
	return &SplitRepository{db: db}
}

func (repo *SplitRepository) GetSplitById(ctx context.Context, splitId uuid.UUID) (*Domain.Split, error) {
	var split Domain.Split
	if err := repo.db.DB.WithContext(ctx).First(&split, "id = ?", splitId).Error; err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, Errors.ErrSplitNotFound)
	}
	return &split, nil
}

func (repo *SplitRepository) GetSplitByIdempotencyKey(ctx context.Context, idempotencyKey string) (*Domain.Split, error) {
	var split Domain.Split
	if err := repo.db.DB.WithContext(ctx).Preload("Participants.User").Preload("CreatedBy").First(&split, "idempotency_key = ?", idempotencyKey).Error; err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, Errors.ErrSplitNotFound)
	}
	return &split, nil
}

func (repo *SplitRepository) GetSplitWithParticipants(ctx context.Context, splitId uuid.UUID) (*Domain.Split, error) {
	var split Domain.Split
	if err := repo.db.DB.WithContext(ctx).Preload("Participants.User").Preload("CreatedBy").First(&split, "id = ?", splitId).Error; err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, Errors.ErrSplitNotFound)
	}
	return &split, nil
}

func (repo *SplitRepository) GetSplitsByGroupId(ctx context.Context, groupId uuid.UUID, limit, offset int) ([]Domain.Split, int64, error) {
	var splits []Domain.Split
	var total int64

	query := repo.db.DB.WithContext(ctx).Model(&Domain.Split{}).Where("group_id = ?", groupId)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	if total == 0 {
		return []Domain.Split{}, 0, nil
	}

	if err := query.Preload("Participants.User").Order("created_at DESC").Limit(limit).Offset(offset).Find(&splits).Error; err != nil {
		return nil, 0, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}
	return splits, total, nil
}

func (repo *SplitRepository) GetSplitsByUserId(ctx context.Context, userId uuid.UUID, limit, offset int) ([]Domain.Split, int64, error) {
	var splits []Domain.Split
	var total int64

	subQuery := repo.db.DB.WithContext(ctx).
		Table("split_participants").
		Select("split_id").
		Where("user_id = ?", userId)

	baseQuery := repo.db.DB.WithContext(ctx).Model(&Domain.Split{}).
		Where("created_by_id = ? OR id IN (?)", userId, subQuery)

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	if total == 0 {
		return []Domain.Split{}, 0, nil
	}

	if err := baseQuery.
		Preload("Participants.User").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&splits).Error; err != nil {
		return nil, 0, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}
	return splits, total, nil
}

func (repo *SplitRepository) GetParticipant(ctx context.Context, splitId, userId uuid.UUID) (*Domain.SplitParticipant, error) {
	var participant Domain.SplitParticipant
	if err := repo.db.DB.WithContext(ctx).Preload("User").Where("split_id = ? AND user_id = ?", splitId, userId).First(&participant).Error; err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, Errors.ErrParticipantNotFound)
	}
	return &participant, nil
}

func (repo *SplitRepository) GetPendingSettlementCountBySplitId(ctx context.Context, splitId uuid.UUID) (int64, error) {
	var count int64
	if err := repo.db.DB.WithContext(ctx).
		Model(&Domain.Settlement{}).
		Where("split_id = ? AND confirmed = ?", splitId, false).
		Count(&count).Error; err != nil {
		return 0, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	return count, nil
}

func (repo *SplitRepository) GetConfirmedSettlementTotalsByPayer(ctx context.Context, splitId uuid.UUID) (map[uuid.UUID]int64, error) {
	type payerTotalRow struct {
		PayerID uuid.UUID
		Total   int64
	}

	rows := make([]payerTotalRow, 0)
	if err := repo.db.DB.WithContext(ctx).
		Model(&Domain.Settlement{}).
		Select("payer_id, COALESCE(SUM(amount), 0) AS total").
		Where("split_id = ? AND confirmed = ?", splitId, true).
		Group("payer_id").
		Scan(&rows).Error; err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	result := make(map[uuid.UUID]int64, len(rows))
	for _, row := range rows {
		result[row.PayerID] = row.Total
	}

	return result, nil
}

func (repo *SplitRepository) DeleteSplitWithBalanceRollback(ctx context.Context, split *Domain.Split, participants []Domain.SplitParticipant) error {
	tx := repo.db.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	inverseSplit := &Domain.Split{
		CreatedByID: split.CreatedByID,
		TotalAmount: -split.TotalAmount,
		Currency:    split.Currency,
		GroupID:     split.GroupID,
	}

	inverseParticipants := make([]Domain.SplitParticipant, len(participants))
	for i, p := range participants {
		inverseParticipants[i] = Domain.SplitParticipant{
			UserID:      p.UserID,
			ShareAmount: -p.ShareAmount,
			Currency:    p.Currency,
		}
	}

	confirmedSettlements := make([]Domain.Settlement, 0)
	if err := tx.Where("split_id = ? AND confirmed = ?", split.Id, true).Find(&confirmedSettlements).Error; err != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	for _, settlement := range confirmedSettlements {
		if err := repo.rollbackConfirmedSettlementBalanceTx(tx, split.GroupID, &settlement); err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := repo.applyBalanceUpdatesForSplitTx(tx, inverseSplit, inverseParticipants); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Delete(&Domain.Settlement{}, "split_id = ?", split.Id).Error; err != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	if err := tx.Delete(&Domain.Split{}, "id = ?", split.Id).Error; err != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	if err := tx.Commit().Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	return nil
}

func (repo *SplitRepository) rollbackConfirmedSettlementBalanceTx(tx *gorm.DB, groupID *uuid.UUID, settlement *Domain.Settlement) error {
	amount := settlement.Amount
	currency := settlement.Currency

	var payerToPayee Domain.UserBalance
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ? AND other_user_id = ? AND currency = ?", settlement.PayerID, settlement.PayeeID, currency).
		First(&payerToPayee).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			payerToPayee = Domain.UserBalance{UserID: settlement.PayerID, OtherUserID: settlement.PayeeID, NetAmount: 0, Currency: currency}
			if createErr := tx.Create(&payerToPayee).Error; createErr != nil {
				return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
			}
		} else {
			return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
		}
	}
	payerToPayee.NetAmount -= amount
	if err := tx.Save(&payerToPayee).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	var payeeToPayer Domain.UserBalance
	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ? AND other_user_id = ? AND currency = ?", settlement.PayeeID, settlement.PayerID, currency).
		First(&payeeToPayer).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			payeeToPayer = Domain.UserBalance{UserID: settlement.PayeeID, OtherUserID: settlement.PayerID, NetAmount: 0, Currency: currency}
			if createErr := tx.Create(&payeeToPayer).Error; createErr != nil {
				return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
			}
		} else {
			return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
		}
	}
	payeeToPayer.NetAmount += amount
	if err := tx.Save(&payeeToPayer).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	if groupID == nil {
		return nil
	}

	var payerGroupBalance Domain.GroupBalance
	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("group_id = ? AND user_id = ? AND currency = ?", *groupID, settlement.PayerID, currency).
		First(&payerGroupBalance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			payerGroupBalance = Domain.GroupBalance{GroupID: *groupID, UserID: settlement.PayerID, NetAmount: 0, Currency: currency}
			if createErr := tx.Create(&payerGroupBalance).Error; createErr != nil {
				return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
			}
		} else {
			return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
		}
	}
	payerGroupBalance.NetAmount -= amount
	if err := tx.Save(&payerGroupBalance).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	var payeeGroupBalance Domain.GroupBalance
	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("group_id = ? AND user_id = ? AND currency = ?", *groupID, settlement.PayeeID, currency).
		First(&payeeGroupBalance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			payeeGroupBalance = Domain.GroupBalance{GroupID: *groupID, UserID: settlement.PayeeID, NetAmount: 0, Currency: currency}
			if createErr := tx.Create(&payeeGroupBalance).Error; createErr != nil {
				return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
			}
		} else {
			return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
		}
	}
	payeeGroupBalance.NetAmount += amount
	if err := tx.Save(&payeeGroupBalance).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	return nil
}

func (repo *SplitRepository) CreateSplitWithParticipants(ctx context.Context, split *Domain.Split, participants []Domain.SplitParticipant) (*Domain.Split, []Domain.SplitParticipant, error) {
	tx := repo.db.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(split).Error; err != nil {
		tx.Rollback()
		return nil, nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	createdParticipants := make([]Domain.SplitParticipant, len(participants))
	for i := range participants {
		participants[i].SplitID = split.Id
		participants[i].SettledAmount = 0
		if err := tx.Create(&participants[i]).Error; err != nil {
			tx.Rollback()
			return nil, nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
		}

		if err := tx.Preload("User").First(&participants[i], "id = ?", participants[i].Id).Error; err != nil {
			tx.Rollback()
			return nil, nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
		}
		createdParticipants[i] = participants[i]
	}

	if err := repo.applyBalanceUpdatesForSplitTx(tx, split, createdParticipants); err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	return split, createdParticipants, nil
}

func (repo *SplitRepository) HasPendingSplitsInGroup(ctx context.Context, userId, groupId uuid.UUID) (bool, error) {
	var count int64

	err := repo.db.DB.WithContext(ctx).
		Model(&Domain.Split{}).
		Joins("JOIN split_participants ON split_participants.split_id = splits.id").
		Where("splits.group_id = ? AND split_participants.user_id = ?", groupId, userId).
		Count(&count).Error

	if err != nil {
		return false, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	return count > 0, nil
}

func (repo *SplitRepository) applyBalanceUpdatesForSplitTx(tx *gorm.DB, split *Domain.Split, participants []Domain.SplitParticipant) error {
	creatorId := split.CreatedByID
	currency := split.Currency

	for _, participant := range participants {
		if participant.UserID == creatorId {
			continue
		}

		var participantToCreator Domain.UserBalance
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND other_user_id = ? AND currency = ?", participant.UserID, creatorId, currency).
			First(&participantToCreator).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				participantToCreator = Domain.UserBalance{
					UserID:      participant.UserID,
					OtherUserID: creatorId,
					NetAmount:   0,
					Currency:    currency,
				}
				if createErr := tx.Create(&participantToCreator).Error; createErr != nil {
					return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
				}
			} else {
				return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
			}
		}
		participantToCreator.NetAmount -= participant.ShareAmount
		if err := tx.Save(&participantToCreator).Error; err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
		}

		var creatorToParticipant Domain.UserBalance
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND other_user_id = ? AND currency = ?", creatorId, participant.UserID, currency).
			First(&creatorToParticipant).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				creatorToParticipant = Domain.UserBalance{
					UserID:      creatorId,
					OtherUserID: participant.UserID,
					NetAmount:   0,
					Currency:    currency,
				}
				if createErr := tx.Create(&creatorToParticipant).Error; createErr != nil {
					return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
				}
			} else {
				return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
			}
		}
		creatorToParticipant.NetAmount += participant.ShareAmount
		if err := tx.Save(&creatorToParticipant).Error; err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
		}
	}

	if split.GroupID == nil {
		return nil
	}

	groupID := *split.GroupID
	for _, participant := range participants {
		netChange := int64(0)
		if participant.UserID == creatorId {
			netChange = split.TotalAmount - participant.ShareAmount
		} else {
			netChange = -participant.ShareAmount
		}

		var groupBalance Domain.GroupBalance
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("group_id = ? AND user_id = ? AND currency = ?", groupID, participant.UserID, currency).
			First(&groupBalance).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				groupBalance = Domain.GroupBalance{
					UserID:    participant.UserID,
					GroupID:   groupID,
					NetAmount: 0,
					Currency:  currency,
				}
				if createErr := tx.Create(&groupBalance).Error; createErr != nil {
					return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
				}
			} else {
				return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
			}
		}

		groupBalance.NetAmount += netChange
		if err := tx.Save(&groupBalance).Error; err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
		}
	}

	return nil
}
