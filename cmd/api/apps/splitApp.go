package apps

import (
	SplitAdapter "autobill-service/internal/adapters/inbound/http/split"
	RepositoryAdapters "autobill-service/internal/adapters/outbound/db"
	SplitApp "autobill-service/internal/application/split"
	DB "autobill-service/internal/infrastructure/db"
	JWTUtil "autobill-service/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

func CreateSplitApp(util JWTUtil.JWTUtil, db DB.PostgresDB) SplitAdapter.SplitRouter {
	splitAppFiber := fiber.New(fiber.Config{
		AppName: "autobill-split-service",
	})

	splitRepo := RepositoryAdapters.CreateSplitRepository(db)
	groupRepo := RepositoryAdapters.CreateGroupRepository(db)
	balanceRepo := RepositoryAdapters.CreateBalanceRepository(db)

	splitService := SplitApp.CreateSplitService(splitRepo, groupRepo, balanceRepo)

	splitHandler := SplitAdapter.CreateSplitHandler(splitService)

	router := SplitAdapter.CreateSplitRouter(splitAppFiber, splitHandler, util)
	router.RegisterRoutes()

	return router
}
