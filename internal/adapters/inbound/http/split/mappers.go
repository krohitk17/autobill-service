package SplitAdapter

import (
	AdapterDtos "autobill-service/internal/adapters/inbound/http/split/dtos"
	ServiceDtos "autobill-service/internal/application/split/dtos"
	Helpers "autobill-service/pkg/helpers"
)

func ToParticipantInputList(dtos []AdapterDtos.ParticipantInput) []ServiceDtos.ParticipantInput {
	participants := make([]ServiceDtos.ParticipantInput, len(dtos))
	for i, p := range dtos {
		participants[i] = ServiceDtos.ParticipantInput{
			UserID:      p.UserID,
			ShareAmount: p.ShareAmount,
		}
	}
	return participants
}

func ToCreateSplitInput(dto *AdapterDtos.CreateSplitRequestDto) ServiceDtos.CreateSplitInput {
	return ServiceDtos.CreateSplitInput{
		Type:          dto.Type,
		DivisionType:  dto.DivisionType,
		TotalAmount:   dto.TotalAmount,
		Currency:      dto.Currency,
		Description:   dto.Description,
		GroupID:       dto.GroupID,
		SimplifyDebts: dto.SimplifyDebts,
		Participants:  ToParticipantInputList(dto.Participants),
	}
}

func ToAddParticipantInput(dto *AdapterDtos.AddParticipantRequestDto) ServiceDtos.AddParticipantInput {
	return ServiceDtos.AddParticipantInput{
		UserID:      dto.UserID,
		ShareAmount: dto.ShareAmount,
	}
}

func ToUpdateParticipantInput(dto *AdapterDtos.UpdateParticipantRequestDto) ServiceDtos.UpdateParticipantInput {
	return ServiceDtos.UpdateParticipantInput{
		ShareAmount: dto.ShareAmount,
		IsSettled:   dto.IsSettled,
	}
}

func ToParticipantResponseDto(result *ServiceDtos.ParticipantResult) AdapterDtos.ParticipantResponseDto {
	return AdapterDtos.ParticipantResponseDto{
		UserID:      result.UserID,
		UserName:    result.UserName,
		ShareAmount: result.ShareAmount,
		Currency:    result.Currency,
		IsSettled:   result.IsSettled,
	}
}

func ToParticipantResponseDtoList(results []ServiceDtos.ParticipantResult) []AdapterDtos.ParticipantResponseDto {
	participants := make([]AdapterDtos.ParticipantResponseDto, len(results))
	for i, p := range results {
		participants[i] = ToParticipantResponseDto(&p)
	}
	return participants
}

func ToSplitResponseDto(result *ServiceDtos.SplitResult) AdapterDtos.SplitResponseDto {
	return AdapterDtos.SplitResponseDto{
		ID:            result.ID,
		Type:          result.Type,
		DivisionType:  result.DivisionType,
		TotalAmount:   result.TotalAmount,
		Currency:      result.Currency,
		Description:   result.Description,
		GroupID:       result.GroupID,
		CreatedByID:   result.CreatedByID,
		CreatedAt:     result.CreatedAt,
		IsFinalized:   result.IsFinalized,
		SimplifyDebts: result.SimplifyDebts,
		Participants:  ToParticipantResponseDtoList(result.Participants),
	}
}

func ToSplitResponseDtoList(results []ServiceDtos.SplitResult) []AdapterDtos.SplitResponseDto {
	splits := make([]AdapterDtos.SplitResponseDto, len(results))
	for i, s := range results {
		splits[i] = ToSplitResponseDto(&s)
	}
	return splits
}

func ToSplitListResponseDto(result *ServiceDtos.SplitListResult) AdapterDtos.SplitListResponseDto {
	return AdapterDtos.SplitListResponseDto{
		Splits:     ToSplitResponseDtoList(result.Splits),
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalItems: result.TotalItems,
		TotalPages: Helpers.CalculateTotalPages(result.PageSize, result.TotalItems),
	}
}
