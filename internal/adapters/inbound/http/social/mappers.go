package SocialAdapter

import (
	AdapterDtos "autobill-service/internal/adapters/inbound/http/social/dtos"
	ServiceDtos "autobill-service/internal/application/social/dtos"
	Errors "autobill-service/pkg/errors"
	Helpers "autobill-service/pkg/helpers"

	"github.com/gofiber/fiber/v2"
)

func ToFriendRequestDto(result *ServiceDtos.FriendRequestResult) AdapterDtos.FriendRequestDto {
	return AdapterDtos.FriendRequestDto{
		RequestId: result.RequestId,
		Id:        result.UserId,
		Name:      result.Name,
		Email:     result.Email,
		Status:    AdapterDtos.FriendStatus(result.Status),
	}
}

func ToFriendRequestDtoList(results []ServiceDtos.FriendRequestResult) []AdapterDtos.FriendRequestDto {
	requests := make([]AdapterDtos.FriendRequestDto, len(results))
	for i, r := range results {
		requests[i] = ToFriendRequestDto(&r)
	}
	return requests
}

func ToGetFriendRequestsListResponseDto(result *ServiceDtos.FriendRequestListResult) AdapterDtos.GetFriendRequestsListResponseDto {
	return AdapterDtos.GetFriendRequestsListResponseDto{
		Type:       AdapterDtos.RequestType(result.Type),
		Users:      ToFriendRequestDtoList(result.Requests),
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalItems: result.TotalItems,
		TotalPages: Helpers.CalculateTotalPages(result.PageSize, result.TotalItems),
	}
}

func ToCreateFriendRequestResponseDto(result *ServiceDtos.FriendRequestResult) AdapterDtos.CreateFriendRequestResponseDto {
	return AdapterDtos.CreateFriendRequestResponseDto{
		RequestId: result.RequestId,
	}
}

func ToUserDto(result *ServiceDtos.FriendResult) AdapterDtos.UserDto {
	return AdapterDtos.UserDto{
		Id:     result.ID,
		Name:   result.Name,
		Email:  result.Email,
		Status: AdapterDtos.FriendStatus(result.Status),
	}
}

func ToUserDtoList(results []ServiceDtos.FriendResult) []AdapterDtos.UserDto {
	users := make([]AdapterDtos.UserDto, len(results))
	for i, f := range results {
		users[i] = ToUserDto(&f)
	}
	return users
}

func ToGetFriendsListResponseDto(result *ServiceDtos.FriendsListResult) AdapterDtos.GetFriendsListResponseDto {
	return AdapterDtos.GetFriendsListResponseDto{
		Id:         result.UserID,
		Friends:    ToUserDtoList(result.Friends),
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalItems: result.TotalItems,
		TotalPages: Helpers.CalculateTotalPages(result.PageSize, result.TotalItems),
	}
}

func ToRequestType(requestTypeStr string) (ServiceDtos.RequestType, error) {
	requestType := ServiceDtos.RequestType(requestTypeStr)
	if requestType == "" {
		return "", fiber.NewError(fiber.StatusBadRequest, Errors.ErrMissingQueryParam)
	}
	if requestType != "sent" && requestType != "received" {
		return "", fiber.NewError(fiber.StatusBadRequest, Errors.ErrInvalidQueryParam)
	}
	return ServiceDtos.RequestType(requestType), nil
}
