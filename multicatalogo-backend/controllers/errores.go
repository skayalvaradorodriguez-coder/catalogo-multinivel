package controllers

import (
	"log"

	"multicatalogo-backend/models"

	"github.com/gofiber/fiber/v2"
)

// errorInterno registra el error real en consola y responde un 500 genérico,
// para no exponer detalles de la base de datos al cliente.
func errorInterno(c *fiber.Ctx, err error) error {
	log.Printf("Error de base de datos en %s %s: %v", c.Method(), c.Path(), err)
	return c.Status(fiber.StatusInternalServerError).JSON(
		models.NewAPIError(fiber.StatusInternalServerError, "Error interno del servidor", nil))
}
