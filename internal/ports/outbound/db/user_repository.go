package RepositoryPorts

import (
	"context"

	Domain "autobill-service/internal/domain"

	"github.com/google/uuid"
)

type UpdateUserData struct {
	Email string
	Name  string
}

type UserRepositoryPort interface {
	FindUserById(ctx context.Context, id uuid.UUID) (*Domain.User, error)
	FindUserByEmail(ctx context.Context, email string) (*Domain.User, error)

	UpdateUser(ctx context.Context, id uuid.UUID, updatedUser UpdateUserData) (*Domain.User, error)
}
