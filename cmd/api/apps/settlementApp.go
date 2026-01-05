package apps

import (
	SettlementAdapter "autobill-service/internal/adapters/inbound/http/settlement"
	RepositoryAdapters "autobill-service/internal/adapters/outbound/db"
	SettlementApp "autobill-service/internal/application/settlement"
	DB "autobill-service/internal/infrastructure/db"
	JWTUtil "autobill-service/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

func CreateSettlementApp(util JWTUtil.JWTUtil, db DB.PostgresDB) SettlementAdapter.SettlementRouter {
	settlementAppFiber := fiber.New(fiber.Config{
		AppName: "autobill-settlement-service",
	})

	settlementRepo := RepositoryAdapters.CreateSettlementRepository(db)
	splitRepo := RepositoryAdapters.CreateSplitRepository(db)

	settlementService := SettlementApp.CreateSettlementService(settlementRepo, splitRepo)

	settlementHandler := SettlementAdapter.CreateSettlementHandler(settlementService)

	router := SettlementAdapter.CreateSettlementRouter(settlementAppFiber, settlementHandler, util)
	router.RegisterRoutes()

	return router
}
