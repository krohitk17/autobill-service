package AuthAdapter

import (
	Middlewares "autobill-service/internal/adapters/inbound/http/middleware"
	JWTUtil "autobill-service/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

type AuthRouter struct {
	App     *fiber.App
	handler AuthHandler
	util    JWTUtil.JWTUtil
}

func CreateAuthRouter(app *fiber.App, handler AuthHandler, util JWTUtil.JWTUtil) AuthRouter {
	return AuthRouter{
		App:     app,
		handler: handler,
		util:    util,
	}
}

func (ar AuthRouter) RegisterRoutes() {
	ar.App.Post("/register", ar.handler.RegisterUserHandler).Name("registerUser")
	ar.App.Post("/login", ar.handler.LoginHandler).Name("loginUser")
	ar.App.Post("/reactivate", ar.handler.ReactivateUserHandler).Name("reactivateUser")
	ar.App.Post("/refresh", ar.handler.RefreshTokenHandler).Name("refreshToken")
	ar.App.Post("/logout", ar.handler.LogoutHandler).Name("logoutUser")
	ar.App.Use(Middlewares.AuthMiddleware(ar.util))
	ar.App.Put("/password", ar.handler.UpdatePasswordHandler).Name("updatePassword")
	ar.App.Post("/logout-all", ar.handler.LogoutAllHandler).Name("logoutAllDevices")
	ar.App.Delete("/deactivate", ar.handler.DeactivateUserHandler).Name("deactivateUser")
}
