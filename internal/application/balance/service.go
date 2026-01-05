package balance

import (
	"context"

	"github.com/gofiber/fiber/v2"

	Dtos "autobill-service/internal/application/balance/dtos"
	Domain "autobill-service/internal/domain"
	RepositoryPorts "autobill-service/internal/ports/outbound/db"
	Errors "autobill-service/pkg/errors"

	"github.com/google/uuid"
)

type BalanceService struct {
	repo      RepositoryPorts.BalanceRepositoryPort
	groupRepo RepositoryPorts.GroupRepositoryPort
}

func CreateBalanceService(repo RepositoryPorts.BalanceRepositoryPort, groupRepo RepositoryPorts.GroupRepositoryPort) *BalanceService {
	return &BalanceService{
		repo:      repo,
		groupRepo: groupRepo,
	}
}

func (s *BalanceService) GetMyBalance(ctx context.Context, userId uuid.UUID) (*Dtos.UserBalanceResult, error) {
	balances, dbErr := s.repo.GetUserBalances(ctx, userId)
	if dbErr != nil {
		return nil, dbErr
	}

	balanceItems := make([]Dtos.UserBalanceItemResult, len(balances))
	for i, b := range balances {
		balanceItems[i] = Dtos.UserBalanceItemResult{
			OtherUserID:   b.OtherUserID.String(),
			OtherUserName: b.OtherUser.Name,
			NetAmount:     b.NetAmount,
			Currency:      string(b.Currency),
		}
	}

	return &Dtos.UserBalanceResult{
		UserID:   userId.String(),
		Balances: balanceItems,
	}, nil
}

func (s *BalanceService) GetGroupBalance(ctx context.Context, userId, groupId uuid.UUID) (*Dtos.GroupBalanceResult, error) {
	_, memberErr := s.groupRepo.GetMembership(ctx, groupId, userId)
	if memberErr != nil {
		return nil, memberErr
	}

	group, groupErr := s.groupRepo.GetGroupById(ctx, groupId)
	if groupErr != nil {
		return nil, groupErr
	}

	balances, dbErr := s.repo.GetGroupBalances(ctx, groupId)
	if dbErr != nil {
		return nil, dbErr
	}

	balanceItems := make([]Dtos.GroupBalanceItemResult, len(balances))
	for i, b := range balances {
		userName := b.User.Name
		balanceItems[i] = Dtos.GroupBalanceItemResult{
			UserID:    b.UserID.String(),
			UserName:  userName,
			NetAmount: b.NetAmount,
			Currency:  string(b.Currency),
		}
	}

	return &Dtos.GroupBalanceResult{
		GroupID:   groupId.String(),
		GroupName: group.Name,
		Balances:  balanceItems,
	}, nil
}

func (s *BalanceService) RecalculateGroupBalance(ctx context.Context, userId, groupId uuid.UUID) (*Dtos.GroupBalanceResult, error) {
	isAdmin, adminErr := s.groupRepo.IsGroupAdmin(ctx, groupId, userId)
	if adminErr != nil {
		return nil, adminErr
	}
	if !isAdmin {
		return nil, fiber.NewError(fiber.StatusNotFound, Errors.ErrGroupNotFound)
	}

	group, groupErr := s.groupRepo.GetGroupById(ctx, groupId)
	if groupErr != nil {
		return nil, groupErr
	}

	splits, splitErr := s.repo.GetFinalizedSplitsWithParticipants(ctx, groupId)
	if splitErr != nil {
		return nil, splitErr
	}

	var splitIDs []uuid.UUID
	for _, split := range splits {
		splitIDs = append(splitIDs, split.Id)
	}

	settlements, settlementErr := s.repo.GetSettlementsForSplits(ctx, splitIDs)
	if settlementErr != nil {
		return nil, settlementErr
	}

	calculatedBalances := s.calculateGroupBalances(ctx, groupId, splits, settlements)

	balances, dbErr := s.repo.ReplaceGroupBalances(ctx, groupId, calculatedBalances)
	if dbErr != nil {
		return nil, dbErr
	}

	balanceItems := make([]Dtos.GroupBalanceItemResult, len(balances))
	for i, b := range balances {
		userName := b.User.Name
		balanceItems[i] = Dtos.GroupBalanceItemResult{
			UserID:    b.UserID.String(),
			UserName:  userName,
			NetAmount: b.NetAmount,
			Currency:  string(b.Currency),
		}
	}

	return &Dtos.GroupBalanceResult{
		GroupID:   groupId.String(),
		GroupName: group.Name,
		Balances:  balanceItems,
	}, nil
}

func (s *BalanceService) calculateGroupBalances(ctx context.Context, groupId uuid.UUID, splits []Domain.Split, settlements []Domain.Settlement) []Domain.GroupBalance {
	settlementMap := make(map[uuid.UUID][]Domain.Settlement)
	for _, settlement := range settlements {
		settlementMap[settlement.SplitID] = append(settlementMap[settlement.SplitID], settlement)
	}

	balanceMap := make(map[uuid.UUID]map[Domain.Currency]int64)

	for _, split := range splits {
		if balanceMap[split.CreatedByID] == nil {
			balanceMap[split.CreatedByID] = make(map[Domain.Currency]int64)
		}

		for _, participant := range split.Participants {
			if participant.IsSettled {
				continue
			}

			if balanceMap[participant.UserID] == nil {
				balanceMap[participant.UserID] = make(map[Domain.Currency]int64)
			}

			if participant.UserID == split.CreatedByID {
				balanceMap[participant.UserID][split.Currency] += (split.TotalAmount - participant.ShareAmount)
			} else {
				balanceMap[participant.UserID][split.Currency] -= participant.ShareAmount
			}
		}
	}

	for _, settlement := range settlements {
		isSettled, _ := s.repo.GetSettledParticipants(ctx, settlement.SplitID, settlement.PayerID)
		if !isSettled {
			continue
		}

		if balanceMap[settlement.PayerID] == nil {
			balanceMap[settlement.PayerID] = make(map[Domain.Currency]int64)
		}
		if balanceMap[settlement.PayeeID] == nil {
			balanceMap[settlement.PayeeID] = make(map[Domain.Currency]int64)
		}

		balanceMap[settlement.PayerID][settlement.Currency] += settlement.Amount
		balanceMap[settlement.PayeeID][settlement.Currency] -= settlement.Amount
	}

	var balances []Domain.GroupBalance
	for userId, currencies := range balanceMap {
		for currency, amount := range currencies {
			if amount == 0 {
				continue
			}
			balances = append(balances, Domain.GroupBalance{
				UserID:    userId,
				GroupID:   groupId,
				NetAmount: amount,
				Currency:  currency,
			})
		}
	}

	return balances
}

func (s *BalanceService) GetSimplifiedDebts(ctx context.Context, userId, groupId uuid.UUID) (*Dtos.SimplifiedDebtsResult, error) {
	_, memberErr := s.groupRepo.GetMembership(ctx, groupId, userId)
	if memberErr != nil {
		return nil, memberErr
	}

	debts, dbErr := s.repo.GetSimplifiedDebts(ctx, groupId)
	if dbErr != nil {
		return nil, dbErr
	}

	debtResults := make([]Dtos.SimplifiedDebtResult, len(debts))
	for i, d := range debts {
		debtResults[i] = Dtos.SimplifiedDebtResult{
			FromUserID:   d.FromUserID.String(),
			FromUserName: d.FromUserName,
			ToUserID:     d.ToUserID.String(),
			ToUserName:   d.ToUserName,
			Amount:       d.Amount,
			Currency:     string(d.Currency),
		}
	}

	return &Dtos.SimplifiedDebtsResult{
		GroupID: groupId.String(),
		Debts:   debtResults,
	}, nil
}
