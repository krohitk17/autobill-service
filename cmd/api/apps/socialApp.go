package apps

import (
	SocialAdapter "autobill-service/internal/adapters/inbound/http/social"
	RepositoryAdapters "autobill-service/internal/adapters/outbound/db"
	SocialApp "autobill-service/internal/application/social"
	DB "autobill-service/internal/infrastructure/db"
	JWTUtil "autobill-service/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

func CreateSocialApp(util JWTUtil.JWTUtil, db DB.PostgresDB) SocialAdapter.SocialRouter {
	socialAppFiber := fiber.New(fiber.Config{
		AppName: "autobill-social-service",
	})

	socialRepo := RepositoryAdapters.CreateSocialRepository(db)

	socialService := SocialApp.CreateSocialService(socialRepo)

	socialHandler := SocialAdapter.CreateSocialHandler(socialService)

	router := SocialAdapter.CreateSocialRouter(socialAppFiber, socialHandler, util)
	router.RegisterRoutes()

	return router
}
