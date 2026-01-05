package AuthAdapter

import (
	AdapterDtos "autobill-service/internal/adapters/inbound/http/auth/dtos"
	ServiceDtos "autobill-service/internal/application/auth/dtos"
)

func ToRegisterUserInput(dto *AdapterDtos.RegisterUserRequestDto) ServiceDtos.RegisterUserInput {
	return ServiceDtos.RegisterUserInput{
		Email:    dto.Email,
		Name:     dto.Name,
		Password: dto.Password,
	}
}

func ToLoginInput(dto *AdapterDtos.FindUserRequestDto) ServiceDtos.LoginInput {
	return ServiceDtos.LoginInput{
		Email:    dto.Email,
		Password: dto.Password,
	}
}

func ToRefreshTokenInput(dto *AdapterDtos.RefreshTokenRequestDto) ServiceDtos.RefreshTokenInput {
	return ServiceDtos.RefreshTokenInput{
		RefreshToken: dto.RefreshToken,
	}
}

func ToUserLoginResponseDto(result *ServiceDtos.AuthResult) AdapterDtos.UserLoginResponseDto {
	return AdapterDtos.UserLoginResponseDto{
		Id:           result.ID,
		Token:        result.Token,
		RefreshToken: result.RefreshToken,
	}
}
