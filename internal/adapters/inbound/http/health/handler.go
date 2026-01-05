package health

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Handler struct {
	db        *gorm.DB
	startTime time.Time
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		db:        db,
		startTime: time.Now(),
	}
}

type HealthResponse struct {
	Status    string           `json:"status"`
	Timestamp time.Time        `json:"timestamp"`
	Uptime    string           `json:"uptime"`
	Version   string           `json:"version"`
	Checks    map[string]Check `json:"checks"`
}

type Check struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func (h *Handler) Health(c *fiber.Ctx) error {
	checks := make(map[string]Check)
	overallStatus := "healthy"

	dbCheck := h.checkDatabase()
	checks["database"] = dbCheck
	if dbCheck.Status != "healthy" {
		overallStatus = "unhealthy"
	}

	response := HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now().UTC(),
		Uptime:    time.Since(h.startTime).String(),
		Version:   "1.0.0",
		Checks:    checks,
	}

	statusCode := fiber.StatusOK
	if overallStatus != "healthy" {
		statusCode = fiber.StatusServiceUnavailable
	}

	return c.Status(statusCode).JSON(response)
}

func (h *Handler) Liveness(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "alive",
	})
}

func (h *Handler) Readiness(c *fiber.Ctx) error {
	dbCheck := h.checkDatabase()

	if dbCheck.Status != "healthy" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":  "not ready",
			"message": dbCheck.Message,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "ready",
	})
}

func (h *Handler) checkDatabase() Check {
	sqlDB, err := h.db.DB()
	if err != nil {
		return Check{
			Status:  "unhealthy",
			Message: "Failed to get database connection",
		}
	}

	if err := sqlDB.Ping(); err != nil {
		return Check{
			Status:  "unhealthy",
			Message: "Database ping failed",
		}
	}

	return Check{
		Status: "healthy",
	}
}

func RegisterRoutes(app *fiber.App, db *gorm.DB) {
	handler := NewHandler(db)

	health := app.Group("/health")
	health.Get("", handler.Health)
	health.Get("/live", handler.Liveness)
	health.Get("/ready", handler.Readiness)
}
