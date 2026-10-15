package controllers

import (
	"strings"

	"multicatalogo-backend/cache"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"

	"github.com/gofiber/fiber/v2"
)

// RedController maneja la red multinivel. Recibe el repositorio y la caché por inyección de dependencias.
type RedController struct {
	repo  repository.DBRepository
	cache cache.CacheRepository
}

// NuevoRedController crea el controlador con sus dependencias.
func NuevoRedController(repo repository.DBRepository, cache cache.CacheRepository) *RedController {
	return &RedController{repo: repo, cache: cache}
}

// GetRed devuelve la red multinivel como árbol (misma forma que usaba el mock del frontend).
// Armar el árbol es costoso, así que se guarda en Redis por usuario:
// clave red:tree:{email}, con TTL de 10 minutos.
//
// El email llega como parámetro: GET /api/red?email=admin@upse.edu.ec
func (ctl *RedController) GetRed(c *fiber.Ctx) error {
	ctx := c.Context()

	email := strings.ToLower(strings.TrimSpace(c.Query("email")))
	if email == "" {
		email = "invitado"
	}
	clave := cache.ClaveRed(email)

	// Cache Hit: el árbol ya está armado en Redis.
	var red models.Referido
	if ctl.cache.Obtener(ctx, clave, &red) {
		c.Set("X-Cache", "HIT")
		return c.JSON(red)
	}

	// Cache Miss: lo armamos desde PostgreSQL y lo guardamos.
	arbol, err := ctl.repo.ObtenerRed(ctx)
	if err != nil {
		return errorInterno(c, err)
	}
	if arbol == nil {
		return c.Status(fiber.StatusNotFound).JSON(
			models.NewAPIError(fiber.StatusNotFound, "La red de referidos está vacía", nil))
	}
	ctl.cache.Guardar(ctx, clave, arbol, cache.TTLRed)

	c.Set("X-Cache", "MISS")
	return c.JSON(arbol)
}
