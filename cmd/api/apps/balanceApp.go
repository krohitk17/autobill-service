package apps

import (
	BalanceAdapter "autobill-service/internal/adapters/inbound/http/balance"
	RepositoryAdapters "autobill-service/internal/adapters/outbound/db"
	BalanceApp "autobill-service/internal/application/balance"
	DB "autobill-service/internal/infrastructure/db"
	JWTUtil "autobill-service/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

func CreateBalanceApp(util JWTUtil.JWTUtil, db DB.PostgresDB) BalanceAdapter.BalanceRouter {
	balanceAppFiber := fiber.New(fiber.Config{
		AppName: "autobill-balance-service",
	})

	balanceRepo := RepositoryAdapters.CreateBalanceRepository(db)
	groupRepo := RepositoryAdapters.CreateGroupRepository(db)

	balanceService := BalanceApp.CreateBalanceService(balanceRepo, groupRepo)

	balanceHandler := BalanceAdapter.CreateBalanceHandler(balanceService)

	router := BalanceAdapter.CreateBalanceRouter(balanceAppFiber, balanceHandler, util)
	router.RegisterRoutes()

	return router
}
