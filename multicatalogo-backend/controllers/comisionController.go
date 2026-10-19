package controllers

import (
	"multicatalogo-backend/repository"

	"github.com/gofiber/fiber/v2"
)

// ComisionController expone el ledger de comisiones del usuario.
type ComisionController struct {
	base
	repo repository.DBRepository
}

// Mias: GET /api/comisiones/mias — ledger del nodo del usuario del token + totales por estado.
func (ctl *ComisionController) Mias(c *fiber.Ctx) error {
	usuario, err := ctl.repo.BuscarUsuarioPorID(c.Context(), usuarioToken(c).UsuarioID())
	if err != nil {
		return ctl.errorRepo(c, err, "Usuario no encontrado")
	}
	if usuario.ReferidoID == nil {
		return ctl.errorJSON(c, fiber.StatusNotFound, "Tu cuenta todavía no pertenece a la red", nil)
	}
	cuenta, err := ctl.repo.EstadoCuenta(c.Context(), *usuario.ReferidoID)
	if err != nil {
		return ctl.errorInterno(c, err)
	}
	return c.JSON(cuenta)
}
