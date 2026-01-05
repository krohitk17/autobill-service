package RepositoryAdapters

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	Domain "autobill-service/internal/domain"
	DB "autobill-service/internal/infrastructure/db"
	RepositoryPorts "autobill-service/internal/ports/outbound/db"
	Errors "autobill-service/pkg/errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthRepository struct {
	db DB.PostgresDB
}

func CreateAuthRepository(db DB.PostgresDB) RepositoryPorts.AuthRepositoryPort {
	return &AuthRepository{db: db}
}

func (repo *AuthRepository) CreateUser(ctx context.Context, email, name, password string) (*Domain.User, error) {
	hash, hashErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if hashErr != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrPasswordHashFailed)
	}

	tx := repo.db.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	user := Domain.User{
		Email:  email,
		Name:   name,
		Status: Domain.AccountActive,
	}
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, fiber.NewError(fiber.StatusConflict, Errors.ErrEmailAlreadyExists)
		}
		return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	credential := Domain.Credential{
		UserID:       user.Id,
		PasswordHash: string(hash),
	}
	if err := tx.Create(&credential).Error; err != nil {
		tx.Rollback()
		return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	return &user, nil
}

func (repo *AuthRepository) FindUser(ctx context.Context, email string, password string) (string, error) {
	var user Domain.User
	if err := repo.db.DB.WithContext(ctx).First(&user, "email = ? AND status = ?", email, Domain.AccountActive).Error; err != nil {
		return "", fiber.NewError(fiber.StatusNotFound, Errors.ErrUserNotFound)
	}

	var cred Domain.Credential
	if err := repo.db.DB.WithContext(ctx).Where("user_id = ?", user.Id).First(&cred).Error; err != nil {
		return "", fiber.NewError(fiber.StatusNotFound, Errors.ErrUserNotFound)
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(cred.PasswordHash),
		[]byte(password),
	); err != nil {
		return "", fiber.NewError(fiber.StatusUnauthorized, Errors.ErrUnauthorized)
	}
	return user.Id.String(), nil
}

func (repo *AuthRepository) UpdatePassword(ctx context.Context, userId uuid.UUID, oldPassword string, newPassword string) error {
	var cred Domain.Credential
	if err := repo.db.DB.WithContext(ctx).Where("user_id = ?", userId).First(&cred).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, Errors.ErrUserNotFound)
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(cred.PasswordHash),
		[]byte(oldPassword),
	); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, Errors.ErrUnauthorized)
	}

	hash, hashErr := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if hashErr != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrPasswordHashFailed)
	}

	cred.PasswordHash = string(hash)
	if err := repo.db.DB.WithContext(ctx).Save(&cred).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}
	return nil
}

func (repo *AuthRepository) DeactivateUser(ctx context.Context, userId uuid.UUID, password string) error {
	var user Domain.User
	if err := repo.db.DB.WithContext(ctx).Preload("Credential").Where("id = ? AND status = ?", userId, Domain.AccountActive).First(&user).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, Errors.ErrUserNotFound)
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.Credential.PasswordHash),
		[]byte(password),
	); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, Errors.ErrUnauthorized)
	}

	if err := repo.db.DB.WithContext(ctx).
		Model(&Domain.User{}).
		Where("id = ?", userId).
		Update("status", Domain.AccountDeactivated).
		Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}
	return nil
}

func (repo *AuthRepository) ReactivateUser(ctx context.Context, email string, password string) error {
	var user Domain.User
	if err := repo.db.DB.WithContext(ctx).Preload("Credential").Where("email = ? AND status = ?", email, Domain.AccountDeactivated).First(&user).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, Errors.ErrUserNotFound)
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.Credential.PasswordHash),
		[]byte(password),
	); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, Errors.ErrUnauthorized)
	}

	if err := repo.db.DB.WithContext(ctx).
		Model(&Domain.User{}).
		Where("email = ?", email).
		Update("status", Domain.AccountActive).
		Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}
	return nil
}

func (repo *AuthRepository) CreateRefreshToken(ctx context.Context, userId uuid.UUID, token string, expiresAt time.Time) (*Domain.RefreshToken, error) {
	refreshToken := &Domain.RefreshToken{
		UserID:    userId,
		Token:     token,
		ExpiresAt: expiresAt,
		Revoked:   false,
	}

	if err := repo.db.DB.WithContext(ctx).Create(refreshToken).Error; err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}

	return refreshToken, nil
}

func (repo *AuthRepository) GetRefreshToken(ctx context.Context, token string) (*Domain.RefreshToken, error) {
	var refreshToken Domain.RefreshToken
	if err := repo.db.DB.WithContext(ctx).Preload("User").Where("token = ?", token).First(&refreshToken).Error; err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, Errors.ErrInvalidRefreshToken)
	}
	return &refreshToken, nil
}

func (repo *AuthRepository) RevokeRefreshToken(ctx context.Context, token string) error {
	result := repo.db.DB.WithContext(ctx).
		Model(&Domain.RefreshToken{}).
		Where("token = ?", token).
		Update("revoked", true)

	if result.Error != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}
	if result.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusNotFound, Errors.ErrInvalidRefreshToken)
	}
	return nil
}

func (repo *AuthRepository) RevokeAllUserRefreshTokens(ctx context.Context, userId uuid.UUID) error {
	if err := repo.db.DB.WithContext(ctx).
		Model(&Domain.RefreshToken{}).
		Where("user_id = ? AND revoked = ?", userId, false).
		Update("revoked", true).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}
	return nil
}
