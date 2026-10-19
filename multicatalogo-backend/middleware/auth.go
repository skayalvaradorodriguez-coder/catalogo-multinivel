// Package middleware contiene los filtros HTTP de autenticación y autorización.
package middleware

import (
	"strings"

	"multicatalogo-backend/models"
	"multicatalogo-backend/seguridad"

	"github.com/gofiber/fiber/v2"
)

const claveUsuario = "usuario"

// RequireAuth exige un header "Authorization: Bearer <token>" válido.
// Guarda los claims en c.Locals para que los controladores tomen la identidad del TOKEN.
func RequireAuth(secreto string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get(fiber.HeaderAuthorization)
		token, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || strings.TrimSpace(token) == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(
				models.NewAPIError(fiber.StatusUnauthorized, "Debes iniciar sesión", nil))
		}

		claims, err := seguridad.ValidarToken(secreto, strings.TrimSpace(token))
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(
				models.NewAPIError(fiber.StatusUnauthorized, "Sesión inválida o expirada, vuelve a iniciar sesión", nil))
		}

		c.Locals(claveUsuario, claims)
		return c.Next()
	}
}

// RequireAdmin exige que el token sea de un administrador. Va DESPUÉS de RequireAuth.
func RequireAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := UsuarioActual(c)
		if claims == nil || !claims.EsAdmin() {
			return c.Status(fiber.StatusForbidden).JSON(
				models.NewAPIError(fiber.StatusForbidden, "No tienes permisos de administrador", nil))
		}
		return c.Next()
	}
}

// UsuarioActual devuelve los claims del token de la petición (nil si la ruta no usa RequireAuth).
func UsuarioActual(c *fiber.Ctx) *seguridad.Claims {
	claims, _ := c.Locals(claveUsuario).(*seguridad.Claims)
	return claims
}
