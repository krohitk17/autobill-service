package GroupAdapter

import (
	AdapterDtos "autobill-service/internal/adapters/inbound/http/group/dtos"
	ServiceDtos "autobill-service/internal/application/group/dtos"
	Helpers "autobill-service/pkg/helpers"
)

func ToCreateGroupInput(dto *AdapterDtos.CreateGroupRequestDto) ServiceDtos.CreateGroupInput {
	return ServiceDtos.CreateGroupInput{
		Name:          dto.Name,
		SimplifyDebts: dto.SimplifyDebts,
	}
}

func ToUpdateGroupInput(dto *AdapterDtos.UpdateGroupRequestDto) ServiceDtos.UpdateGroupInput {
	return ServiceDtos.UpdateGroupInput{
		Name:          dto.Name,
		SimplifyDebts: dto.SimplifyDebts,
	}
}

func ToAddMemberInput(dto *AdapterDtos.AddMemberRequestDto) ServiceDtos.AddMemberInput {
	return ServiceDtos.AddMemberInput{
		UserID: dto.UserID,
		Role:   dto.Role,
	}
}

func ToGroupResponseDto(result *ServiceDtos.GroupResult) AdapterDtos.GroupResponseDto {
	return AdapterDtos.GroupResponseDto{
		ID:            result.ID,
		Name:          result.Name,
		SimplifyDebts: result.SimplifyDebts,
		CreatedAt:     result.CreatedAt,
	}
}

func ToMemberResponseDto(result *ServiceDtos.MemberResult) AdapterDtos.MemberResponseDto {
	return AdapterDtos.MemberResponseDto{
		UserID: result.UserID,
		Name:   result.Name,
		Email:  result.Email,
		Role:   result.Role,
	}
}

func ToMemberResponseDtoList(results []ServiceDtos.MemberResult) []AdapterDtos.MemberResponseDto {
	members := make([]AdapterDtos.MemberResponseDto, len(results))
	for i, m := range results {
		members[i] = ToMemberResponseDto(&m)
	}
	return members
}

func ToGroupDetailResponseDto(result *ServiceDtos.GroupDetailResult) AdapterDtos.GroupDetailResponseDto {
	return AdapterDtos.GroupDetailResponseDto{
		ID:            result.ID,
		Name:          result.Name,
		SimplifyDebts: result.SimplifyDebts,
		CreatedAt:     result.CreatedAt,
		Members:       ToMemberResponseDtoList(result.Members),
	}
}

func ToGroupResponseDtoList(results []ServiceDtos.GroupResult) []AdapterDtos.GroupResponseDto {
	groups := make([]AdapterDtos.GroupResponseDto, len(results))
	for i, g := range results {
		groups[i] = ToGroupResponseDto(&g)
	}
	return groups
}

func ToGroupListResponseDto(result *ServiceDtos.GroupListResult) AdapterDtos.GroupListResponseDto {
	return AdapterDtos.GroupListResponseDto{
		Groups:     ToGroupResponseDtoList(result.Groups),
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalItems: result.TotalItems,
		TotalPages: Helpers.CalculateTotalPages(result.PageSize, result.TotalItems),
	}
}
