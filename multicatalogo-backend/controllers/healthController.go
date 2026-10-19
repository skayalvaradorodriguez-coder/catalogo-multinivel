package controllers

import (
	"context"
	"time"

	"multicatalogo-backend/cache"
	"multicatalogo-backend/repository"

	"github.com/gofiber/fiber/v2"
)

// HealthController informa el estado real de las dependencias.
type HealthController struct {
	base
	repo  repository.DBRepository
	cache cache.CacheRepository
}

// Health: 200 "ok" si todo responde, 200 "degraded" si cae Redis (la API sigue sin caché)
// y 503 "down" si cae PostgreSQL.
func (ctl *HealthController) Health(c *fiber.Ctx) error {
	ctx, cancelar := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelar()

	respuesta := fiber.Map{"status": "ok", "postgres": "ok", "redis": "ok", "message": "Servidor Go/Fiber operativo"}
	status := fiber.StatusOK

	if err := ctl.cache.Ping(ctx); err != nil {
		respuesta["status"] = "degraded"
		respuesta["redis"] = "down"
		respuesta["message"] = "Redis no responde: la API funciona sin caché"
	}
	if err := ctl.repo.Ping(ctx); err != nil {
		respuesta["status"] = "down"
		respuesta["postgres"] = "down"
		respuesta["message"] = "PostgreSQL no responde"
		status = fiber.StatusServiceUnavailable
	}
	return c.Status(status).JSON(respuesta)
}
