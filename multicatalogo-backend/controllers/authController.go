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
		return c.Status(fiber.StatusBadRequest).JSON(
			models.NewAPIError(fiber.StatusBadRequest, "Cuerpo de petición inválido", nil))
	}

	// Limpiar espacios innecesarios.
	email := strings.TrimSpace(req.Email)
	password := strings.TrimSpace(req.Password)

	// Validar que email y password no vengan vacíos.
	if email == "" || password == "" {
		details := map[string]string{}
		if email == "" {
			details["email"] = "El email es obligatorio"
		}
		if password == "" {
			details["password"] = "La contraseña es obligatoria"
		}
		return c.Status(fiber.StatusBadRequest).JSON(
			models.NewAPIError(fiber.StatusBadRequest, "Email y contraseña son obligatorios", details))
	}

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
	return c.Status(fiber.StatusUnauthorized).JSON(
		models.NewAPIError(fiber.StatusUnauthorized, "Credenciales incorrectas", nil))
}