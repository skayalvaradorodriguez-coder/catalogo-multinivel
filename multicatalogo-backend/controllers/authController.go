package controllers

import (
	"strings"

	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"

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

	// Buscar el usuario en PostgreSQL a través del repositorio.
	usuario, err := repository.BuscarUsuarioPorCredenciales(c.Context(), email, password)
	if err != nil {
		return errorInterno(c, err)
	}

	if usuario != nil {
		// Token falso por rol (igual que antes) hasta implementar JWT real.
		token := "fake-jwt-token-456"
		if usuario.Rol == "admin" {
			token = "fake-jwt-token-123"
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"token": token,
			"email": usuario.Email,
			"rol":   usuario.Rol,
		})
	}

	// Si ningún usuario coincide.
	return c.Status(fiber.StatusUnauthorized).JSON(
		models.NewAPIError(fiber.StatusUnauthorized, "Credenciales incorrectas", nil))
}