package SplitAdapter

import (
	Middlewares "autobill-service/internal/adapters/inbound/http/middleware"
	JWTUtil "autobill-service/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

type SplitRouter struct {
	App     *fiber.App
	handler SplitHandler
	util    JWTUtil.JWTUtil
}

func CreateSplitRouter(app *fiber.App, handler SplitHandler, util JWTUtil.JWTUtil) SplitRouter {
	return SplitRouter{
		App:     app,
		handler: handler,
		util:    util,
	}
}

func (r SplitRouter) RegisterRoutes() {
	r.App.Use(Middlewares.AuthMiddleware(r.util))

	r.App.Post("/", r.handler.CreateSplitHandler).Name("createSplit")
	r.App.Get("/me", r.handler.GetMySplitsHandler).Name("getMySplits")
	r.App.Get("/:splitId", r.handler.GetSplitHandler).Name("getSplit")
	r.App.Delete("/:splitId", r.handler.DeleteSplitHandler).Name("deleteSplit")
	r.App.Get("/groups/:groupId", r.handler.GetGroupSplitsHandler).Name("getGroupSplits")

	r.App.Post("/:splitId/participants", r.handler.AddParticipantHandler).Name("addParticipant")
	r.App.Patch("/:splitId/participants/:userId", r.handler.UpdateParticipantHandler).Name("updateParticipant")

	r.App.Post("/:splitId/finalize", r.handler.FinalizeSplitHandler).Name("finalizeSplit")
	r.App.Post("/:splitId/reverse", r.handler.ReverseSplitHandler).Name("reverseSplit")
}
