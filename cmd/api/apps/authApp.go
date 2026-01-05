package apps

import (
	AuthAdapter "autobill-service/internal/adapters/inbound/http/auth"
	RepositoryAdapters "autobill-service/internal/adapters/outbound/db"
	AuthApp "autobill-service/internal/application/auth"
	DB "autobill-service/internal/infrastructure/db"
	JWTUtil "autobill-service/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

func CreateAuthApp(util JWTUtil.JWTUtil, db DB.PostgresDB) AuthAdapter.AuthRouter {
	authAppFiber := fiber.New(fiber.Config{
		AppName: "autobill-auth-service",
	})

	authRepo := RepositoryAdapters.CreateAuthRepository(db)

	authService := AuthApp.CreateAuthService(authRepo, util)

	authHandler := AuthAdapter.CreateAuthHandler(authService)

	router := AuthAdapter.CreateAuthRouter(authAppFiber, authHandler, util)
	router.RegisterRoutes()

	return router
}
