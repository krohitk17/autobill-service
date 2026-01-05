package UserAdapter

import (
	AdapterDtos "autobill-service/internal/adapters/inbound/http/user/dtos"
	ServiceDtos "autobill-service/internal/application/user/dtos"
)

func ToUpdateUserInput(dto *AdapterDtos.UpdateUserRequestDto) ServiceDtos.UpdateUserInput {
	return ServiceDtos.UpdateUserInput{
		Name:  dto.Name,
		Email: dto.Email,
	}
}

func ToUserResponseDto(result *ServiceDtos.UserResult) AdapterDtos.UserResponseDto {
	return AdapterDtos.UserResponseDto{
		Id:        result.ID,
		Name:      result.Name,
		Email:     result.Email,
		CreatedAt: result.CreatedAt,
	}
}

func ToUpdateUserResponseDto(result *ServiceDtos.UserResult) AdapterDtos.UpdateUserResponseDto {
	return AdapterDtos.UpdateUserResponseDto{
		Id:        result.ID,
		Name:      result.Name,
		Email:     result.Email,
		CreatedAt: result.CreatedAt,
		UpdatedAt: result.UpdatedAt,
	}
}
