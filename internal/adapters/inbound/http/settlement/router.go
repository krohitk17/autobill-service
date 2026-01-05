package SettlementAdapter

import (
	Middlewares "autobill-service/internal/adapters/inbound/http/middleware"
	JWTUtil "autobill-service/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

type SettlementRouter struct {
	App     *fiber.App
	handler SettlementHandler
	util    JWTUtil.JWTUtil
}

func CreateSettlementRouter(app *fiber.App, handler SettlementHandler, util JWTUtil.JWTUtil) SettlementRouter {
	return SettlementRouter{
		App:     app,
		handler: handler,
		util:    util,
	}
}

func (r SettlementRouter) RegisterRoutes() {
	r.App.Use(Middlewares.AuthMiddleware(r.util))

	r.App.Post("/", r.handler.CreateSettlementHandler).Name("createSettlement")
	r.App.Get("/pending", r.handler.GetPendingSettlementsHandler).Name("getPendingSettlements")
	r.App.Get("/history", r.handler.GetSettlementHistoryHandler).Name("getSettlementHistory")
	r.App.Post("/:settlementId/confirm", r.handler.ConfirmSettlementHandler).Name("confirmSettlement")
	r.App.Delete("/:settlementId", r.handler.DeleteSettlementHandler).Name("deleteSettlement")
}
