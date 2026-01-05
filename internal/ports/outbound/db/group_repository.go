package RepositoryPorts

import (
	"context"

	Domain "autobill-service/internal/domain"

	"github.com/google/uuid"
)

type GroupRepositoryPort interface {
	CreateGroup(ctx context.Context, name string, ownerId uuid.UUID, simplifyDebts bool) (*Domain.Group, error)
	UpdateGroup(ctx context.Context, groupId uuid.UUID, updates map[string]any) (*Domain.Group, error)
	GetGroupsByUserId(ctx context.Context, userId uuid.UUID, limit, offset int) ([]Domain.Group, int64, error)
	GetGroupById(ctx context.Context, groupId uuid.UUID) (*Domain.Group, error)
	GetGroupWithMembers(ctx context.Context, groupId uuid.UUID) (*Domain.Group, error)
	DeleteGroup(ctx context.Context, groupId uuid.UUID) error

	AddMember(ctx context.Context, groupId, userId uuid.UUID, role Domain.GroupRole) (*Domain.GroupMembership, error)
	GetMembership(ctx context.Context, groupId, userId uuid.UUID) (*Domain.GroupMembership, error)
	UpdateMemberRole(ctx context.Context, groupId, userId uuid.UUID, role Domain.GroupRole) error
	RemoveMember(ctx context.Context, groupId, userId uuid.UUID) error
	IsGroupAdmin(ctx context.Context, groupId, userId uuid.UUID) (bool, error)
	IsGroupOwner(ctx context.Context, groupId, userId uuid.UUID) (bool, error)
}
