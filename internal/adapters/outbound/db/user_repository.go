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

type UserRepository struct {
	db DB.PostgresDB
}

func CreateUserRepository(db DB.PostgresDB) RepositoryPorts.UserRepositoryPort {
	return &UserRepository{db: db}
}

func (repo *UserRepository) FindUserById(ctx context.Context, id uuid.UUID) (*Domain.User, error) {
	var user Domain.User
	if err := repo.db.DB.WithContext(ctx).Where("id = ? AND status = ?", id, Domain.AccountActive).First(&user).Error; err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, Errors.ErrUserNotFound)
	}
	return &user, nil
}

func (repo *UserRepository) FindUserByEmail(ctx context.Context, email string) (*Domain.User, error) {
	var user Domain.User
	if err := repo.db.DB.WithContext(ctx).Where("email = ? AND status = ?", email, Domain.AccountActive).First(&user).Error; err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, Errors.ErrUserNotFound)
	}
	return &user, nil
}

func (repo *UserRepository) UpdateUser(ctx context.Context, id uuid.UUID, updatedData RepositoryPorts.UpdateUserData) (*Domain.User, error) {
	var user Domain.User
	if repo.db.DB.WithContext(ctx).Where("id = ? AND status = ?", id, Domain.AccountActive).First(&user).Error != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, Errors.ErrUserNotFound)
	}

	result := repo.db.DB.WithContext(ctx).Model(&user).Updates(updatedData)
	if result.Error != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrDatabaseFailure)
	}
	return &user, nil
}
