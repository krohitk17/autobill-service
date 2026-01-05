package HttpPorts

import (
	Dtos "autobill-service/internal/application/auth/dtos"
	"context"

	"github.com/google/uuid"
)

type AuthUseCase interface {
	RegisterUser(ctx context.Context, input Dtos.RegisterUserInput) (*Dtos.AuthResult, error)
	AuthenticateUser(ctx context.Context, input Dtos.LoginInput) (*Dtos.AuthResult, error)
	RefreshToken(ctx context.Context, input Dtos.RefreshTokenInput) (*Dtos.AuthResult, error)
	UpdatePassword(ctx context.Context, id uuid.UUID, currentPassword, newPassword string) error

	Logout(ctx context.Context, refreshToken string) error
	LogoutAll(ctx context.Context, userId uuid.UUID) error
	DeactivateUser(ctx context.Context, id uuid.UUID, password string) error
	ReactivateUser(ctx context.Context, email, password string) error
}
