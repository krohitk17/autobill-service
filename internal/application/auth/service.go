package auth

import (
	"context"

	"github.com/gofiber/fiber/v2"

	Dtos "autobill-service/internal/application/auth/dtos"
	RepositoryPorts "autobill-service/internal/ports/outbound/db"
	Errors "autobill-service/pkg/errors"
	JWTUtil "autobill-service/pkg/jwt"
	Logger "autobill-service/pkg/logger"

	"github.com/google/uuid"
)

type AuthService struct {
	db   RepositoryPorts.AuthRepositoryPort
	util JWTUtil.JWTUtil
}

func CreateAuthService(db RepositoryPorts.AuthRepositoryPort, util JWTUtil.JWTUtil) *AuthService {
	return &AuthService{db: db, util: util}
}

func (service *AuthService) generateTokenPair(ctx context.Context, userId uuid.UUID) (*Dtos.AuthResult, error) {
	jwtToken, jwtErr := service.util.Generate(userId.String())
	if jwtErr != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrJWTGenerationFailed)
	}

	refreshToken, refreshErr := service.util.GenerateRefreshToken()
	if refreshErr != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, Errors.ErrJWTGenerationFailed)
	}

	expiresAt := service.util.GetRefreshTokenExpiry()
	_, dbErr := service.db.CreateRefreshToken(ctx, userId, refreshToken, expiresAt)
	if dbErr != nil {
		return nil, dbErr
	}

	return &Dtos.AuthResult{
		ID:           userId.String(),
		Token:        jwtToken,
		RefreshToken: refreshToken,
	}, nil
}

func (service *AuthService) RegisterUser(ctx context.Context, input Dtos.RegisterUserInput) (*Dtos.AuthResult, error) {
	if len(input.Password) > 72 {
		return nil, fiber.NewError(fiber.StatusBadRequest, Errors.ErrPasswordTooLong)
	}

	user, dbErr := service.db.CreateUser(ctx, input.Email, input.Name, input.Password)
	if dbErr != nil {
		return nil, dbErr
	}

	result, err := service.generateTokenPair(ctx, user.Id)
	if err != nil {
		return nil, err
	}

	Logger.Debug().
		Str("operation", "RegisterUser").
		Str("userId", user.Id.String()).
		Str("email", input.Email).
		Msg("User registered successfully")

	return result, nil
}

func (service *AuthService) AuthenticateUser(ctx context.Context, input Dtos.LoginInput) (*Dtos.AuthResult, error) {
	id, dbErr := service.db.FindUser(ctx, input.Email, input.Password)
	if dbErr != nil {
		return nil, fiber.NewError(fiber.StatusUnauthorized, Errors.ErrUnauthorized)
	}

	userId, _ := uuid.Parse(id)
	result, err := service.generateTokenPair(ctx, userId)
	if err != nil {
		return nil, err
	}

	Logger.Debug().
		Str("operation", "AuthenticateUser").
		Str("userId", id).
		Str("email", input.Email).
		Msg("User authenticated successfully")

	return result, nil
}

func (service *AuthService) RefreshToken(ctx context.Context, input Dtos.RefreshTokenInput) (*Dtos.AuthResult, error) {
	storedToken, err := service.db.GetRefreshToken(ctx, input.RefreshToken)
	if err != nil {
		return nil, err
	}

	if !storedToken.IsValid() {
		return nil, fiber.NewError(fiber.StatusUnauthorized, Errors.ErrRefreshTokenRevoked)
	}

	if storedToken.User.Status != "active" {
		return nil, fiber.NewError(fiber.StatusUnauthorized, Errors.ErrUnauthorized)
	}

	_ = service.db.RevokeRefreshToken(ctx, input.RefreshToken)

	result, tokenErr := service.generateTokenPair(ctx, storedToken.UserID)
	if tokenErr != nil {
		return nil, tokenErr
	}

	Logger.Debug().
		Str("operation", "RefreshToken").
		Str("userId", storedToken.UserID.String()).
		Msg("Token refreshed successfully")

	return result, nil
}

func (service *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return service.db.RevokeRefreshToken(ctx, refreshToken)
}

func (service *AuthService) LogoutAll(ctx context.Context, userId uuid.UUID) error {
	return service.db.RevokeAllUserRefreshTokens(ctx, userId)
}

func (service *AuthService) UpdatePassword(ctx context.Context, id uuid.UUID, currentPassword, newPassword string) error {
	if len(newPassword) > 72 {
		return fiber.NewError(fiber.StatusBadRequest, Errors.ErrPasswordTooLong)
	}

	return service.db.UpdatePassword(ctx, id, currentPassword, newPassword)
}

func (service *AuthService) DeactivateUser(ctx context.Context, id uuid.UUID, password string) error {
	return service.db.DeactivateUser(ctx, id, password)
}

func (service *AuthService) ReactivateUser(ctx context.Context, email, password string) error {
	return service.db.ReactivateUser(ctx, email, password)
}
