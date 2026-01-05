package SettlementApplication

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	Dtos "autobill-service/internal/application/settlement/dtos"
	Domain "autobill-service/internal/domain"
	HttpPorts "autobill-service/internal/ports/inbound/http"
	RepositoryPorts "autobill-service/internal/ports/outbound/db"
	Errors "autobill-service/pkg/errors"
	Helpers "autobill-service/pkg/helpers"
	Logger "autobill-service/pkg/logger"

	"github.com/google/uuid"
)

type SettlementService struct {
	repo      RepositoryPorts.SettlementRepositoryPort
	splitRepo RepositoryPorts.SplitRepositoryPort
}

func CreateSettlementService(repo RepositoryPorts.SettlementRepositoryPort, splitRepo RepositoryPorts.SplitRepositoryPort) HttpPorts.SettlementUseCase {
	return &SettlementService{
		repo:      repo,
		splitRepo: splitRepo,
	}
}

func (s *SettlementService) CreateSettlement(ctx context.Context, userId uuid.UUID, input Dtos.CreateSettlementInput) (*Dtos.SettlementResult, error) {
	if input.IdempotencyKey != "" {
		existing, err := s.repo.GetSettlementByIdempotencyKey(ctx, input.IdempotencyKey)
		if err == nil && existing != nil {
			Logger.Debug().
				Str("operation", "CreateSettlement").
				Str("idempotencyKey", input.IdempotencyKey).
				Str("settlementId", existing.Id.String()).
				Msg("Returning existing settlement for idempotency key")
			return s.settlementToDto(existing, false), nil
		}
	}

	splitUUID, err := Helpers.ParseUUID(input.SplitID)
	if err != nil {
		return nil, err
	}

	payeeUUID, err := Helpers.ParseUUID(input.PayeeID)
	if err != nil {
		return nil, err
	}

	if !Domain.IsValidCurrency(input.Currency) {
		return nil, fiber.NewError(fiber.StatusBadRequest, Errors.ErrInvalidCurrency)
	}

	if input.Amount <= 0 {
		return nil, fiber.NewError(fiber.StatusBadRequest, Errors.ErrInvalidSettlementAmount)
	}

	split, splitErr := s.splitRepo.GetSplitById(ctx, splitUUID)
	if splitErr != nil {
		return nil, splitErr
	}

	if Domain.Currency(input.Currency) != split.Currency {
		return nil, fiber.NewError(fiber.StatusBadRequest, Errors.ErrCurrencyMismatch)
	}

	_, participantErr := s.splitRepo.GetParticipant(ctx, splitUUID, userId)
	if participantErr != nil {
		return nil, participantErr
	}

	_, payeeParticipantErr := s.splitRepo.GetParticipant(ctx, splitUUID, payeeUUID)
	if payeeParticipantErr != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, Errors.ErrPayeeNotParticipant)
	}

	var idempotencyKeyPtr *string
	if input.IdempotencyKey != "" {
		idempotencyKeyPtr = &input.IdempotencyKey
	}

	settlement := &Domain.Settlement{
		SplitID:        splitUUID,
		PayerID:        userId,
		PayeeID:        payeeUUID,
		Amount:         input.Amount,
		Currency:       Domain.Currency(input.Currency),
		Date:           time.Now(),
		IdempotencyKey: idempotencyKeyPtr,
	}

	created, dbErr := s.repo.CreateSettlement(ctx, settlement)
	if dbErr != nil {
		return nil, dbErr
	}

	Logger.Debug().
		Str("operation", "CreateSettlement").
		Str("payerId", userId.String()).
		Str("payeeId", payeeUUID.String()).
		Str("settlementId", created.Id.String()).
		Int64("amount", input.Amount).
		Str("currency", input.Currency).
		Msg("Settlement created successfully")

	return s.settlementToDto(created, false), nil
}

func (s *SettlementService) GetPendingSettlements(ctx context.Context, userId uuid.UUID, pagination Helpers.PaginationParams) (*Dtos.SettlementListResult, error) {
	offset := pagination.Offset()
	settlements, total, dbErr := s.repo.GetPendingSettlementsByUserId(ctx, userId, pagination.PageSize, offset)
	if dbErr != nil {
		return nil, dbErr
	}

	results := make([]Dtos.SettlementResult, len(settlements))
	for i, settlement := range settlements {
		results[i] = *s.settlementToDto(&settlement, false)
	}

	return &Dtos.SettlementListResult{
		Settlements: results,
		Page:        pagination.Page,
		PageSize:    pagination.PageSize,
		TotalItems:  total,
	}, nil
}

func (s *SettlementService) GetSettlementHistory(ctx context.Context, userId uuid.UUID, pagination Helpers.PaginationParams) (*Dtos.SettlementListResult, error) {
	offset := pagination.Offset()
	settlements, confirmedMap, total, dbErr := s.repo.GetSettlementHistoryWithConfirmation(ctx, userId, pagination.PageSize, offset)
	if dbErr != nil {
		return nil, dbErr
	}

	results := make([]Dtos.SettlementResult, len(settlements))
	for i, settlement := range settlements {
		isConfirmed := confirmedMap[settlement.Id]
		results[i] = *s.settlementToDto(&settlement, isConfirmed)
	}

	return &Dtos.SettlementListResult{
		Settlements: results,
		Page:        pagination.Page,
		PageSize:    pagination.PageSize,
		TotalItems:  total,
	}, nil
}

func (s *SettlementService) ConfirmSettlement(ctx context.Context, userId, settlementId uuid.UUID) error {
	settlement, dbErr := s.repo.GetSettlementById(ctx, settlementId)
	if dbErr != nil {
		return dbErr
	}

	if settlement.PayeeID != userId {
		return fiber.NewError(fiber.StatusNotFound, Errors.ErrSettlementNotFound)
	}

	isConfirmed, _ := s.repo.IsSettlementConfirmed(ctx, settlementId)
	if isConfirmed {
		return fiber.NewError(fiber.StatusBadRequest, Errors.ErrSettlementAlreadyConfirmed)
	}

	err := s.repo.ConfirmSettlement(ctx, settlementId)
	if err != nil {
		return err
	}

	Logger.Debug().
		Str("operation", "ConfirmSettlement").
		Str("userId", userId.String()).
		Str("settlementId", settlementId.String()).
		Msg("Settlement confirmed successfully")

	return nil
}

func (s *SettlementService) DeleteSettlement(ctx context.Context, userId, settlementId uuid.UUID) error {
	settlement, dbErr := s.repo.GetSettlementById(ctx, settlementId)
	if dbErr != nil {
		return dbErr
	}

	if settlement.PayerID != userId {
		return fiber.NewError(fiber.StatusForbidden, Errors.ErrNotSettlementPayer)
	}

	isConfirmed, _ := s.repo.IsSettlementConfirmed(ctx, settlementId)
	if isConfirmed {
		return fiber.NewError(fiber.StatusBadRequest, Errors.ErrCannotDeleteConfirmedSettlement)
	}

	return s.repo.DeleteSettlement(ctx, settlementId)
}

func (s *SettlementService) settlementToDto(settlement *Domain.Settlement, confirmed bool) *Dtos.SettlementResult {
	payerName := ""
	if settlement.Payer.Id != (uuid.UUID{}) {
		payerName = settlement.Payer.Name
	}
	payeeName := ""
	if settlement.Payee.Id != (uuid.UUID{}) {
		payeeName = settlement.Payee.Name
	}

	return &Dtos.SettlementResult{
		ID:        settlement.Id.String(),
		SplitID:   settlement.SplitID.String(),
		PayerID:   settlement.PayerID.String(),
		PayerName: payerName,
		PayeeID:   settlement.PayeeID.String(),
		PayeeName: payeeName,
		Amount:    settlement.Amount,
		Currency:  string(settlement.Currency),
		Date:      settlement.Date,
		Confirmed: confirmed,
	}
}
