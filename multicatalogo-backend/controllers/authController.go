package controllers

import (
	"strings"

	"multicatalogo-backend/models"

	"github.com/gofiber/fiber/v2"
)

// Login valida las credenciales y devuelve el token y el rol del usuario.
func Login(c *fiber.Ctx) error {

	var req models.LoginRequest

	// Leer el JSON enviado por el cliente.
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cuerpo de petición inválido",
		})
	}

	// Limpiar espacios innecesarios.
	email := strings.TrimSpace(req.Email)
	password := strings.TrimSpace(req.Password)

	// Validar usuario administrador.
	if email == "admin@upse.edu.ec" && password == "123456" {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"token": "fake-jwt-token-123",
			"email": email,
			"rol":   "admin",
		})
	}

	// Validar usuario cliente.
	if email == "cliente@upse.edu.ec" && password == "123456" {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"token": "fake-jwt-token-456",
			"email": email,
			"rol":   "cliente",
		})
	}

	// Si ningún usuario coincide.
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"error": "Credenciales incorrectas",
	})
}
