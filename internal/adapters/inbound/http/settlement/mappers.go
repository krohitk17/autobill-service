package SettlementAdapter

import (
	AdapterDtos "autobill-service/internal/adapters/inbound/http/settlement/dtos"
	ServiceDtos "autobill-service/internal/application/settlement/dtos"
	Helpers "autobill-service/pkg/helpers"
)

func ToCreateSettlementInput(dto *AdapterDtos.CreateSettlementRequestDto) ServiceDtos.CreateSettlementInput {
	return ServiceDtos.CreateSettlementInput{
		SplitID:        dto.SplitID,
		PayeeID:        dto.PayeeID,
		Amount:         dto.Amount,
		Currency:       dto.Currency,
		IdempotencyKey: dto.IdempotencyKey,
	}
}

func ToSettlementResponseDto(result *ServiceDtos.SettlementResult) AdapterDtos.SettlementResponseDto {
	return AdapterDtos.SettlementResponseDto{
		ID:        result.ID,
		SplitID:   result.SplitID,
		PayerID:   result.PayerID,
		PayerName: result.PayerName,
		PayeeID:   result.PayeeID,
		PayeeName: result.PayeeName,
		Amount:    result.Amount,
		Currency:  result.Currency,
		Date:      result.Date,
		Confirmed: result.Confirmed,
	}
}

func ToSettlementResponseDtoList(results []ServiceDtos.SettlementResult) []AdapterDtos.SettlementResponseDto {
	settlements := make([]AdapterDtos.SettlementResponseDto, len(results))
	for i, s := range results {
		settlements[i] = ToSettlementResponseDto(&s)
	}
	return settlements
}

func ToSettlementListResponseDto(result *ServiceDtos.SettlementListResult) AdapterDtos.SettlementListResponseDto {
	return AdapterDtos.SettlementListResponseDto{
		Settlements: ToSettlementResponseDtoList(result.Settlements),
		Page:        result.Page,
		PageSize:    result.PageSize,
		TotalItems:  result.TotalItems,
		TotalPages:  Helpers.CalculateTotalPages(result.PageSize, result.TotalItems),
	}
}
