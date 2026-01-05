package BalanceAdapter

import (
	Middlewares "autobill-service/internal/adapters/inbound/http/middleware"
	JWTUtil "autobill-service/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

type BalanceRouter struct {
	App     *fiber.App
	handler BalanceHandler
	util    JWTUtil.JWTUtil
}

func CreateBalanceRouter(app *fiber.App, handler BalanceHandler, util JWTUtil.JWTUtil) BalanceRouter {
	return BalanceRouter{
		App:     app,
		handler: handler,
		util:    util,
	}
}

func (r BalanceRouter) RegisterRoutes() {
	r.App.Use(Middlewares.AuthMiddleware(r.util))

	r.App.Get("/me", r.handler.GetMyBalanceHandler).Name("getMyBalance")
	r.App.Get("/users/:userId", r.handler.GetUserBalanceHandler).Name("getUserBalance")

	r.App.Get("/groups/:groupId", r.handler.GetGroupBalanceHandler).Name("getGroupBalance")
	r.App.Post("/groups/:groupId/recalculate", r.handler.RecalculateGroupBalanceHandler).Name("recalculateGroupBalance")
	r.App.Get("/groups/:groupId/simplify", r.handler.GetSimplifiedDebtsHandler).Name("getSimplifiedDebts")
}
