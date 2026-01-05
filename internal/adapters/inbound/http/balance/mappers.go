package BalanceAdapter

import (
	AdapterDtos "autobill-service/internal/adapters/inbound/http/balance/dtos"
	ServiceDtos "autobill-service/internal/application/balance/dtos"
)

func ToUserBalanceItemDto(result *ServiceDtos.UserBalanceItemResult) AdapterDtos.UserBalanceItemDto {
	return AdapterDtos.UserBalanceItemDto{
		OtherUserID:   result.OtherUserID,
		OtherUserName: result.OtherUserName,
		NetAmount:     result.NetAmount,
		Currency:      result.Currency,
	}
}

func ToUserBalanceItemDtoList(results []ServiceDtos.UserBalanceItemResult) []AdapterDtos.UserBalanceItemDto {
	items := make([]AdapterDtos.UserBalanceItemDto, len(results))
	for i, b := range results {
		items[i] = ToUserBalanceItemDto(&b)
	}
	return items
}

func ToUserBalanceResponseDto(result *ServiceDtos.UserBalanceResult) AdapterDtos.UserBalanceResponseDto {
	return AdapterDtos.UserBalanceResponseDto{
		UserID:   result.UserID,
		Balances: ToUserBalanceItemDtoList(result.Balances),
	}
}

func ToGroupBalanceItemDto(result *ServiceDtos.GroupBalanceItemResult) AdapterDtos.GroupBalanceItemDto {
	return AdapterDtos.GroupBalanceItemDto{
		UserID:    result.UserID,
		UserName:  result.UserName,
		NetAmount: result.NetAmount,
		Currency:  result.Currency,
	}
}

func ToGroupBalanceItemDtoList(results []ServiceDtos.GroupBalanceItemResult) []AdapterDtos.GroupBalanceItemDto {
	items := make([]AdapterDtos.GroupBalanceItemDto, len(results))
	for i, b := range results {
		items[i] = ToGroupBalanceItemDto(&b)
	}
	return items
}

func ToGroupBalanceResponseDto(result *ServiceDtos.GroupBalanceResult) AdapterDtos.GroupBalanceResponseDto {
	return AdapterDtos.GroupBalanceResponseDto{
		GroupID:   result.GroupID,
		GroupName: result.GroupName,
		Balances:  ToGroupBalanceItemDtoList(result.Balances),
	}
}

func ToSimplifiedDebtDto(result *ServiceDtos.SimplifiedDebtResult) AdapterDtos.SimplifiedDebtDto {
	return AdapterDtos.SimplifiedDebtDto{
		FromUserID:   result.FromUserID,
		FromUserName: result.FromUserName,
		ToUserID:     result.ToUserID,
		ToUserName:   result.ToUserName,
		Amount:       result.Amount,
		Currency:     result.Currency,
	}
}

func ToSimplifiedDebtDtoList(results []ServiceDtos.SimplifiedDebtResult) []AdapterDtos.SimplifiedDebtDto {
	items := make([]AdapterDtos.SimplifiedDebtDto, len(results))
	for i, d := range results {
		items[i] = ToSimplifiedDebtDto(&d)
	}
	return items
}

func ToSimplifiedDebtsResponseDto(result *ServiceDtos.SimplifiedDebtsResult) AdapterDtos.SimplifiedDebtsResponseDto {
	return AdapterDtos.SimplifiedDebtsResponseDto{
		GroupID: result.GroupID,
		Debts:   ToSimplifiedDebtDtoList(result.Debts),
	}
}
