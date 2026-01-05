package HttpPorts

import (
	"context"

	Dtos "autobill-service/internal/application/group/dtos"
	Helpers "autobill-service/pkg/helpers"

	"github.com/google/uuid"
)

type GroupUseCase interface {
	CreateGroup(ctx context.Context, userId uuid.UUID, input Dtos.CreateGroupInput) (*Dtos.GroupResult, error)
	UpdateGroup(ctx context.Context, userId, groupId uuid.UUID, input Dtos.UpdateGroupInput) (*Dtos.GroupResult, error)
	GetGroups(ctx context.Context, userId uuid.UUID, pagination Helpers.PaginationParams) (*Dtos.GroupListResult, error)
	GetGroup(ctx context.Context, userId, groupId uuid.UUID) (*Dtos.GroupDetailResult, error)
	DeleteGroup(ctx context.Context, userId, groupId uuid.UUID) error
	AddMember(ctx context.Context, userId, groupId uuid.UUID, input Dtos.AddMemberInput) (*Dtos.MemberResult, error)
	UpdateMemberRole(ctx context.Context, userId, groupId, memberId uuid.UUID, role string) error
	RemoveMember(ctx context.Context, userId, groupId, memberId uuid.UUID) error
	LeaveGroup(ctx context.Context, userId, groupId uuid.UUID) error
}
