package RepositoryAdapters

import (
	"context"
	"testing"

	Domain "autobill-service/internal/domain"
	Config "autobill-service/internal/infrastructure/config"
	DB "autobill-service/internal/infrastructure/db"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestUser(t *testing.T, db *DB.PostgresDB) *Domain.User {
	user := Domain.User{
		Email:  uuid.NewString() + "@example.com",
		Name:   "Test User",
		Status: Domain.AccountActive,
	}
	if err := db.DB.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return &user
}

func createTestGroup(t *testing.T, db *DB.PostgresDB, ownerID uuid.UUID, name string) *Domain.Group {
	group := Domain.Group{
		Name:          name,
		OwnerID:       ownerID,
		SimplifyDebts: false,
	}
	if err := db.DB.Create(&group).Error; err != nil {
		t.Fatalf("Failed to create test group: %v", err)
	}

	membership := Domain.GroupMembership{
		GroupID: group.Id,
		UserID:  ownerID,
		Role:    Domain.GroupRoleOwner,
	}
	if err := db.DB.Create(&membership).Error; err != nil {
		t.Fatalf("Failed to create owner membership: %v", err)
	}

	return &group
}

func OpenTestDB(t *testing.T) *DB.PostgresDB {
	config := Config.LoadTestConfig()
	postgres, err := DB.CreatePostgresDb(config.Database)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	return postgres
}

func DBCleanup(t *testing.T, db *DB.PostgresDB) {
	t.Helper()
	if err := db.DB.Exec("TRUNCATE TABLE group_memberships, groups, users RESTART IDENTITY CASCADE").Error; err != nil {
		t.Fatalf("Failed to clean up database: %v", err)
	}
}

func TestGroupRepository_CreateGroup(t *testing.T) {
	postgres := OpenTestDB(t)
	defer DBCleanup(t, postgres)

	repo := CreateGroupRepository(*postgres)
	user := createTestUser(t, postgres)
	ctx := context.Background()

	tests := []struct {
		name          string
		groupName     string
		ownerID       uuid.UUID
		simplifyDebts bool
		expectError   bool
	}{
		{
			name:          "CreateGroup_Success",
			groupName:     "Test Group",
			ownerID:       user.Id,
			simplifyDebts: false,
			expectError:   false,
		},
		{
			name:          "CreateGroup_WithSimplifyDebts",
			groupName:     "Trip Expenses",
			ownerID:       user.Id,
			simplifyDebts: true,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, err := repo.CreateGroup(ctx, tt.groupName, tt.ownerID, tt.simplifyDebts)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, group)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, group)
				assert.Equal(t, tt.groupName, group.Name)
				assert.Equal(t, tt.ownerID, group.OwnerID)
				assert.Equal(t, tt.simplifyDebts, group.SimplifyDebts)
				assert.NotEqual(t, uuid.UUID{}, group.Id)

				var membership Domain.GroupMembership
				err := postgres.DB.Where("group_id = ? AND user_id = ?", group.Id, tt.ownerID).
					First(&membership).Error
				assert.NoError(t, err)
				assert.Equal(t, Domain.GroupRoleOwner, membership.Role)
			}
		})
	}
}

func TestGroupRepository_GetGroupById(t *testing.T) {
	postgres := OpenTestDB(t)
	defer DBCleanup(t, postgres)

	repo := CreateGroupRepository(*postgres)
	user := createTestUser(t, postgres)
	group := createTestGroup(t, postgres, user.Id, "Test Group")
	ctx := context.Background()

	tests := []struct {
		name        string
		groupID     uuid.UUID
		expectError bool
		expectFound bool
	}{
		{
			name:        "GetGroupById_Success",
			groupID:     group.Id,
			expectError: false,
			expectFound: true,
		},
		{
			name:        "GetGroupById_NotFound",
			groupID:     uuid.New(),
			expectError: true,
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retrievedGroup, err := repo.GetGroupById(ctx, tt.groupID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, retrievedGroup)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, retrievedGroup)
				assert.Equal(t, group.Id, retrievedGroup.Id)
				assert.Equal(t, group.Name, retrievedGroup.Name)
				assert.Equal(t, group.OwnerID, retrievedGroup.OwnerID)
			}
		})
	}
}

func TestGroupRepository_GetGroupsByUserId(t *testing.T) {
	postgres := OpenTestDB(t)
	defer DBCleanup(t, postgres)

	repo := CreateGroupRepository(*postgres)
	user1 := createTestUser(t, postgres)
	user2 := createTestUser(t, postgres)
	ctx := context.Background()

	group1 := createTestGroup(t, postgres, user1.Id, "Group 1")
	group2 := createTestGroup(t, postgres, user1.Id, "Group 2")

	createTestGroup(t, postgres, user2.Id, "Group 3")

	tests := []struct {
		name        string
		userID      uuid.UUID
		limit       int
		offset      int
		expectCount int
		expectTotal int64
		expectError bool
	}{
		{
			name:        "GetGroupsByUserId_Success",
			userID:      user1.Id,
			limit:       10,
			offset:      0,
			expectCount: 2,
			expectTotal: 2,
			expectError: false,
		},
		{
			name:        "GetGroupsByUserId_WithPagination",
			userID:      user1.Id,
			limit:       1,
			offset:      0,
			expectCount: 1,
			expectTotal: 2,
			expectError: false,
		},
		{
			name:        "GetGroupsByUserId_NoGroups",
			userID:      uuid.New(),
			limit:       10,
			offset:      0,
			expectCount: 0,
			expectTotal: 0,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groups, total, err := repo.GetGroupsByUserId(ctx, tt.userID, tt.limit, tt.offset)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectTotal, total)
			assert.Len(t, groups, tt.expectCount)
		})
	}

	groups, _, err := repo.GetGroupsByUserId(ctx, user1.Id, 10, 0)
	assert.NoError(t, err)
	groupIDs := make(map[uuid.UUID]bool)
	for _, g := range groups {
		groupIDs[g.Id] = true
	}
	assert.True(t, groupIDs[group1.Id])
	assert.True(t, groupIDs[group2.Id])
}

func TestGroupRepository_UpdateGroup(t *testing.T) {
	postgres := OpenTestDB(t)
	defer DBCleanup(t, postgres)

	repo := CreateGroupRepository(*postgres)
	user := createTestUser(t, postgres)
	group := createTestGroup(t, postgres, user.Id, "Original Name")
	ctx := context.Background()

	tests := []struct {
		name        string
		groupID     uuid.UUID
		updates     map[string]interface{}
		expectError bool
		verifyFunc  func(*testing.T, *Domain.Group)
	}{
		{
			name:    "UpdateGroup_Name",
			groupID: group.Id,
			updates: map[string]interface{}{
				"name": "Updated Name",
			},
			expectError: false,
			verifyFunc: func(t *testing.T, g *Domain.Group) {
				assert.Equal(t, "Updated Name", g.Name)
			},
		},
		{
			name:    "UpdateGroup_SimplifyDebts",
			groupID: group.Id,
			updates: map[string]interface{}{
				"simplify_debts": true,
			},
			expectError: false,
			verifyFunc: func(t *testing.T, g *Domain.Group) {
				assert.True(t, g.SimplifyDebts)
			},
		},
		{
			name:        "UpdateGroup_NotFound",
			groupID:     uuid.New(),
			updates:     map[string]interface{}{"name": "New Name"},
			expectError: true,
			verifyFunc:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatedGroup, err := repo.UpdateGroup(ctx, tt.groupID, tt.updates)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, updatedGroup)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, updatedGroup)
				if tt.verifyFunc != nil {
					tt.verifyFunc(t, updatedGroup)
				}
			}
		})
	}
}

func TestGroupRepository_GetGroupWithMembers(t *testing.T) {
	postgres := OpenTestDB(t)
	defer DBCleanup(t, postgres)

	repo := CreateGroupRepository(*postgres)
	owner := createTestUser(t, postgres)
	member := createTestUser(t, postgres)
	group := createTestGroup(t, postgres, owner.Id, "Test Group")
	ctx := context.Background()

	repo.AddMember(ctx, group.Id, member.Id, Domain.GroupRoleAdmin)

	tests := []struct {
		name        string
		groupID     uuid.UUID
		expectError bool
	}{
		{
			name:        "GetGroupWithMembers_Success",
			groupID:     group.Id,
			expectError: false,
		},
		{
			name:        "GetGroupWithMembers_NotFound",
			groupID:     uuid.New(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retrievedGroup, err := repo.GetGroupWithMembers(ctx, tt.groupID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, retrievedGroup)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, retrievedGroup)
				assert.Equal(t, group.Id, retrievedGroup.Id)
				assert.Equal(t, 2, len(retrievedGroup.Memberships))
			}
		})
	}
}

func TestGroupRepository_AddMember(t *testing.T) {
	postgres := OpenTestDB(t)
	defer DBCleanup(t, postgres)

	repo := CreateGroupRepository(*postgres)
	owner := createTestUser(t, postgres)
	member1 := createTestUser(t, postgres)
	member2 := createTestUser(t, postgres)
	group := createTestGroup(t, postgres, owner.Id, "Test Group")
	ctx := context.Background()

	tests := []struct {
		name        string
		groupID     uuid.UUID
		userID      uuid.UUID
		role        Domain.GroupRole
		expectError bool
		errorCode   int
	}{
		{
			name:        "AddMember_Success",
			groupID:     group.Id,
			userID:      member1.Id,
			role:        Domain.GroupRoleAdmin,
			expectError: false,
		},
		{
			name:        "AddMember_AsMember",
			groupID:     group.Id,
			userID:      member2.Id,
			role:        Domain.GroupRoleMember,
			expectError: false,
		},
		{
			name:        "AddMember_Duplicate",
			groupID:     group.Id,
			userID:      member1.Id,
			role:        Domain.GroupRoleMember,
			expectError: true,
			errorCode:   409,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			membership, err := repo.AddMember(ctx, tt.groupID, tt.userID, tt.role)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, membership)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, membership)
				assert.Equal(t, tt.groupID, membership.GroupID)
				assert.Equal(t, tt.userID, membership.UserID)
				assert.Equal(t, tt.role, membership.Role)
			}
		})
	}
}

func TestGroupRepository_GetMembership(t *testing.T) {
	postgres := OpenTestDB(t)
	defer DBCleanup(t, postgres)

	repo := CreateGroupRepository(*postgres)
	owner := createTestUser(t, postgres)
	member := createTestUser(t, postgres)
	group := createTestGroup(t, postgres, owner.Id, "Test Group")
	ctx := context.Background()

	repo.AddMember(ctx, group.Id, member.Id, Domain.GroupRoleAdmin)

	tests := []struct {
		name        string
		groupID     uuid.UUID
		userID      uuid.UUID
		expectError bool
	}{
		{
			name:        "GetMembership_Owner",
			groupID:     group.Id,
			userID:      owner.Id,
			expectError: false,
		},
		{
			name:        "GetMembership_Member",
			groupID:     group.Id,
			userID:      member.Id,
			expectError: false,
		},
		{
			name:        "GetMembership_NotMember",
			groupID:     group.Id,
			userID:      uuid.New(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			membership, err := repo.GetMembership(ctx, tt.groupID, tt.userID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, membership)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, membership)
				assert.Equal(t, tt.groupID, membership.GroupID)
				assert.Equal(t, tt.userID, membership.UserID)
			}
		})
	}
}

func TestGroupRepository_UpdateMemberRole(t *testing.T) {
	postgres := OpenTestDB(t)
	defer DBCleanup(t, postgres)

	repo := CreateGroupRepository(*postgres)
	owner := createTestUser(t, postgres)
	member := createTestUser(t, postgres)
	group := createTestGroup(t, postgres, owner.Id, "Test Group")
	ctx := context.Background()

	repo.AddMember(ctx, group.Id, member.Id, Domain.GroupRoleMember)

	tests := []struct {
		name        string
		groupID     uuid.UUID
		userID      uuid.UUID
		newRole     Domain.GroupRole
		expectError bool
	}{
		{
			name:        "UpdateMemberRole_Success",
			groupID:     group.Id,
			userID:      member.Id,
			newRole:     Domain.GroupRoleAdmin,
			expectError: false,
		},
		{
			name:        "UpdateMemberRole_NotMember",
			groupID:     group.Id,
			userID:      uuid.New(),
			newRole:     Domain.GroupRoleAdmin,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.UpdateMemberRole(ctx, tt.groupID, tt.userID, tt.newRole)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				membership, err := repo.GetMembership(ctx, tt.groupID, tt.userID)
				assert.NoError(t, err)
				assert.Equal(t, tt.newRole, membership.Role)
			}
		})
	}
}

func TestGroupRepository_RemoveMember(t *testing.T) {
	postgres := OpenTestDB(t)
	defer DBCleanup(t, postgres)

	repo := CreateGroupRepository(*postgres)
	owner := createTestUser(t, postgres)
	member := createTestUser(t, postgres)
	group := createTestGroup(t, postgres, owner.Id, "Test Group")
	ctx := context.Background()

	repo.AddMember(ctx, group.Id, member.Id, Domain.GroupRoleMember)

	tests := []struct {
		name        string
		groupID     uuid.UUID
		userID      uuid.UUID
		expectError bool
	}{
		{
			name:        "RemoveMember_Success",
			groupID:     group.Id,
			userID:      member.Id,
			expectError: false,
		},
		{
			name:        "RemoveMember_NotMember",
			groupID:     group.Id,
			userID:      uuid.New(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.RemoveMember(ctx, tt.groupID, tt.userID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				_, err := repo.GetMembership(ctx, tt.groupID, tt.userID)
				assert.Error(t, err)
			}
		})
	}
}

func TestGroupRepository_DeleteGroup(t *testing.T) {
	postgres := OpenTestDB(t)
	defer DBCleanup(t, postgres)

	repo := CreateGroupRepository(*postgres)
	owner := createTestUser(t, postgres)
	ctx := context.Background()

	tests := []struct {
		name          string
		setupFunc     func() uuid.UUID
		expectError   bool
		verifyDeleted bool
	}{
		{
			name: "DeleteGroup_Success",
			setupFunc: func() uuid.UUID {
				group := createTestGroup(t, postgres, owner.Id, "Empty Group")
				return group.Id
			},
			expectError:   false,
			verifyDeleted: true,
		},
		{
			name: "DeleteGroup_NotFound",
			setupFunc: func() uuid.UUID {
				return uuid.New()
			},
			expectError:   false,
			verifyDeleted: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groupID := tt.setupFunc()
			err := repo.DeleteGroup(ctx, groupID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				if tt.verifyDeleted {
					_, err := repo.GetGroupById(ctx, groupID)
					assert.Error(t, err)
				}
			}
		})
	}
}

func TestGroupRepository_IsGroupAdmin(t *testing.T) {
	postgres := OpenTestDB(t)
	defer DBCleanup(t, postgres)

	repo := CreateGroupRepository(*postgres)
	owner := createTestUser(t, postgres)
	admin := createTestUser(t, postgres)
	member := createTestUser(t, postgres)
	group := createTestGroup(t, postgres, owner.Id, "Test Group")
	ctx := context.Background()

	repo.AddMember(ctx, group.Id, admin.Id, Domain.GroupRoleAdmin)
	repo.AddMember(ctx, group.Id, member.Id, Domain.GroupRoleMember)

	tests := []struct {
		name        string
		groupID     uuid.UUID
		userID      uuid.UUID
		expectAdmin bool
		expectError bool
	}{
		{
			name:        "IsGroupAdmin_Owner",
			groupID:     group.Id,
			userID:      owner.Id,
			expectAdmin: true,
			expectError: false,
		},
		{
			name:        "IsGroupAdmin_Admin",
			groupID:     group.Id,
			userID:      admin.Id,
			expectAdmin: true,
			expectError: false,
		},
		{
			name:        "IsGroupAdmin_Member",
			groupID:     group.Id,
			userID:      member.Id,
			expectAdmin: false,
			expectError: false,
		},
		{
			name:        "IsGroupAdmin_NotMember",
			groupID:     group.Id,
			userID:      uuid.New(),
			expectAdmin: false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isAdmin, err := repo.IsGroupAdmin(ctx, tt.groupID, tt.userID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectAdmin, isAdmin)
			}
		})
	}
}

func TestGroupRepository_IsGroupOwner(t *testing.T) {
	postgres := OpenTestDB(t)
	defer DBCleanup(t, postgres)

	repo := CreateGroupRepository(*postgres)
	owner := createTestUser(t, postgres)
	admin := createTestUser(t, postgres)
	member := createTestUser(t, postgres)
	group := createTestGroup(t, postgres, owner.Id, "Test Group")
	ctx := context.Background()

	repo.AddMember(ctx, group.Id, admin.Id, Domain.GroupRoleAdmin)
	repo.AddMember(ctx, group.Id, member.Id, Domain.GroupRoleMember)

	tests := []struct {
		name        string
		groupID     uuid.UUID
		userID      uuid.UUID
		expectOwner bool
		expectError bool
	}{
		{
			name:        "IsGroupOwner_Owner",
			groupID:     group.Id,
			userID:      owner.Id,
			expectOwner: true,
			expectError: false,
		},
		{
			name:        "IsGroupOwner_Admin",
			groupID:     group.Id,
			userID:      admin.Id,
			expectOwner: false,
			expectError: false,
		},
		{
			name:        "IsGroupOwner_Member",
			groupID:     group.Id,
			userID:      member.Id,
			expectOwner: false,
			expectError: false,
		},
		{
			name:        "IsGroupOwner_NotMember",
			groupID:     group.Id,
			userID:      uuid.New(),
			expectOwner: false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isOwner, err := repo.IsGroupOwner(ctx, tt.groupID, tt.userID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectOwner, isOwner)
			}
		})
	}
}
