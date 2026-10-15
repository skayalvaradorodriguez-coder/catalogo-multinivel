package controllers

import (
	"strings"

	"multicatalogo-backend/cache"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"

	"github.com/gofiber/fiber/v2"
)

// GetRed devuelve la red multinivel como árbol (misma forma que usaba el mock del frontend).
// Armar el árbol es costoso, así que se guarda en Redis por usuario:
// clave red:tree:{email}, con TTL de 10 minutos.
//
// El email llega como parámetro: GET /api/red?email=admin@upse.edu.ec
func GetRed(c *fiber.Ctx) error {
	ctx := c.Context()

	email := strings.ToLower(strings.TrimSpace(c.Query("email")))
	if email == "" {
		email = "invitado"
	}
	clave := cache.PrefijoRed + email

	// Cache Hit: el árbol ya está armado en Redis.
	var red models.Referido
	if cache.Obtener(ctx, clave, &red) {
		c.Set("X-Cache", "HIT")
		return c.JSON(red)
	}

	// Cache Miss: lo armamos desde PostgreSQL y lo guardamos.
	arbol, err := repository.ObtenerRed(ctx)
	if err != nil {
		return errorInterno(c, err)
	}
	if arbol == nil {
		return c.Status(fiber.StatusNotFound).JSON(
			models.NewAPIError(fiber.StatusNotFound, "La red de referidos está vacía", nil))
	}
	cache.Guardar(ctx, clave, arbol, cache.TTLRed)

	c.Set("X-Cache", "MISS")
	return c.JSON(arbol)
}
