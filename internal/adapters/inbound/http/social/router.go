package SocialAdapter

import (
	Middlewares "autobill-service/internal/adapters/inbound/http/middleware"
	JWTUtil "autobill-service/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

type SocialRouter struct {
	App     *fiber.App
	handler SocialHandler
	util    JWTUtil.JWTUtil
}

func CreateSocialRouter(app *fiber.App, handler SocialHandler, util JWTUtil.JWTUtil) SocialRouter {
	return SocialRouter{
		App:     app,
		handler: handler,
		util:    util,
	}
}

func (r *SocialRouter) RegisterRoutes() {
	r.App.Use(Middlewares.AuthMiddleware(r.util))

	r.App.Get("/requests", r.handler.GetFriendRequestsListHandler).Name("getFriendRequests")
	r.App.Post("/requests", r.handler.SendFriendRequestHandler).Name("sendFriendRequest")
	r.App.Post("/requests/:requestId/accept", r.handler.AcceptFriendRequestHandler).Name("acceptFriendRequest")
	r.App.Delete("/requests/:requestId", r.handler.RejectFriendRequestHandler).Name("rejectFriendRequest")

	r.App.Get("/friends", r.handler.GetFriendsListHandler).Name("getFriends")
	r.App.Delete("/friends/:friendId", r.handler.RemoveFriendHandler).Name("removeFriend")
}
