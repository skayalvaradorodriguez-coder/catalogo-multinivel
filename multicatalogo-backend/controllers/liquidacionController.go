package controllers

import (
	"multicatalogo-backend/repository"

	"github.com/gofiber/fiber/v2"
)

// LiquidacionController gestiona los pagos a distribuidores.
type LiquidacionController struct {
	base
	repo repository.DBRepository
}

// Crear: POST /api/liquidaciones {referido_id} (admin) — paga todo lo pendiente del distribuidor.
func (ctl *LiquidacionController) Crear(c *fiber.Ctx) error {
	var cuerpo struct {
		ReferidoID *int `json:"referido_id"`
	}
	if err := c.BodyParser(&cuerpo); err != nil || cuerpo.ReferidoID == nil {
		return ctl.errorJSON(c, fiber.StatusBadRequest, "Envía el referido_id del distribuidor a liquidar", nil)
	}
	liq, err := ctl.repo.Liquidar(c.Context(), *cuerpo.ReferidoID)
	if err != nil {
		return ctl.errorRepo(c, err, "Distribuidor no encontrado")
	}
	return c.Status(fiber.StatusCreated).JSON(liq)
}

// Listar: GET /api/liquidaciones?page&limit (admin) — historial de pagos.
func (ctl *LiquidacionController) Listar(c *fiber.Ctx) error {
	page, limit := paginacion(c)
	pagina, err := ctl.repo.ListarLiquidaciones(c.Context(), nil, page, limit)
	if err != nil {
		return ctl.errorInterno(c, err)
	}
	return c.JSON(pagina)
}

// Mias: GET /api/liquidaciones/mias — los cobros del distribuidor del token.
func (ctl *LiquidacionController) Mias(c *fiber.Ctx) error {
	usuario, err := ctl.repo.BuscarUsuarioPorID(c.Context(), usuarioToken(c).UsuarioID())
	if err != nil {
		return ctl.errorRepo(c, err, "Usuario no encontrado")
	}
	if usuario.ReferidoID == nil {
		return ctl.errorJSON(c, fiber.StatusNotFound, "Tu cuenta todavía no pertenece a la red", nil)
	}
	page, limit := paginacion(c)
	pagina, err := ctl.repo.ListarLiquidaciones(c.Context(), usuario.ReferidoID, page, limit)
	if err != nil {
		return ctl.errorInterno(c, err)
	}
	return c.JSON(pagina)
}
