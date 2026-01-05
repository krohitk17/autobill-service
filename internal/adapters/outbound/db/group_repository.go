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

type GroupRepository struct {
	db DB.PostgresDB
}

func CreateGroupRepository(db DB.PostgresDB) RepositoryPorts.GroupRepositoryPort {
	return &GroupRepository{db: db}
}

func (repo *GroupRepository) CreateGroup(ctx context.Context, name string, ownerId uuid.UUID, simplifyDebts bool) (*Domain.Group, error) {
	tx := repo.db.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	group := Domain.Group{
		Name:          name,
		SimplifyDebts: simplifyDebts,
	}
	if err := tx.Create(&group).Error; err != nil {
		tx.Rollback()
		return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	membership := Domain.GroupMembership{
		GroupID: group.Id,
		UserID:  ownerId,
		Role:    Domain.GroupRoleOwner,
	}
	if err := tx.Create(&membership).Error; err != nil {
		tx.Rollback()
		return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	return &group, nil
}

func (repo *GroupRepository) GetGroupsByUserId(ctx context.Context, userId uuid.UUID, limit, offset int) ([]Domain.Group, int64, error) {
	var groups []Domain.Group
	var total int64

	baseQuery := repo.db.DB.WithContext(ctx).Model(&Domain.Group{}).
		Joins("JOIN group_memberships ON group_memberships.group_id = groups.id").
		Where("group_memberships.user_id = ?", userId)

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	if total == 0 {
		return []Domain.Group{}, 0, nil
	}

	if err := baseQuery.Limit(limit).Offset(offset).Find(&groups).Error; err != nil {
		return nil, 0, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	return groups, total, nil
}

func (repo *GroupRepository) GetGroupById(ctx context.Context, groupId uuid.UUID) (*Domain.Group, error) {
	var group Domain.Group
	if err := repo.db.DB.WithContext(ctx).First(&group, "id = ?", groupId).Error; err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, Errors.ErrGroupNotFound)
	}
	return &group, nil
}

func (repo *GroupRepository) UpdateGroup(ctx context.Context, groupId uuid.UUID, updates map[string]interface{}) (*Domain.Group, error) {
	var group Domain.Group
	if err := repo.db.DB.WithContext(ctx).First(&group, "id = ?", groupId).Error; err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, Errors.ErrGroupNotFound)
	}

	if err := repo.db.DB.WithContext(ctx).Model(&group).Updates(updates).Error; err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	return &group, nil
}

func (repo *GroupRepository) GetGroupWithMembers(ctx context.Context, groupId uuid.UUID) (*Domain.Group, error) {
	var group Domain.Group
	if err := repo.db.DB.WithContext(ctx).Preload("Memberships.User").First(&group, "id = ?", groupId).Error; err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, Errors.ErrGroupNotFound)
	}
	return &group, nil
}

func (repo *GroupRepository) DeleteGroup(ctx context.Context, groupId uuid.UUID) error {
	var activeSplitCount int64
	if err := repo.db.DB.WithContext(ctx).Model(&Domain.Split{}).Where("group_id = ? AND is_finalized = ?", groupId, false).Count(&activeSplitCount).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}
	if activeSplitCount > 0 {
		return fiber.NewError(fiber.StatusBadRequest, Errors.ErrGroupHasActiveSplits)
	}

	if err := repo.db.DB.WithContext(ctx).Delete(&Domain.Group{}, "id = ?", groupId).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}
	return nil
}

func (repo *GroupRepository) AddMember(ctx context.Context, groupId, userId uuid.UUID, role Domain.GroupRole) (*Domain.GroupMembership, error) {
	tx := repo.db.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var existing Domain.GroupMembership
	err := tx.Raw("SELECT * FROM group_memberships WHERE group_id = ? AND user_id = ? FOR UPDATE", groupId, userId).Scan(&existing).Error
	if err == nil && existing.Id != (uuid.UUID{}) {
		tx.Rollback()
		return nil, fiber.NewError(fiber.StatusConflict, Errors.ErrUserAlreadyMember)
	}

	membership := Domain.GroupMembership{
		GroupID: groupId,
		UserID:  userId,
		Role:    role,
	}
	if err := tx.Create(&membership).Error; err != nil {
		tx.Rollback()
		return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	if err := tx.Preload("User").First(&membership, "id = ?", membership.Id).Error; err != nil {
		tx.Rollback()
		return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	return &membership, nil
}

func (repo *GroupRepository) GetMembership(ctx context.Context, groupId, userId uuid.UUID) (*Domain.GroupMembership, error) {
	var membership Domain.GroupMembership
	if err := repo.db.DB.WithContext(ctx).Preload("User").Where("group_id = ? AND user_id = ?", groupId, userId).First(&membership).Error; err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, Errors.ErrNotGroupMember)
	}
	return &membership, nil
}

func (repo *GroupRepository) UpdateMemberRole(ctx context.Context, groupId, userId uuid.UUID, role Domain.GroupRole) error {
	result := repo.db.DB.WithContext(ctx).Model(&Domain.GroupMembership{}).
		Where("group_id = ? AND user_id = ?", groupId, userId).
		Update("role", role)
	if result.Error != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}
	if result.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusNotFound, Errors.ErrNotGroupMember)
	}
	return nil
}

func (repo *GroupRepository) RemoveMember(ctx context.Context, groupId, userId uuid.UUID) error {
	result := repo.db.DB.WithContext(ctx).Delete(&Domain.GroupMembership{}, "group_id = ? AND user_id = ?", groupId, userId)
	if result.Error != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}
	if result.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusNotFound, Errors.ErrNotGroupMember)
	}
	return nil
}

func (repo *GroupRepository) IsGroupAdmin(ctx context.Context, groupId, userId uuid.UUID) (bool, error) {
	var membership Domain.GroupMembership
	if err := repo.db.DB.WithContext(ctx).Where("group_id = ? AND user_id = ?", groupId, userId).First(&membership).Error; err != nil {
		return false, fiber.NewError(fiber.StatusNotFound, Errors.ErrNotGroupMember)
	}
	return membership.Role == Domain.GroupRoleAdmin || membership.Role == Domain.GroupRoleOwner, nil
}

func (repo *GroupRepository) IsGroupOwner(ctx context.Context, groupId, userId uuid.UUID) (bool, error) {
	var membership Domain.GroupMembership
	if err := repo.db.DB.WithContext(ctx).Where("group_id = ? AND user_id = ?", groupId, userId).First(&membership).Error; err != nil {
		return false, fiber.NewError(fiber.StatusNotFound, Errors.ErrNotGroupMember)
	}
	return membership.Role == Domain.GroupRoleOwner, nil
}
