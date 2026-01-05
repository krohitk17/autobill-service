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

func (repo *SplitRepository) DeleteSplit(ctx context.Context, splitId uuid.UUID) error {
	tx := repo.db.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Delete(&Domain.SplitParticipant{}, "split_id = ?", splitId).Error; err != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	if err := tx.Delete(&Domain.ReversalSplit{}, "original_split_id = ? OR reversal_split_id = ?", splitId, splitId).Error; err != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	result := tx.Delete(&Domain.Split{}, "id = ?", splitId)
	if result.Error != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}
	if result.RowsAffected == 0 {
		tx.Rollback()
		return fiber.NewError(fiber.StatusNotFound, Errors.ErrSplitNotFound)
	}

	if err := tx.Commit().Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}
	return nil
}

func (repo *SplitRepository) IsSplitReversed(ctx context.Context, splitId uuid.UUID) (bool, error) {
	var count int64
	if err := repo.db.DB.WithContext(ctx).Model(&Domain.ReversalSplit{}).Where("original_split_id = ?", splitId).Count(&count).Error; err != nil {
		return false, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}
	return count > 0, nil
}

func (repo *SplitRepository) AddParticipant(ctx context.Context, participant *Domain.SplitParticipant) (*Domain.SplitParticipant, error) {
	tx := repo.db.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var existing Domain.SplitParticipant
	err := tx.Raw("SELECT * FROM split_participants WHERE split_id = ? AND user_id = ? FOR UPDATE", participant.SplitID, participant.UserID).Scan(&existing).Error
	if err == nil && existing.Id != (uuid.UUID{}) {
		tx.Rollback()
		return nil, fiber.NewError(fiber.StatusConflict, Errors.ErrParticipantExists)
	}

	if err := tx.Create(participant).Error; err != nil {
		tx.Rollback()
		return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	if err := tx.Preload("User").First(participant, "id = ?", participant.Id).Error; err != nil {
		tx.Rollback()
		return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	return participant, nil
}

func (repo *SplitRepository) GetParticipant(ctx context.Context, splitId, userId uuid.UUID) (*Domain.SplitParticipant, error) {
	var participant Domain.SplitParticipant
	if err := repo.db.DB.WithContext(ctx).Preload("User").Where("split_id = ? AND user_id = ?", splitId, userId).First(&participant).Error; err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, Errors.ErrParticipantNotFound)
	}
	return &participant, nil
}

func (repo *SplitRepository) UpdateParticipant(ctx context.Context, participant *Domain.SplitParticipant) error {
	if err := repo.db.DB.WithContext(ctx).Save(participant).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}
	return nil
}

func (repo *SplitRepository) FinalizeSplit(ctx context.Context, splitId uuid.UUID) error {
	result := repo.db.DB.WithContext(ctx).Model(&Domain.Split{}).
		Where("id = ? AND is_finalized = ?", splitId, false).
		Update("is_finalized", true)

	if result.Error != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}
	if result.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusBadRequest, Errors.ErrSplitAlreadyFinalized)
	}
	return nil
}

func (repo *SplitRepository) CreateReversalSplitWithParticipants(ctx context.Context, originalSplitId uuid.UUID, reversalSplit *Domain.Split, participants []Domain.SplitParticipant) (*Domain.Split, []Domain.SplitParticipant, error) {
	tx := repo.db.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(reversalSplit).Error; err != nil {
		tx.Rollback()
		return nil, nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	reversalRecord := Domain.ReversalSplit{
		OriginalSplitID: originalSplitId,
		ReversalSplitID: reversalSplit.Id,
		Reason:          "Split reversal",
	}
	if err := tx.Create(&reversalRecord).Error; err != nil {
		tx.Rollback()
		return nil, nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	createdParticipants := make([]Domain.SplitParticipant, len(participants))
	for i := range participants {
		participants[i].SplitID = reversalSplit.Id
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

	if err := tx.Commit().Error; err != nil {
		return nil, nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	return reversalSplit, createdParticipants, nil
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
		Where("splits.group_id = ? AND split_participants.user_id = ? AND splits.is_finalized = ?", groupId, userId, false).
		Count(&count).Error

	if err != nil {
		return false, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	return count > 0, nil
}
