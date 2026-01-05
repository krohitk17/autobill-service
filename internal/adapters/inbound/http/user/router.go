package UserAdapter

import (
	Middlewares "autobill-service/internal/adapters/inbound/http/middleware"
	JWTUtil "autobill-service/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

type UserRouter struct {
	App     *fiber.App
	handler UserHandler
	util    JWTUtil.JWTUtil
}

func CreateUserRouter(app *fiber.App, handler UserHandler, util JWTUtil.JWTUtil) UserRouter {
	return UserRouter{
		App:     app,
		handler: handler,
		util:    util,
	}
}

func (ur UserRouter) RegisterRoutes() {
	ur.App.Use(Middlewares.AuthMiddleware(ur.util))
	ur.App.Get("/", ur.handler.GetUserHandler).Name("getUser")
	ur.App.Post("/search", ur.handler.FindUserByEmailHandler).Name("findUserByEmail")
	ur.App.Put("/update", ur.handler.UpdateUserHandler).Name("updateUser")
}
