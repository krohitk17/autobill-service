package main

import (
	"autobill-service/cmd/api/apps"
	"autobill-service/internal/adapters/inbound/http/health"
	Middlewares "autobill-service/internal/adapters/inbound/http/middleware"
	Config "autobill-service/internal/infrastructure/config"
	DB "autobill-service/internal/infrastructure/db"
	JWTUtil "autobill-service/pkg/jwt"
	Logger "autobill-service/pkg/logger"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	config := Config.Load()

	Logger.Info().
		Str("environment", string(config.Environment)).
		Str("port", config.Server.Port).
		Msg("Starting autobill-service")

	db, dbErr := DB.CreatePostgresDb(config.DBUrl)
	if dbErr != nil {
		Logger.Fatal().Err(dbErr).Msg("Failed to connect to database")
	}

	util := JWTUtil.CreateJwtUtil(config.JWT.Secret, config.JWT.Expiration, config.JWT.RefreshTokenExpiration)

	app := fiber.New(fiber.Config{
		AppName:      "autobill-service",
		ReadTimeout:  config.Server.ReadTimeout,
		WriteTimeout: config.Server.WriteTimeout,
		ErrorHandler: Middlewares.GlobalErrorHandler,
	})

	registerMiddleware(app, config)

	MountApps(app, util, *db)

	Logger.Info().
		Str("app", app.Config().AppName).
		Str("port", config.Server.Port).
		Msg("Server starting")

	go func() {
		if err := app.Listen(":" + config.Server.Port); err != nil {
			Logger.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	gracefulShutdown(app, 30*time.Second)
}

func registerMiddleware(app *fiber.App, config Config.Config) {
	app.Use(recover.New(recover.Config{
		EnableStackTrace: config.IsDevelopment(),
	}))

	app.Use(Middlewares.TimeoutMiddleware(Middlewares.TimeoutConfig{
		Timeout: 30 * time.Second,
		Message: "Request timeout",
	}))

	app.Use(Middlewares.CORSMiddleware(Middlewares.CORSConfig{
		AllowOrigins:     config.CORS.AllowOrigins,
		AllowMethods:     config.CORS.AllowMethods,
		AllowHeaders:     config.CORS.AllowHeaders,
		AllowCredentials: config.CORS.AllowCredentials,
		MaxAge:           86400,
	}))

	app.Use(Middlewares.RequestContextMiddleware())

	app.Use(Middlewares.NewRateLimiter(Middlewares.RateLimitConfig{
		Max:        config.RateLimit.MaxRequests,
		Expiration: config.RateLimit.Window,
		Message:    "Too many requests. Please try again later.",
	}))
}

func MountApps(app *fiber.App, util JWTUtil.JWTUtil, db DB.PostgresDB) {
	health.RegisterRoutes(app, db.DB)
	app.Mount("/auth", apps.CreateAuthApp(util, db).App)
	app.Mount("/user", apps.CreateUserApp(util, db).App)
	app.Mount("/social", apps.CreateSocialApp(util, db).App)
	app.Mount("/groups", apps.CreateGroupApp(util, db).App)
	app.Mount("/splits", apps.CreateSplitApp(util, db).App)
	app.Mount("/settlements", apps.CreateSettlementApp(util, db).App)
	app.Mount("/balances", apps.CreateBalanceApp(util, db).App)
}

func gracefulShutdown(app *fiber.App, timeout time.Duration) {
	Logger.Info().Msg("Shutting down server...")
	if err := app.ShutdownWithTimeout(timeout); err != nil {
		Logger.Error().Err(err).Msg("Error during shutdown")
	}
	Logger.Info().Msg("Server stopped")
}
