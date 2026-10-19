package controllers

import (
	"strings"

	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"

	"github.com/gofiber/fiber/v2"
)

// UsuarioController es la administración de cuentas (solo admin).
type UsuarioController struct {
	base
	repo repository.DBRepository
}

// Listar: GET /api/usuarios?page&limit&q
func (ctl *UsuarioController) Listar(c *fiber.Ctx) error {
	page, limit := paginacion(c)
	pagina, err := ctl.repo.ListarUsuarios(c.Context(), c.Query("q"), page, limit)
	if err != nil {
		return ctl.errorInterno(c, err)
	}
	return c.JSON(pagina)
}

// Actualizar: PATCH /api/usuarios/:id {rol?, activo?}
func (ctl *UsuarioController) Actualizar(c *fiber.Ctx) error {
	id, err := ctl.paramID(c)
	if err != nil {
		return err
	}
	var cambios models.CambiosUsuario
	if err := c.BodyParser(&cambios); err != nil {
		return ctl.errorJSON(c, fiber.StatusBadRequest, "Cuerpo de petición inválido", nil)
	}
	if cambios.Rol == nil && cambios.Activo == nil {
		return ctl.errorJSON(c, fiber.StatusBadRequest, "Envía rol y/o activo", nil)
	}
	if cambios.Rol != nil {
		rol := strings.ToLower(strings.TrimSpace(*cambios.Rol))
		if rol != "admin" && rol != "cliente" {
			return ctl.errorJSON(c, fiber.StatusBadRequest, "El rol debe ser admin o cliente", nil)
		}
		cambios.Rol = &rol
	}
	if id == usuarioToken(c).UsuarioID() {
		return ctl.errorJSON(c, fiber.StatusBadRequest, "No puedes cambiar el rol ni desactivar tu propia cuenta", nil)
	}

	usuario, err := ctl.repo.ActualizarUsuario(c.Context(), id, cambios)
	if err != nil {
		return ctl.errorRepo(c, err, "Usuario no encontrado")
	}
	return c.JSON(usuario)
}
