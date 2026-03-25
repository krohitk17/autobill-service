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
	repo      RepositoryPorts.SplitRepositoryPort
	groupRepo RepositoryPorts.GroupRepositoryPort
}

func CreateSplitService(repo RepositoryPorts.SplitRepositoryPort, groupRepo RepositoryPorts.GroupRepositoryPort) HttpPorts.SplitUseCase {
	return &SplitService{
		repo:      repo,
		groupRepo: groupRepo,
	}
}

func (s *SplitService) CreateSplit(ctx context.Context, userId uuid.UUID, input Dtos.CreateSplitInput) (*Dtos.SplitResult, error) {
	if input.IdempotencyKey != "" {
		existing, err := s.repo.GetSplitByIdempotencyKey(ctx, input.IdempotencyKey)
		if err == nil && existing != nil {
			if existing.CreatedByID != userId {
				return nil, fiber.NewError(fiber.StatusConflict, Errors.ErrIdempotencyKeyConflict)
			}
			return s.splitToDto(existing), nil
		}
	}

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

	if err := s.validateSplitAmountMatchesShares(input.TotalAmount, domainParticipants); err != nil {
		return nil, err
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
	if input.IdempotencyKey != "" {
		split.IdempotencyKey = &input.IdempotencyKey
	}

	createdSplit, createdParticipants, dbErr := s.repo.CreateSplitWithParticipants(ctx, split, domainParticipants)
	if dbErr != nil {
		return nil, dbErr
	}
	createdSplit.Participants = createdParticipants

	Logger.Debug().
		Str("operation", "CreateSplit").
		Str("userId", userId.String()).
		Str("splitId", createdSplit.Id.String()).
		Str("type", input.Type).
		Int64("amount", input.TotalAmount).
		Str("currency", input.Currency).
		Int("participants", len(input.Participants)).
		Msg("Split created successfully")

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

func (s *SplitService) ReverseSplit(ctx context.Context, userId, splitId uuid.UUID) (*Dtos.SplitResult, error) {
	originalSplit, dbErr := s.repo.GetSplitWithParticipants(ctx, splitId)
	if dbErr != nil {
		return nil, dbErr
	}

	if originalSplit.CreatedByID != userId {
		return nil, fiber.NewError(fiber.StatusNotFound, Errors.ErrSplitNotFound)
	}

	pendingSettlementCount, err := s.repo.GetPendingSettlementCountBySplitId(ctx, splitId)
	if err != nil {
		return nil, err
	}
	if pendingSettlementCount > 0 {
		return nil, fiber.NewError(fiber.StatusConflict, Errors.ErrSplitHasPendingSettlements)
	}

	confirmedTotalsByPayer, err := s.repo.GetConfirmedSettlementTotalsByPayer(ctx, splitId)
	if err != nil {
		return nil, err
	}

	for payerID, paidAmount := range confirmedTotalsByPayer {
		if paidAmount <= 0 || payerID == originalSplit.CreatedByID {
			continue
		}

		reverseSplit := &Domain.Split{
			Type:          originalSplit.Type,
			DivisionType:  Domain.SplitDivisionCustom,
			TotalAmount:   paidAmount,
			Currency:      originalSplit.Currency,
			Description:   "Refund reverse split: " + originalSplit.Description,
			GroupID:       originalSplit.GroupID,
			CreatedByID:   payerID,
			SimplifyDebts: originalSplit.SimplifyDebts,
		}

		reverseParticipants := []Domain.SplitParticipant{
			{
				UserID:      originalSplit.CreatedByID,
				ShareAmount: paidAmount,
				Currency:    originalSplit.Currency,
				IsSettled:   false,
			},
		}

		_, _, createErr := s.repo.CreateSplitWithParticipants(ctx, reverseSplit, reverseParticipants)
		if createErr != nil {
			return nil, createErr
		}
	}

	err = s.repo.DeleteSplitWithBalanceRollback(ctx, originalSplit, originalSplit.Participants)
	if err != nil {
		return nil, err
	}

	return s.splitToDto(originalSplit), nil
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

func (s *SplitService) validateSplitAmountMatchesShares(totalAmount int64, participants []Domain.SplitParticipant) error {
	var participantTotal int64
	for _, p := range participants {
		participantTotal += p.ShareAmount
	}

	if participantTotal != totalAmount {
		return fiber.NewError(fiber.StatusBadRequest, Errors.ErrInvalidSplitAmount)
	}

	return nil
}
