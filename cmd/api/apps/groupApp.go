package apps

import (
	GroupAdapter "autobill-service/internal/adapters/inbound/http/group"
	RepositoryAdapters "autobill-service/internal/adapters/outbound/db"
	GroupApp "autobill-service/internal/application/group"
	DB "autobill-service/internal/infrastructure/db"
	JWTUtil "autobill-service/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

func CreateGroupApp(util JWTUtil.JWTUtil, db DB.PostgresDB) GroupAdapter.GroupRouter {
	groupAppFiber := fiber.New(fiber.Config{
		AppName: "autobill-group-service",
	})

	groupRepo := RepositoryAdapters.CreateGroupRepository(db)
	splitRepo := RepositoryAdapters.CreateSplitRepository(db)

	groupService := GroupApp.CreateGroupService(groupRepo, splitRepo)

	groupHandler := GroupAdapter.CreateGroupHandler(groupService)

	router := GroupAdapter.CreateGroupRouter(groupAppFiber, groupHandler, util)
	router.RegisterRoutes()

	return router
}
