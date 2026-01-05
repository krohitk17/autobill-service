package HttpPorts

import (
	Dtos "autobill-service/internal/application/user/dtos"
	"context"

	"github.com/google/uuid"
)

type UserUseCase interface {
	FindUserById(ctx context.Context, id uuid.UUID) (*Dtos.UserResult, error)
	FindUserByEmail(ctx context.Context, email string) (*Dtos.UserResult, error)

	UpdateUser(ctx context.Context, id uuid.UUID, input Dtos.UpdateUserInput) (*Dtos.UserResult, error)
}
