package GroupAdapter

import (
	GroupDtos "autobill-service/internal/adapters/inbound/http/group/dtos"
	Middlewares "autobill-service/internal/adapters/inbound/http/middleware"
	HttpPorts "autobill-service/internal/ports/inbound/http"
	Errors "autobill-service/pkg/errors"
	Helpers "autobill-service/pkg/helpers"

	"github.com/gofiber/fiber/v2"
)

type GroupHandler struct {
	service HttpPorts.GroupUseCase
}

func CreateGroupHandler(service HttpPorts.GroupUseCase) GroupHandler {
	return GroupHandler{service: service}
}

func (h *GroupHandler) CreateGroupHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	userId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}
	reqBody := new(GroupDtos.CreateGroupRequestDto)

	if err := c.BodyParser(reqBody); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Errors.ErrInvalidRequestBody)
	}

	if err := Helpers.ValidateRequest(reqBody); err != nil {
		return err
	}

	result, err := h.service.CreateGroup(ctx, userId, ToCreateGroupInput(reqBody))
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(ToGroupResponseDto(result))
}

func (h *GroupHandler) GetGroupsHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	userId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}

	pagination := Helpers.ParsePagination(c)
	result, err := h.service.GetGroups(ctx, userId, pagination)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(ToGroupListResponseDto(result))
}

func (h *GroupHandler) GetGroupHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	userId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}
	groupId, err := Helpers.ParseUUID(c.Params("groupId"))
	if err != nil {
		return err
	}

	result, err := h.service.GetGroup(ctx, userId, groupId)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(ToGroupDetailResponseDto(result))
}

func (h *GroupHandler) UpdateGroupHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	userId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}
	groupId, err := Helpers.ParseUUID(c.Params("groupId"))
	if err != nil {
		return err
	}
	reqBody := new(GroupDtos.UpdateGroupRequestDto)

	if err := c.BodyParser(reqBody); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Errors.ErrInvalidRequestBody)
	}

	if err := Helpers.ValidateRequest(reqBody); err != nil {
		return err
	}

	result, err := h.service.UpdateGroup(ctx, userId, groupId, ToUpdateGroupInput(reqBody))
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(ToGroupResponseDto(result))
}

func (h *GroupHandler) DeleteGroupHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	userId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}
	groupId, err := Helpers.ParseUUID(c.Params("groupId"))
	if err != nil {
		return err
	}

	err = h.service.DeleteGroup(ctx, userId, groupId)
	if err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *GroupHandler) AddMemberHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	userId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}
	groupId, err := Helpers.ParseUUID(c.Params("groupId"))
	if err != nil {
		return err
	}
	reqBody := new(GroupDtos.AddMemberRequestDto)

	if err := c.BodyParser(reqBody); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Errors.ErrInvalidRequestBody)
	}

	if err := Helpers.ValidateRequest(reqBody); err != nil {
		return err
	}

	result, err := h.service.AddMember(ctx, userId, groupId, ToAddMemberInput(reqBody))
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(ToMemberResponseDto(result))
}

func (h *GroupHandler) UpdateMemberRoleHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	userId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}
	groupId, err := Helpers.ParseUUID(c.Params("groupId"))
	if err != nil {
		return err
	}
	memberId, err := Helpers.ParseUUID(c.Params("userId"))
	if err != nil {
		return err
	}
	reqBody := new(GroupDtos.UpdateMemberRoleRequestDto)

	if err := c.BodyParser(reqBody); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Errors.ErrInvalidRequestBody)
	}

	if err := Helpers.ValidateRequest(reqBody); err != nil {
		return err
	}

	err = h.service.UpdateMemberRole(ctx, userId, groupId, memberId, reqBody.Role)
	if err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *GroupHandler) RemoveMemberHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	userId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}
	groupId, err := Helpers.ParseUUID(c.Params("groupId"))
	if err != nil {
		return err
	}
	memberId, err := Helpers.ParseUUID(c.Params("userId"))
	if err != nil {
		return err
	}

	err = h.service.RemoveMember(ctx, userId, groupId, memberId)
	if err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *GroupHandler) LeaveGroupHandler(c *fiber.Ctx) error {
	ctx := Middlewares.GetContext(c)
	userId, err := Helpers.GetUserIdFromContext(c)
	if err != nil {
		return err
	}
	groupId, err := Helpers.ParseUUID(c.Params("groupId"))
	if err != nil {
		return err
	}

	err = h.service.LeaveGroup(ctx, userId, groupId)
	if err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}
