package GroupApplication

import (
	"context"

	"github.com/gofiber/fiber/v2"

	Dtos "autobill-service/internal/application/group/dtos"
	Domain "autobill-service/internal/domain"
	HttpPorts "autobill-service/internal/ports/inbound/http"
	RepositoryPorts "autobill-service/internal/ports/outbound/db"
	Errors "autobill-service/pkg/errors"
	Helpers "autobill-service/pkg/helpers"
	Logger "autobill-service/pkg/logger"

	"github.com/google/uuid"
)

type GroupService struct {
	repo      RepositoryPorts.GroupRepositoryPort
	splitRepo RepositoryPorts.SplitRepositoryPort
}

func CreateGroupService(repo RepositoryPorts.GroupRepositoryPort, splitRepo RepositoryPorts.SplitRepositoryPort) HttpPorts.GroupUseCase {
	return &GroupService{repo: repo, splitRepo: splitRepo}
}

func (s *GroupService) CreateGroup(ctx context.Context, userId uuid.UUID, input Dtos.CreateGroupInput) (*Dtos.GroupResult, error) {
	if input.Name == "" {
		return nil, fiber.NewError(fiber.StatusBadRequest, Errors.ErrGroupNameRequired)
	}

	simplifyDebts := true
	if input.SimplifyDebts != nil {
		simplifyDebts = *input.SimplifyDebts
	}

	group, dbErr := s.repo.CreateGroup(ctx, input.Name, userId, simplifyDebts)
	if dbErr != nil {
		return nil, dbErr
	}

	Logger.Debug().
		Str("operation", "CreateGroup").
		Str("userId", userId.String()).
		Str("groupId", group.Id.String()).
		Str("name", group.Name).
		Msg("Group created successfully")

	return &Dtos.GroupResult{
		ID:            group.Id.String(),
		Name:          group.Name,
		SimplifyDebts: group.SimplifyDebts,
		CreatedAt:     group.CreatedAt,
	}, nil
}

func (s *GroupService) UpdateGroup(ctx context.Context, userId, groupId uuid.UUID, input Dtos.UpdateGroupInput) (*Dtos.GroupResult, error) {
	isAdmin, adminErr := s.repo.IsGroupAdmin(ctx, groupId, userId)
	if adminErr != nil {
		return nil, adminErr
	}
	if !isAdmin {
		return nil, fiber.NewError(fiber.StatusNotFound, Errors.ErrGroupNotFound)
	}

	updates := make(map[string]interface{})
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.SimplifyDebts != nil {
		updates["simplify_debts"] = *input.SimplifyDebts
	}

	if len(updates) == 0 {
		return nil, fiber.NewError(fiber.StatusBadRequest, Errors.ErrNoFieldsToUpdate)
	}

	group, dbErr := s.repo.UpdateGroup(ctx, groupId, updates)
	if dbErr != nil {
		return nil, dbErr
	}

	return &Dtos.GroupResult{
		ID:            group.Id.String(),
		Name:          group.Name,
		SimplifyDebts: group.SimplifyDebts,
		CreatedAt:     group.CreatedAt,
	}, nil
}

func (s *GroupService) GetGroups(ctx context.Context, userId uuid.UUID, pagination Helpers.PaginationParams) (*Dtos.GroupListResult, error) {
	offset := pagination.Offset()
	groups, total, dbErr := s.repo.GetGroupsByUserId(ctx, userId, pagination.PageSize, offset)
	if dbErr != nil {
		return nil, dbErr
	}

	groupResults := make([]Dtos.GroupResult, len(groups))
	for i, g := range groups {
		groupResults[i] = Dtos.GroupResult{
			ID:            g.Id.String(),
			Name:          g.Name,
			SimplifyDebts: g.SimplifyDebts,
			CreatedAt:     g.CreatedAt,
		}
	}

	return &Dtos.GroupListResult{
		Groups:     groupResults,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalItems: total,
	}, nil
}

func (s *GroupService) GetGroup(ctx context.Context, userId, groupId uuid.UUID) (*Dtos.GroupDetailResult, error) {
	_, memberErr := s.repo.GetMembership(ctx, groupId, userId)
	if memberErr != nil {
		return nil, memberErr
	}

	group, dbErr := s.repo.GetGroupWithMembers(ctx, groupId)
	if dbErr != nil {
		return nil, dbErr
	}

	members := make([]Dtos.MemberResult, len(group.Memberships))
	for i, m := range group.Memberships {
		members[i] = Dtos.MemberResult{
			UserID: m.UserID.String(),
			Name:   m.User.Name,
			Email:  m.User.Email,
			Role:   string(m.Role),
		}
	}

	return &Dtos.GroupDetailResult{
		ID:            group.Id.String(),
		Name:          group.Name,
		SimplifyDebts: group.SimplifyDebts,
		CreatedAt:     group.CreatedAt,
		Members:       members,
	}, nil
}

func (s *GroupService) DeleteGroup(ctx context.Context, userId, groupId uuid.UUID) error {
	isOwner, ownerErr := s.repo.IsGroupOwner(ctx, groupId, userId)
	if ownerErr != nil {
		return ownerErr
	}
	if !isOwner {
		return fiber.NewError(fiber.StatusNotFound, Errors.ErrGroupNotFound)
	}

	return s.repo.DeleteGroup(ctx, groupId)
}

func (s *GroupService) AddMember(ctx context.Context, userId, groupId uuid.UUID, input Dtos.AddMemberInput) (*Dtos.MemberResult, error) {
	newMemberUUID, err := Helpers.ParseUUID(input.UserID)
	if err != nil {
		return nil, err
	}

	isAdmin, adminErr := s.repo.IsGroupAdmin(ctx, groupId, userId)
	if adminErr != nil {
		return nil, adminErr
	}
	if !isAdmin {
		return nil, fiber.NewError(fiber.StatusNotFound, Errors.ErrGroupNotFound)
	}

	if !Domain.IsValidAssignableRole(input.Role) {
		return nil, fiber.NewError(fiber.StatusBadRequest, Errors.ErrInvalidRole)
	}

	role := Domain.GroupRole(input.Role)
	membership, dbErr := s.repo.AddMember(ctx, groupId, newMemberUUID, role)
	if dbErr != nil {
		return nil, dbErr
	}

	return &Dtos.MemberResult{
		UserID: membership.UserID.String(),
		Name:   membership.User.Name,
		Email:  membership.User.Email,
		Role:   string(membership.Role),
	}, nil
}

func (s *GroupService) UpdateMemberRole(ctx context.Context, userId, groupId, memberId uuid.UUID, role string) error {
	isOwner, ownerErr := s.repo.IsGroupOwner(ctx, groupId, userId)
	if ownerErr != nil {
		return ownerErr
	}
	if !isOwner {
		return fiber.NewError(fiber.StatusNotFound, Errors.ErrGroupNotFound)
	}

	isTargetOwner, targetOwnerErr := s.repo.IsGroupOwner(ctx, groupId, memberId)
	if targetOwnerErr != nil {
		return targetOwnerErr
	}
	if isTargetOwner {
		return fiber.NewError(fiber.StatusBadRequest, Errors.ErrCannotChangeOwnerRole)
	}

	if !Domain.IsValidAssignableRole(role) {
		return fiber.NewError(fiber.StatusBadRequest, Errors.ErrInvalidRole)
	}

	groupRole := Domain.GroupRole(role)
	return s.repo.UpdateMemberRole(ctx, groupId, memberId, groupRole)
}

func (s *GroupService) RemoveMember(ctx context.Context, userId, groupId, memberId uuid.UUID) error {
	isAdmin, adminErr := s.repo.IsGroupAdmin(ctx, groupId, userId)
	if adminErr != nil {
		return adminErr
	}
	if !isAdmin {
		return fiber.NewError(fiber.StatusNotFound, Errors.ErrGroupNotFound)
	}

	isTargetOwner, targetOwnerErr := s.repo.IsGroupOwner(ctx, groupId, memberId)
	if targetOwnerErr != nil {
		return targetOwnerErr
	}
	if isTargetOwner {
		return fiber.NewError(fiber.StatusBadRequest, Errors.ErrCannotRemoveOwner)
	}

	return s.repo.RemoveMember(ctx, groupId, memberId)
}

func (s *GroupService) LeaveGroup(ctx context.Context, userId, groupId uuid.UUID) error {
	_, memberErr := s.repo.GetMembership(ctx, groupId, userId)
	if memberErr != nil {
		return memberErr
	}

	isOwner, ownerErr := s.repo.IsGroupOwner(ctx, groupId, userId)
	if ownerErr != nil {
		return ownerErr
	}
	if isOwner {
		return fiber.NewError(fiber.StatusBadRequest, Errors.ErrOwnerCannotLeaveGroup)
	}

	hasPending, pendingErr := s.splitRepo.HasPendingSplitsInGroup(ctx, userId, groupId)
	if pendingErr != nil {
		return pendingErr
	}
	if hasPending {
		return fiber.NewError(fiber.StatusBadRequest, Errors.ErrHasPendingSplits)
	}

	return s.repo.RemoveMember(ctx, groupId, userId)
}
