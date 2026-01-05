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
)

type BalanceRepository struct {
	db DB.PostgresDB
}

func CreateBalanceRepository(db DB.PostgresDB) RepositoryPorts.BalanceRepositoryPort {
	return &BalanceRepository{db: db}
}

func (repo *BalanceRepository) GetUserBalances(ctx context.Context, userId uuid.UUID) ([]Domain.UserBalance, error) {
	var balances []Domain.UserBalance
	if err := repo.db.DB.WithContext(ctx).Preload("OtherUser").Where("user_id = ?", userId).Find(&balances).Error; err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}
	return balances, nil
}

func (repo *BalanceRepository) GetOrCreateUserBalance(ctx context.Context, userId, otherUserId uuid.UUID, currency Domain.Currency) (*Domain.UserBalance, error) {
	var balance Domain.UserBalance
	err := repo.db.DB.WithContext(ctx).Preload("OtherUser").Where("user_id = ? AND other_user_id = ? AND currency = ?", userId, otherUserId, currency).First(&balance).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
		}
		balance = Domain.UserBalance{
			UserID:      userId,
			OtherUserID: otherUserId,
			NetAmount:   0,
			Currency:    currency,
		}
		if err := repo.db.DB.WithContext(ctx).Create(&balance).Error; err != nil {
			return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
		}
		if err := repo.db.DB.WithContext(ctx).Preload("OtherUser").First(&balance, "id = ?", balance.Id).Error; err != nil {
			return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
		}
	}
	return &balance, nil
}

func (repo *BalanceRepository) UpdateUserBalance(ctx context.Context, balance *Domain.UserBalance) error {
	if err := repo.db.DB.WithContext(ctx).Save(balance).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}
	return nil
}

func (repo *BalanceRepository) UpdateBalancesForSplit(ctx context.Context, split *Domain.Split, participants []Domain.SplitParticipant) error {
	tx := repo.db.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	creatorId := split.CreatedByID
	currency := split.Currency

	userIds := make([]uuid.UUID, 0, len(participants))
	for _, p := range participants {
		userIds = append(userIds, p.UserID)
	}

	var existingUserBalances []Domain.UserBalance
	if err := tx.Where("currency = ? AND ((user_id IN ? AND other_user_id = ?) OR (user_id = ? AND other_user_id IN ?))",
		currency, userIds, creatorId, creatorId, userIds).Find(&existingUserBalances).Error; err != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	userBalanceMap := make(map[string]*Domain.UserBalance)
	for i := range existingUserBalances {
		key := existingUserBalances[i].UserID.String() + "-" + existingUserBalances[i].OtherUserID.String()
		userBalanceMap[key] = &existingUserBalances[i]
	}

	var newBalances []Domain.UserBalance
	var updateBalances []*Domain.UserBalance

	for _, participant := range participants {
		if participant.UserID == creatorId {
			continue
		}

		participantKey := participant.UserID.String() + "-" + creatorId.String()
		if balance, exists := userBalanceMap[participantKey]; exists {
			balance.NetAmount -= participant.ShareAmount
			updateBalances = append(updateBalances, balance)
		} else {
			newBalances = append(newBalances, Domain.UserBalance{
				UserID:      participant.UserID,
				OtherUserID: creatorId,
				NetAmount:   -participant.ShareAmount,
				Currency:    currency,
			})
		}

		creatorKey := creatorId.String() + "-" + participant.UserID.String()
		if balance, exists := userBalanceMap[creatorKey]; exists {
			balance.NetAmount += participant.ShareAmount
			updateBalances = append(updateBalances, balance)
		} else {
			newBalances = append(newBalances, Domain.UserBalance{
				UserID:      creatorId,
				OtherUserID: participant.UserID,
				NetAmount:   participant.ShareAmount,
				Currency:    currency,
			})
		}
	}

	if len(newBalances) > 0 {
		if err := tx.Create(&newBalances).Error; err != nil {
			tx.Rollback()
			return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
		}
	}

	for _, balance := range updateBalances {
		if err := tx.Save(balance).Error; err != nil {
			tx.Rollback()
			return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
		}
	}

	if split.GroupID != nil {
		groupId := *split.GroupID

		var existingGroupBalances []Domain.GroupBalance
		if err := tx.Where("group_id = ? AND currency = ? AND user_id IN ?", groupId, currency, userIds).Find(&existingGroupBalances).Error; err != nil {
			tx.Rollback()
			return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
		}

		groupBalanceMap := make(map[uuid.UUID]*Domain.GroupBalance)
		for i := range existingGroupBalances {
			groupBalanceMap[existingGroupBalances[i].UserID] = &existingGroupBalances[i]
		}

		var newGroupBalances []Domain.GroupBalance
		var updateGroupBalances []*Domain.GroupBalance

		for _, participant := range participants {
			netChange := int64(0)
			if participant.UserID == creatorId {
				netChange = split.TotalAmount - participant.ShareAmount
			} else {
				netChange = -participant.ShareAmount
			}

			if balance, exists := groupBalanceMap[participant.UserID]; exists {
				balance.NetAmount += netChange
				updateGroupBalances = append(updateGroupBalances, balance)
			} else {
				newGroupBalances = append(newGroupBalances, Domain.GroupBalance{
					UserID:    participant.UserID,
					GroupID:   groupId,
					NetAmount: netChange,
					Currency:  currency,
				})
			}
		}

		if len(newGroupBalances) > 0 {
			if err := tx.Create(&newGroupBalances).Error; err != nil {
				tx.Rollback()
				return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
			}
		}

		for _, balance := range updateGroupBalances {
			if err := tx.Save(balance).Error; err != nil {
				tx.Rollback()
				return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	return nil
}

func (repo *BalanceRepository) GetGroupBalances(ctx context.Context, groupId uuid.UUID) ([]Domain.GroupBalance, error) {
	var balances []Domain.GroupBalance
	if err := repo.db.DB.WithContext(ctx).Preload("User").Where("group_id = ?", groupId).Find(&balances).Error; err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}
	return balances, nil
}

func (repo *BalanceRepository) GetOrCreateGroupBalance(ctx context.Context, userId, groupId uuid.UUID, currency Domain.Currency) (*Domain.GroupBalance, error) {
	var balance Domain.GroupBalance
	err := repo.db.DB.WithContext(ctx).Preload("User").Where("user_id = ? AND group_id = ? AND currency = ?", userId, groupId, currency).First(&balance).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
		}
		balance = Domain.GroupBalance{
			UserID:    userId,
			GroupID:   groupId,
			NetAmount: 0,
			Currency:  currency,
		}
		if err := repo.db.DB.WithContext(ctx).Create(&balance).Error; err != nil {
			return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
		}
		if err := repo.db.DB.WithContext(ctx).Preload("User").First(&balance, "id = ?", balance.Id).Error; err != nil {
			return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
		}
	}
	return &balance, nil
}

func (repo *BalanceRepository) UpdateGroupBalance(ctx context.Context, balance *Domain.GroupBalance) error {
	if err := repo.db.DB.WithContext(ctx).Save(balance).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}
	return nil
}

func (repo *BalanceRepository) GetFinalizedSplitsWithParticipants(ctx context.Context, groupId uuid.UUID) ([]Domain.Split, error) {
	var splits []Domain.Split
	if err := repo.db.DB.WithContext(ctx).Preload("Participants").Where("group_id = ? AND is_finalized = ?", groupId, true).Find(&splits).Error; err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}
	return splits, nil
}

func (repo *BalanceRepository) GetSettlementsForSplits(ctx context.Context, splitIDs []uuid.UUID) ([]Domain.Settlement, error) {
	if len(splitIDs) == 0 {
		return []Domain.Settlement{}, nil
	}
	var settlements []Domain.Settlement
	if err := repo.db.DB.WithContext(ctx).Where("split_id IN ?", splitIDs).Find(&settlements).Error; err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}
	return settlements, nil
}

func (repo *BalanceRepository) GetSettledParticipants(ctx context.Context, splitId uuid.UUID, userId uuid.UUID) (bool, error) {
	var participant Domain.SplitParticipant
	err := repo.db.DB.WithContext(ctx).Where("split_id = ? AND user_id = ?", splitId, userId).First(&participant).Error
	if err != nil {
		return false, nil
	}
	return participant.IsSettled, nil
}

func (repo *BalanceRepository) ReplaceGroupBalances(ctx context.Context, groupId uuid.UUID, balances []Domain.GroupBalance) ([]Domain.GroupBalance, error) {
	tx := repo.db.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Delete(&Domain.GroupBalance{}, "group_id = ?", groupId).Error; err != nil {
		tx.Rollback()
		return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	var result []Domain.GroupBalance
	for _, balance := range balances {
		if err := tx.Create(&balance).Error; err != nil {
			tx.Rollback()
			return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
		}
		tx.Preload("User").First(&balance, "id = ?", balance.Id)
		result = append(result, balance)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	return result, nil
}

func (repo *BalanceRepository) GetSimplifiedDebts(ctx context.Context, groupId uuid.UUID) ([]Domain.SimplifiedDebt, error) {
	var balances []Domain.GroupBalance
	if err := repo.db.DB.WithContext(ctx).Preload("User").Where("group_id = ?", groupId).Find(&balances).Error; err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	if len(balances) == 0 {
		return []Domain.SimplifiedDebt{}, nil
	}

	currencyBalances := make(map[Domain.Currency][]struct {
		UserID   uuid.UUID
		UserName string
		Amount   int64
	})

	for _, b := range balances {
		currencyBalances[b.Currency] = append(currencyBalances[b.Currency], struct {
			UserID   uuid.UUID
			UserName string
			Amount   int64
		}{
			UserID:   b.UserID,
			UserName: b.User.Name,
			Amount:   b.NetAmount,
		})
	}

	var simplifiedDebts []Domain.SimplifiedDebt

	for currency, userBalances := range currencyBalances {
		type userAmount struct {
			UserID   uuid.UUID
			UserName string
			Amount   int64
		}

		var creditors []userAmount
		var debtors []userAmount

		for _, ub := range userBalances {
			if ub.Amount > 0 {
				creditors = append(creditors, userAmount{ub.UserID, ub.UserName, ub.Amount})
			} else if ub.Amount < 0 {
				debtors = append(debtors, userAmount{ub.UserID, ub.UserName, -ub.Amount})
			}
		}

		i, j := 0, 0
		for i < len(debtors) && j < len(creditors) {
			debtor := &debtors[i]
			creditor := &creditors[j]

			settleAmount := debtor.Amount
			if creditor.Amount < settleAmount {
				settleAmount = creditor.Amount
			}

			if settleAmount > 0 {
				simplifiedDebts = append(simplifiedDebts, Domain.SimplifiedDebt{
					FromUserID:   debtor.UserID,
					FromUserName: debtor.UserName,
					ToUserID:     creditor.UserID,
					ToUserName:   creditor.UserName,
					Amount:       settleAmount,
					Currency:     currency,
				})
			}

			debtor.Amount -= settleAmount
			creditor.Amount -= settleAmount

			if debtor.Amount == 0 {
				i++
			}
			if creditor.Amount == 0 {
				j++
			}
		}
	}

	return simplifiedDebts, nil
}
