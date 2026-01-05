package SocialAdapter

import (
	Middlewares "autobill-service/internal/adapters/inbound/http/middleware"
	SocialDtos "autobill-service/internal/adapters/inbound/http/social/dtos"
	HttpPorts "autobill-service/internal/ports/inbound/http"
	Errors "autobill-service/pkg/errors"
	Helpers "autobill-service/pkg/helpers"

	"github.com/gofiber/fiber/v2"
)

type SocialHandler struct {
	service HttpPorts.SocialUseCase
}

func CreateSocialHandler(service HttpPorts.SocialUseCase) SocialHandler {
	return SocialHandler{service: service}
}

func (h *SocialHandler) GetFriendRequestsListHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	userId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}

	requestType, err := ToRequestType(c.Query("type"))
	if err != nil {
		return err
	}
	pagination := Helpers.ParsePagination(c)
	result, err := h.service.GetFriendRequestsList(ctx, userId, requestType, pagination)
	if err != nil {
		return err
	}

	return c.JSON(ToGetFriendRequestsListResponseDto(result))
}

func (h *SocialHandler) SendFriendRequestHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	senderId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}

	reqBody := new(SocialDtos.SendFriendRequestRequestDto)
	if err := c.BodyParser(reqBody); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Errors.ErrInvalidRequestBody)
	}

	if err := Helpers.ValidateRequest(reqBody); err != nil {
		return err
	}

	result, err := h.service.SendFriendRequest(ctx, senderId, reqBody.ReceiverId)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(ToCreateFriendRequestResponseDto(result))
}

func (h *SocialHandler) AcceptFriendRequestHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	receiverId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}

	requestId, err := Helpers.ParseUUID(c.Params("requestId"))
	if err != nil {
		return err
	}

	err = h.service.AcceptFriendRequest(ctx, receiverId, requestId)
	if err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusCreated)
}

func (h *SocialHandler) RejectFriendRequestHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	receiverId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}

	requestId, err := Helpers.ParseUUID(c.Params("requestId"))
	if err != nil {
		return err
	}

	err = h.service.RejectFriendRequest(ctx, receiverId, requestId)
	if err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *SocialHandler) GetFriendsListHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	userId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}

	pagination := Helpers.ParsePagination(c)
	result, err := h.service.GetFriendsList(ctx, userId, pagination)
	if err != nil {
		return err
	}

	return c.JSON(ToGetFriendsListResponseDto(result))
}

func (h *SocialHandler) RemoveFriendHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	userId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}

	friendId, err := Helpers.ParseUUID(c.Params("friendId"))
	if err != nil {
		return err
	}

	err = h.service.RemoveFriend(ctx, userId, friendId)
	if err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}
