package apps

import (
	UserAdapter "autobill-service/internal/adapters/inbound/http/user"
	RepositoryAdapters "autobill-service/internal/adapters/outbound/db"
	UserApp "autobill-service/internal/application/user"
	DB "autobill-service/internal/infrastructure/db"
	JWTUtil "autobill-service/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

func CreateUserApp(util JWTUtil.JWTUtil, db DB.PostgresDB) *UserAdapter.UserRouter {
	userAppFiber := fiber.New(fiber.Config{
		AppName: "autobill-user-service",
	})

	userRepo := RepositoryAdapters.CreateUserRepository(db)

	userService := UserApp.CreateUserService(userRepo)

	userHandler := UserAdapter.CreateUserHandler(userService)

	router := UserAdapter.CreateUserRouter(userAppFiber, userHandler, util)
	router.RegisterRoutes()

	return &router
}
