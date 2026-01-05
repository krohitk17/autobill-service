package UserApplication

import (
	"github.com/gofiber/fiber/v2"
	"context"

	Dtos "autobill-service/internal/application/user/dtos"
	Domain "autobill-service/internal/domain"
	HttpPorts "autobill-service/internal/ports/inbound/http"
	RepositoryPorts "autobill-service/internal/ports/outbound/db"
	Errors "autobill-service/pkg/errors"

	"github.com/google/uuid"
)

type UserService struct {
	db RepositoryPorts.UserRepositoryPort
}

func CreateUserService(db RepositoryPorts.UserRepositoryPort) HttpPorts.UserUseCase {
	return &UserService{db: db}
}

func (service *UserService) userToDto(user *Domain.User) *Dtos.UserResult {
	return &Dtos.UserResult{
		ID:        user.Id.String(),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func (service *UserService) FindUserById(ctx context.Context, id uuid.UUID) (*Dtos.UserResult, error) {
	user, err := service.db.FindUserById(ctx, id)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, Errors.ErrUserNotFound)
	}
	return service.userToDto(user), nil
}

func (service *UserService) FindUserByEmail(ctx context.Context, email string) (*Dtos.UserResult, error) {
	user, err := service.db.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, Errors.ErrUserNotFound)
	}
	return service.userToDto(user), nil
}

func (service *UserService) UpdateUser(ctx context.Context, id uuid.UUID, input Dtos.UpdateUserInput) (*Dtos.UserResult, error) {
	updateData := RepositoryPorts.UpdateUserData{
		Name:  input.Name,
		Email: input.Email,
	}
	user, err := service.db.UpdateUser(ctx, id, updateData)
	if err != nil {
		return nil, err
	}
	return service.userToDto(user), nil
}
