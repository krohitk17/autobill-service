package SocialApplication

import (
	"context"

	"github.com/gofiber/fiber/v2"

	Dtos "autobill-service/internal/application/social/dtos"
	Domain "autobill-service/internal/domain"
	HttpPorts "autobill-service/internal/ports/inbound/http"
	RepositoryPorts "autobill-service/internal/ports/outbound/db"
	Errors "autobill-service/pkg/errors"
	Helpers "autobill-service/pkg/helpers"
	Logger "autobill-service/pkg/logger"

	"github.com/google/uuid"
)

type SocialService struct {
	db RepositoryPorts.SocialRepositoryPort
}

func CreateSocialService(db RepositoryPorts.SocialRepositoryPort) HttpPorts.SocialUseCase {
	return &SocialService{db: db}
}

func (s *SocialService) GetFriendRequestsList(ctx context.Context, userID uuid.UUID, requestType Dtos.RequestType, pagination Helpers.PaginationParams) (*Dtos.FriendRequestListResult, error) {
	var repoRequestType RepositoryPorts.FriendRequestType
	switch requestType {
	case Dtos.RequestTypeReceived:
		repoRequestType = RepositoryPorts.FriendRequestReceived
	case Dtos.RequestTypeSent:
		repoRequestType = RepositoryPorts.FriendRequestSent
	}

	offset := pagination.Offset()
	requests, total, err := s.db.GetFriendRequestsList(ctx, userID, repoRequestType, pagination.PageSize, offset)
	if err != nil {
		return nil, err
	}

	requestResults := make([]Dtos.FriendRequestResult, 0, len(requests))
	for _, req := range requests {
		var user Domain.User
		switch requestType {
		case Dtos.RequestTypeReceived:
			user = req.Sender
		case Dtos.RequestTypeSent:
			user = req.Receiver
		}
		requestResults = append(requestResults, Dtos.FriendRequestResult{
			RequestId: req.Id.String(),
			UserId:    user.Id.String(),
			Name:      user.Name,
			Email:     user.Email,
			Status:    req.Status,
		})
	}

	return &Dtos.FriendRequestListResult{
		Type:       requestType,
		Requests:   requestResults,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalItems: total,
	}, nil
}

func (s *SocialService) SendFriendRequest(ctx context.Context, senderId, receiverId uuid.UUID) (*Dtos.FriendRequestResult, error) {
	if senderId == receiverId {
		return nil, fiber.NewError(fiber.StatusBadRequest, Errors.ErrCannotSendRequestToSelf)
	}

	existingRequest, err := s.db.CheckExistingRequest(ctx, senderId, receiverId)
	if err != nil {
		return nil, err
	}
	if existingRequest {
		return nil, fiber.NewError(fiber.StatusBadRequest, Errors.ErrFriendRequestExists)
	}

	alreadyFriends, err := s.db.CheckFriendship(ctx, senderId, receiverId)
	if err != nil {
		return nil, err
	}
	if alreadyFriends {
		return nil, fiber.NewError(fiber.StatusBadRequest, Errors.ErrAlreadyFriends)
	}

	request, err := s.db.CreateFriendRequest(ctx, senderId, receiverId)
	if err != nil {
		return nil, err
	}

	Logger.Debug().
		Str("operation", "SendFriendRequest").
		Str("senderId", senderId.String()).
		Str("receiverId", receiverId.String()).
		Str("requestId", request.Id.String()).
		Msg("Friend request sent successfully")

	return &Dtos.FriendRequestResult{RequestId: request.Id.String()}, nil
}

func (s *SocialService) AcceptFriendRequest(ctx context.Context, senderId, requestId uuid.UUID) error {
	return s.db.AcceptFriendRequest(ctx, senderId, requestId)
}

func (s *SocialService) RejectFriendRequest(ctx context.Context, senderId, requestId uuid.UUID) error {
	return s.db.RejectFriendRequest(ctx, senderId, requestId)
}

func (s *SocialService) GetFriendsList(ctx context.Context, userID uuid.UUID, pagination Helpers.PaginationParams) (*Dtos.FriendsListResult, error) {
	offset := pagination.Offset()
	friends, total, err := s.db.GetFriendsList(ctx, userID, pagination.PageSize, offset)
	if err != nil {
		return nil, err
	}

	friendResults := make([]Dtos.FriendResult, 0, len(friends))
	for _, friend := range friends {
		friendResults = append(friendResults, Dtos.FriendResult{
			ID:     friend.Id.String(),
			Name:   friend.Name,
			Email:  friend.Email,
			Status: Domain.FriendAccepted,
		})
	}

	return &Dtos.FriendsListResult{
		UserID:     userID.String(),
		Friends:    friendResults,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalItems: total,
	}, nil
}

func (s *SocialService) RemoveFriend(ctx context.Context, userId, friendId uuid.UUID) error {
	return s.db.RemoveFriend(ctx, userId, friendId)
}
