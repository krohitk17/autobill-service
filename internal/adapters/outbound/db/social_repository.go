package RepositoryAdapters

import (
	"github.com/gofiber/fiber/v2"
	"context"

	Domain "autobill-service/internal/domain"
	DB "autobill-service/internal/infrastructure/db"
	RepositoryPorts "autobill-service/internal/ports/outbound/db"
	Errors "autobill-service/pkg/errors"

	"github.com/google/uuid"
)

type SocialRepository struct {
	db DB.PostgresDB
}

func CreateSocialRepository(db DB.PostgresDB) RepositoryPorts.SocialRepositoryPort {
	return &SocialRepository{db: db}
}

func (repo *SocialRepository) GetFriendRequestsList(ctx context.Context, userId uuid.UUID, requestType RepositoryPorts.FriendRequestType, limit, offset int) ([]*Domain.FriendRequest, int64, error) {
	var requests []*Domain.FriendRequest
	var total int64
	var preloadField, whereField string
	var whereArgs []any
	switch requestType {
	case RepositoryPorts.FriendRequestReceived:
		preloadField = "Sender"
		whereField = "receiver_id = ? AND status != ?"
		whereArgs = []any{userId, Domain.FriendRejected}
	case RepositoryPorts.FriendRequestSent:
		preloadField = "Receiver"
		whereField = "sender_id = ?"
		whereArgs = []any{userId}
	}

	if err := repo.db.DB.WithContext(ctx).Model(&Domain.FriendRequest{}).Where(whereField, whereArgs...).Count(&total).Error; err != nil {
		return nil, 0, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	if err := repo.db.DB.WithContext(ctx).Preload(preloadField).Where(whereField, whereArgs...).Limit(limit).Offset(offset).Find(&requests).Error; err != nil {
		return nil, 0, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}
	return requests, total, nil
}

func (repo *SocialRepository) CreateFriendRequest(ctx context.Context, senderId uuid.UUID, receiverId uuid.UUID) (*Domain.FriendRequest, error) {
	request := &Domain.FriendRequest{
		SenderId:   senderId,
		ReceiverId: receiverId,
		Status:     Domain.FriendPending,
	}
	if err := repo.db.DB.WithContext(ctx).Create(request).Error; err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}
	return request, nil
}

func (repo *SocialRepository) AcceptFriendRequest(ctx context.Context, receiverId, requestId uuid.UUID) error {
	var request Domain.FriendRequest
	if err := repo.db.DB.WithContext(ctx).Where("id = ? AND receiver_id = ? AND status = ?", requestId, receiverId, Domain.FriendPending).First(&request).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, Errors.ErrFriendRequestNotFound)
	}

	tx := repo.db.DB.WithContext(ctx).Begin()
	var err error
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	request.Status = Domain.FriendAccepted
	if err = tx.Save(&request).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	var sentFriend Domain.Friendship
	sentFriend.UserID = request.SenderId
	sentFriend.FriendID = request.ReceiverId
	if err = tx.Create(&sentFriend).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	var receivedFriend Domain.Friendship
	receivedFriend.UserID = request.ReceiverId
	receivedFriend.FriendID = request.SenderId
	if err = tx.Create(&receivedFriend).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	return nil
}

func (repo *SocialRepository) RejectFriendRequest(ctx context.Context, receiverId, requestId uuid.UUID) error {
	var request Domain.FriendRequest
	if err := repo.db.DB.WithContext(ctx).Where("id = ? AND receiver_id = ? AND status = ?", requestId, receiverId, Domain.FriendPending).First(&request).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, Errors.ErrFriendRequestNotFound)
	}

	request.Status = Domain.FriendRejected
	if err := repo.db.DB.WithContext(ctx).Save(&request).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	return nil
}

func (repo *SocialRepository) GetFriendsList(ctx context.Context, userId uuid.UUID, limit, offset int) ([]*Domain.User, int64, error) {
	var friends []*Domain.User
	var total int64

	baseQuery := repo.db.DB.WithContext(ctx).Model(&Domain.User{}).
		Joins("JOIN friendships ON friendships.friend_id = users.id").
		Where("friendships.user_id = ?", userId)

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	if total == 0 {
		return []*Domain.User{}, 0, nil
	}

	if err := baseQuery.Limit(limit).Offset(offset).Find(&friends).Error; err != nil {
		return nil, 0, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	return friends, total, nil
}

func (repo *SocialRepository) RemoveFriend(ctx context.Context, userId, friendId uuid.UUID) error {
	tx := repo.db.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	result1 := tx.Where("user_id = ? AND friend_id = ?", userId, friendId).Delete(&Domain.Friendship{})
	if result1.Error != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	result2 := tx.Where("user_id = ? AND friend_id = ?", friendId, userId).Delete(&Domain.Friendship{})
	if result2.Error != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	if result1.RowsAffected == 0 && result2.RowsAffected == 0 {
		tx.Rollback()
		return fiber.NewError(fiber.StatusNotFound, Errors.ErrFriendshipNotFound)
	}

	if err := tx.Commit().Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	return nil
}

func (repo *SocialRepository) CheckExistingRequest(ctx context.Context, senderId, receiverId uuid.UUID) (bool, error) {
	var count int64
	err := repo.db.DB.WithContext(ctx).Model(&Domain.FriendRequest{}).
		Where("((sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)) AND status = ?",
			senderId, receiverId, receiverId, senderId, Domain.FriendPending).
		Count(&count).Error
	if err != nil {
		return false, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}
	return count > 0, nil
}

func (repo *SocialRepository) CheckFriendship(ctx context.Context, userId, friendId uuid.UUID) (bool, error) {
	var count int64
	err := repo.db.DB.WithContext(ctx).Model(&Domain.Friendship{}).
		Where("user_id = ? AND friend_id = ?", userId, friendId).
		Count(&count).Error
	if err != nil {
		return false, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}
	return count > 0, nil
}
