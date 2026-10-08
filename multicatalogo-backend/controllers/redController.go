package controllers

import (
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"

	"github.com/gofiber/fiber/v2"
)

// GetRed devuelve la red multinivel como árbol (misma forma que usaba el mock del frontend).
func GetRed(c *fiber.Ctx) error {
	red, err := repository.ObtenerRed(c.Context())
	if err != nil {
		return errorInterno(c, err)
	}
	if red == nil {
		return c.Status(fiber.StatusNotFound).JSON(
			models.NewAPIError(fiber.StatusNotFound, "La red de referidos está vacía", nil))
	}
	return c.JSON(red)
}
