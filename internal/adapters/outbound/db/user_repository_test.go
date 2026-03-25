package RepositoryAdapters

import (
	"context"
	"testing"

	Domain "autobill-service/internal/domain"
	DB "autobill-service/internal/infrastructure/db"
	RepositoryPorts "autobill-service/internal/ports/outbound/db"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestUserWithStatus(t *testing.T, postgres *DB.PostgresDB, status Domain.AccountStatus) *Domain.User {
	t.Helper()
	user := createTestUser(t, postgres)

	if status != Domain.AccountActive {
		err := postgres.DB.Model(&Domain.User{}).Where("id = ?", user.Id).Update("status", status).Error
		require.NoError(t, err)

		err = postgres.DB.Where("id = ?", user.Id).First(user).Error
		require.NoError(t, err)
	}

	return user
}

func TestUserRepository_FindUserById(t *testing.T) {
	postgres := OpenTestDB(t)
	defer DBCleanup(t, postgres)

	repo := CreateUserRepository(*postgres)
	ctx := context.Background()

	activeUser := createTestUserWithStatus(t, postgres, Domain.AccountActive)
	deactivatedUser := createTestUserWithStatus(t, postgres, Domain.AccountDeactivated)

	tests := []struct {
		name        string
		userID      uuid.UUID
		expectError bool
	}{
		{
			name:        "FindUserById_Success",
			userID:      activeUser.Id,
			expectError: false,
		},
		{
			name:        "FindUserById_NotFound",
			userID:      uuid.New(),
			expectError: true,
		},
		{
			name:        "FindUserById_DeactivatedUser",
			userID:      deactivatedUser.Id,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.FindUserById(ctx, tt.userID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, user)
				assert.Equal(t, activeUser.Id, user.Id)
				assert.Equal(t, activeUser.Email, user.Email)
				assert.Equal(t, activeUser.Name, user.Name)
				assert.Equal(t, Domain.AccountActive, user.Status)
			}
		})
	}
}

func TestUserRepository_FindUserByEmail(t *testing.T) {
	postgres := OpenTestDB(t)
	defer DBCleanup(t, postgres)

	repo := CreateUserRepository(*postgres)
	ctx := context.Background()

	activeUser := createTestUserWithStatus(t, postgres, Domain.AccountActive)
	deactivatedUser := createTestUserWithStatus(t, postgres, Domain.AccountDeactivated)

	tests := []struct {
		name        string
		email       string
		expectError bool
	}{
		{
			name:        "FindUserByEmail_Success",
			email:       activeUser.Email,
			expectError: false,
		},
		{
			name:        "FindUserByEmail_NotFound",
			email:       uuid.NewString() + "@example.com",
			expectError: true,
		},
		{
			name:        "FindUserByEmail_DeactivatedUser",
			email:       deactivatedUser.Email,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.FindUserByEmail(ctx, tt.email)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, user)
				assert.Equal(t, activeUser.Id, user.Id)
				assert.Equal(t, activeUser.Email, user.Email)
				assert.Equal(t, activeUser.Name, user.Name)
				assert.Equal(t, Domain.AccountActive, user.Status)
			}
		})
	}
}

func TestUserRepository_UpdateUser(t *testing.T) {
	postgres := OpenTestDB(t)
	defer DBCleanup(t, postgres)

	repo := CreateUserRepository(*postgres)
	ctx := context.Background()

	activeUser := createTestUserWithStatus(t, postgres, Domain.AccountActive)
	deactivatedUser := createTestUserWithStatus(t, postgres, Domain.AccountDeactivated)

	tests := []struct {
		name        string
		userID      uuid.UUID
		updatedData RepositoryPorts.UpdateUserData
		expectError bool
	}{
		{
			name:   "UpdateUser_Success",
			userID: activeUser.Id,
			updatedData: RepositoryPorts.UpdateUserData{
				Email: uuid.NewString() + "@example.com",
				Name:  "Updated Name",
			},
			expectError: false,
		},
		{
			name:   "UpdateUser_NotFound",
			userID: uuid.New(),
			updatedData: RepositoryPorts.UpdateUserData{
				Email: uuid.NewString() + "@example.com",
				Name:  "Ghost User",
			},
			expectError: true,
		},
		{
			name:   "UpdateUser_DeactivatedUser",
			userID: deactivatedUser.Id,
			updatedData: RepositoryPorts.UpdateUserData{
				Email: uuid.NewString() + "@example.com",
				Name:  "Should Not Update",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatedUser, err := repo.UpdateUser(ctx, tt.userID, tt.updatedData)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, updatedUser)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, updatedUser)
				assert.Equal(t, tt.userID, updatedUser.Id)

				var persistedUser Domain.User
				err := postgres.DB.Where("id = ?", tt.userID).First(&persistedUser).Error
				require.NoError(t, err)
				assert.Equal(t, tt.updatedData.Email, persistedUser.Email)
				assert.Equal(t, tt.updatedData.Name, persistedUser.Name)
			}
		})
	}
}
