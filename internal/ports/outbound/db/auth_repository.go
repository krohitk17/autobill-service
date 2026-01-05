package RepositoryPorts

import (
	"context"
	"time"

	Domain "autobill-service/internal/domain"

	"github.com/google/uuid"
)

type AuthRepositoryPort interface {
	FindUser(ctx context.Context, email, password string) (string, error)

	CreateUser(ctx context.Context, email, name, password string) (*Domain.User, error)
	UpdatePassword(ctx context.Context, userId uuid.UUID, oldPassword, newPassword string) error

	DeactivateUser(ctx context.Context, userId uuid.UUID, password string) error
	ReactivateUser(ctx context.Context, email, password string) error

	CreateRefreshToken(ctx context.Context, userId uuid.UUID, token string, expiresAt time.Time) (*Domain.RefreshToken, error)
	GetRefreshToken(ctx context.Context, token string) (*Domain.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, token string) error
	RevokeAllUserRefreshTokens(ctx context.Context, userId uuid.UUID) error
}
