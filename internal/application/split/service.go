package SplitApplication

import (
	"context"

	"github.com/gofiber/fiber/v2"

	Dtos "autobill-service/internal/application/split/dtos"
	Domain "autobill-service/internal/domain"
	HttpPorts "autobill-service/internal/ports/inbound/http"
	RepositoryPorts "autobill-service/internal/ports/outbound/db"
	Errors "autobill-service/pkg/errors"
	Helpers "autobill-service/pkg/helpers"
	Logger "autobill-service/pkg/logger"

	"github.com/google/uuid"
)

type SplitService struct {
	repo        RepositoryPorts.SplitRepositoryPort
	groupRepo   RepositoryPorts.GroupRepositoryPort
	balanceRepo RepositoryPorts.BalanceRepositoryPort
}

func CreateSplitService(repo RepositoryPorts.SplitRepositoryPort, groupRepo RepositoryPorts.GroupRepositoryPort, balanceRepo RepositoryPorts.BalanceRepositoryPort) HttpPorts.SplitUseCase {
	return &SplitService{
		repo:        repo,
		groupRepo:   groupRepo,
		balanceRepo: balanceRepo,
	}
}

func (s *SplitService) CreateSplit(ctx context.Context, userId uuid.UUID, input Dtos.CreateSplitInput) (*Dtos.SplitResult, error) {
	if !Domain.IsValidSplitType(input.Type) {
		return nil, fiber.NewError(fiber.StatusBadRequest, Errors.ErrInvalidSplitType)
	}
	if !Domain.IsValidSplitDivisionType(input.DivisionType) {
		return nil, fiber.NewError(fiber.StatusBadRequest, Errors.ErrInvalidDivisionType)
	}
	if !Domain.IsValidCurrency(input.Currency) {
		return nil, fiber.NewError(fiber.StatusBadRequest, Errors.ErrInvalidCurrency)
	}

	if len(input.Participants) == 0 {
		return nil, fiber.NewError(fiber.StatusBadRequest, Errors.ErrNoParticipants)
	}

	participantUUIDs := make([]uuid.UUID, len(input.Participants))
	for i, p := range input.Participants {
		parsed, err := Helpers.ParseUUID(p.UserID)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, Errors.ErrInvalidParticipantId)
		}
		participantUUIDs[i] = parsed
	}

	var groupId *uuid.UUID
	if input.GroupID != "" {
		parsed, err := Helpers.ParseUUID(input.GroupID)
		if err != nil {
			return nil, err
		}
		_, memberErr := s.groupRepo.GetMembership(ctx, parsed, userId)
		if memberErr != nil {
			return nil, memberErr
		}
		groupId = &parsed
	} else if input.Type == string(Domain.SplitTypeGroup) {
		return nil, fiber.NewError(fiber.StatusBadRequest, Errors.ErrGroupIdRequired)
	}

	participantCount := len(input.Participants)
	equalShare := input.TotalAmount / int64(participantCount)
	remainder := input.TotalAmount % int64(participantCount)

	domainParticipants := make([]Domain.SplitParticipant, len(input.Participants))
	for i, p := range input.Participants {
		shareAmount := p.ShareAmount
		if input.DivisionType == "EQUAL" {
			shareAmount = equalShare
			if int64(i) < remainder {
				shareAmount++
			}
		}

		domainParticipants[i] = Domain.SplitParticipant{
			UserID:      participantUUIDs[i],
			ShareAmount: shareAmount,
			Currency:    Domain.Currency(input.Currency),
			IsSettled:   false,
		}
	}

	split := &Domain.Split{
		Type:          Domain.SplitType(input.Type),
		DivisionType:  Domain.SplitDivisionType(input.DivisionType),
		TotalAmount:   input.TotalAmount,
		Currency:      Domain.Currency(input.Currency),
		Description:   input.Description,
		GroupID:       groupId,
		SimplifyDebts: input.SimplifyDebts,
		CreatedByID:   userId,
	}

	createdSplit, createdParticipants, dbErr := s.repo.CreateSplitWithParticipants(ctx, split, domainParticipants)
	if dbErr != nil {
		return nil, dbErr
	}

	Logger.Debug().
		Str("operation", "CreateSplit").
		Str("userId", userId.String()).
		Str("splitId", createdSplit.Id.String()).
		Str("type", input.Type).
		Int64("amount", input.TotalAmount).
		Str("currency", input.Currency).
		Int("participants", len(input.Participants)).
		Msg("Split created successfully")

	createdSplit.Participants = createdParticipants
	return s.splitToDto(createdSplit), nil
}

func (s *SplitService) GetSplit(ctx context.Context, userId, splitId uuid.UUID) (*Dtos.SplitResult, error) {
	split, dbErr := s.repo.GetSplitWithParticipants(ctx, splitId)
	if dbErr != nil {
		return nil, dbErr
	}

	if !s.isUserAuthorizedForSplit(ctx, split, userId) {
		return nil, fiber.NewError(fiber.StatusNotFound, Errors.ErrSplitNotFound)
	}

	return s.splitToDto(split), nil
}

func (s *SplitService) GetGroupSplits(ctx context.Context, userId, groupId uuid.UUID, pagination Helpers.PaginationParams) (*Dtos.SplitListResult, error) {
	_, memberErr := s.groupRepo.GetMembership(ctx, groupId, userId)
	if memberErr != nil {
		return nil, memberErr
	}

	offset := pagination.Offset()
	splits, total, dbErr := s.repo.GetSplitsByGroupId(ctx, groupId, pagination.PageSize, offset)
	if dbErr != nil {
		return nil, dbErr
	}

	splitResults := make([]Dtos.SplitResult, len(splits))
	for i, split := range splits {
		splitResults[i] = *s.splitToDto(&split)
	}

	return &Dtos.SplitListResult{
		Splits:     splitResults,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalItems: total,
	}, nil
}

func (s *SplitService) GetMySplits(ctx context.Context, userId uuid.UUID, pagination Helpers.PaginationParams) (*Dtos.SplitListResult, error) {
	offset := pagination.Offset()
	splits, total, dbErr := s.repo.GetSplitsByUserId(ctx, userId, pagination.PageSize, offset)
	if dbErr != nil {
		return nil, dbErr
	}

	splitResults := make([]Dtos.SplitResult, len(splits))
	for i, split := range splits {
		splitResults[i] = *s.splitToDto(&split)
	}

	return &Dtos.SplitListResult{
		Splits:     splitResults,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalItems: total,
	}, nil
}

func (s *SplitService) DeleteSplit(ctx context.Context, userId, splitId uuid.UUID) error {
	split, dbErr := s.repo.GetSplitById(ctx, splitId)
	if dbErr != nil {
		return dbErr
	}

	if split.CreatedByID != userId {
		return fiber.NewError(fiber.StatusNotFound, Errors.ErrSplitNotFound)
	}

	if split.IsFinalized {
		return fiber.NewError(fiber.StatusBadRequest, Errors.ErrCannotDeleteFinalizedSplit)
	}

	err := s.repo.DeleteSplit(ctx, splitId)
	if err != nil {
		return err
	}

	Logger.Debug().
		Str("operation", "DeleteSplit").
		Str("userId", userId.String()).
		Str("splitId", splitId.String()).
		Msg("Split deleted successfully")

	return nil
}

func (s *SplitService) AddParticipant(ctx context.Context, userId, splitId uuid.UUID, input Dtos.AddParticipantInput) (*Dtos.ParticipantResult, error) {
	participantUUID, err := Helpers.ParseUUID(input.UserID)
	if err != nil {
		return nil, err
	}

	split, dbErr := s.repo.GetSplitById(ctx, splitId)
	if dbErr != nil {
		return nil, dbErr
	}

	if split.CreatedByID != userId {
		return nil, fiber.NewError(fiber.StatusNotFound, Errors.ErrSplitNotFound)
	}

	if split.IsFinalized {
		return nil, fiber.NewError(fiber.StatusBadRequest, Errors.ErrSplitAlreadyFinalized)
	}

	participant := &Domain.SplitParticipant{
		SplitID:     splitId,
		UserID:      participantUUID,
		ShareAmount: input.ShareAmount,
		Currency:    split.Currency,
		IsSettled:   false,
	}

	created, pErr := s.repo.AddParticipant(ctx, participant)
	if pErr != nil {
		return nil, pErr
	}

	return &Dtos.ParticipantResult{
		UserID:      created.UserID.String(),
		UserName:    created.User.Name,
		ShareAmount: created.ShareAmount,
		Currency:    string(created.Currency),
		IsSettled:   created.IsSettled,
	}, nil
}

func (s *SplitService) UpdateParticipant(ctx context.Context, userId, splitId, participantUserId uuid.UUID, input Dtos.UpdateParticipantInput) error {
	split, dbErr := s.repo.GetSplitById(ctx, splitId)
	if dbErr != nil {
		return dbErr
	}

	if split.CreatedByID != userId {
		return fiber.NewError(fiber.StatusNotFound, Errors.ErrSplitNotFound)
	}

	participant, pErr := s.repo.GetParticipant(ctx, splitId, participantUserId)
	if pErr != nil {
		return pErr
	}

	participant.ShareAmount = input.ShareAmount
	if input.IsSettled != nil {
		participant.IsSettled = *input.IsSettled
	}

	return s.repo.UpdateParticipant(ctx, participant)
}

func (s *SplitService) FinalizeSplit(ctx context.Context, userId, splitId uuid.UUID) error {
	split, dbErr := s.repo.GetSplitWithParticipants(ctx, splitId)
	if dbErr != nil {
		return dbErr
	}

	if split.CreatedByID != userId {
		return fiber.NewError(fiber.StatusNotFound, Errors.ErrSplitNotFound)
	}

	if split.IsFinalized {
		return fiber.NewError(fiber.StatusBadRequest, Errors.ErrSplitAlreadyFinalized)
	}

	if err := s.repo.FinalizeSplit(ctx, splitId); err != nil {
		return err
	}

	if err := s.balanceRepo.UpdateBalancesForSplit(ctx, split, split.Participants); err != nil {
		return err
	}

	return nil
}

func (s *SplitService) ReverseSplit(ctx context.Context, userId, splitId uuid.UUID) (*Dtos.SplitResult, error) {
	originalSplit, dbErr := s.repo.GetSplitWithParticipants(ctx, splitId)
	if dbErr != nil {
		return nil, dbErr
	}

	if originalSplit.CreatedByID != userId {
		return nil, fiber.NewError(fiber.StatusNotFound, Errors.ErrSplitNotFound)
	}

	if !originalSplit.IsFinalized {
		return nil, fiber.NewError(fiber.StatusBadRequest, Errors.ErrSplitNotFinalized)
	}

	isReversed, err := s.repo.IsSplitReversed(ctx, splitId)
	if err != nil {
		return nil, err
	}
	if isReversed {
		return nil, fiber.NewError(fiber.StatusBadRequest, Errors.ErrSplitAlreadyReversed)
	}

	reversalSplit := &Domain.Split{
		Type:         originalSplit.Type,
		DivisionType: originalSplit.DivisionType,
		TotalAmount:  -originalSplit.TotalAmount,
		Currency:     originalSplit.Currency,
		Description:  "Reversal: " + originalSplit.Description,
		GroupID:      originalSplit.GroupID,
		CreatedByID:  userId,
	}

	reversalParticipants := make([]Domain.SplitParticipant, len(originalSplit.Participants))
	for i, p := range originalSplit.Participants {
		reversalParticipants[i] = Domain.SplitParticipant{
			UserID:      p.UserID,
			ShareAmount: -p.ShareAmount,
			Currency:    p.Currency,
			IsSettled:   false,
		}
	}

	createdReversal, createdParticipants, rErr := s.repo.CreateReversalSplitWithParticipants(ctx, splitId, reversalSplit, reversalParticipants)
	if rErr != nil {
		return nil, rErr
	}

	createdReversal.Participants = createdParticipants
	return s.splitToDto(createdReversal), nil
}

func (s *SplitService) splitToDto(split *Domain.Split) *Dtos.SplitResult {
	participants := make([]Dtos.ParticipantResult, len(split.Participants))
	for i, p := range split.Participants {
		userName := p.User.Name
		participants[i] = Dtos.ParticipantResult{
			UserID:      p.UserID.String(),
			UserName:    userName,
			ShareAmount: p.ShareAmount,
			Currency:    string(p.Currency),
			IsSettled:   p.IsSettled,
		}
	}

	groupIdStr := ""
	if split.GroupID != nil {
		groupIdStr = split.GroupID.String()
	}

	return &Dtos.SplitResult{
		ID:            split.Id.String(),
		Type:          string(split.Type),
		DivisionType:  string(split.DivisionType),
		TotalAmount:   split.TotalAmount,
		Currency:      string(split.Currency),
		Description:   split.Description,
		GroupID:       groupIdStr,
		CreatedByID:   split.CreatedByID.String(),
		CreatedAt:     split.CreatedAt,
		IsFinalized:   split.IsFinalized,
		SimplifyDebts: split.SimplifyDebts,
		Participants:  participants,
	}
}

func (s *SplitService) isUserAuthorizedForSplit(ctx context.Context, split *Domain.Split, userId uuid.UUID) bool {
	if split.CreatedByID == userId {
		return true
	}
	for _, p := range split.Participants {
		if p.UserID == userId {
			return true
		}
	}
	if split.GroupID != nil {
		_, memberErr := s.groupRepo.GetMembership(ctx, *split.GroupID, userId)
		return memberErr == nil
	}
	return false
}
