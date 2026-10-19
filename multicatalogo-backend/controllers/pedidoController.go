package controllers

import (
	"errors"
	"strings"

	"multicatalogo-backend/cache"
	"multicatalogo-backend/domain"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"

	"github.com/gofiber/fiber/v2"
)

// PedidoController es la consulta y gestión de pedidos.
type PedidoController struct {
	base
	repo  repository.DBRepository
	cache cache.CacheRepository
}

// Mios: GET /api/pedidos/mios?page&limit — solo los del usuario del token.
func (ctl *PedidoController) Mios(c *fiber.Ctx) error {
	id := usuarioToken(c).UsuarioID()
	page, limit := paginacion(c)
	pagina, err := ctl.repo.ListarPedidos(c.Context(), models.FiltroPedidos{UsuarioID: &id, Page: page, Limit: limit})
	if err != nil {
		return ctl.errorInterno(c, err)
	}
	return c.JSON(pagina)
}

// Listar: GET /api/pedidos?page&limit&estado (admin).
func (ctl *PedidoController) Listar(c *fiber.Ctx) error {
	estado := strings.ToLower(strings.TrimSpace(c.Query("estado")))
	if estado != "" && !domain.EstadoValido(estado) {
		return ctl.errorJSON(c, fiber.StatusBadRequest, "Estado de pedido desconocido", estado)
	}
	page, limit := paginacion(c)
	pagina, err := ctl.repo.ListarPedidos(c.Context(), models.FiltroPedidos{Estado: estado, Page: page, Limit: limit})
	if err != nil {
		return ctl.errorInterno(c, err)
	}
	return c.JSON(pagina)
}

// Detalle: GET /api/pedidos/:id — el dueño o un admin. A otros se les responde 404
// (no revelamos que el pedido existe).
func (ctl *PedidoController) Detalle(c *fiber.Ctx) error {
	id, err := ctl.paramID(c)
	if err != nil {
		return err
	}
	pedido, err := ctl.repo.ObtenerPedido(c.Context(), id)
	claims := usuarioToken(c)
	if err == nil && !claims.EsAdmin() && pedido.UsuarioID != claims.UsuarioID() {
		err = repository.ErrNoEncontrado
	}
	if err != nil {
		return ctl.errorRepo(c, err, "Pedido no encontrado")
	}
	return c.JSON(pedido)
}

// CambiarEstado: PATCH /api/pedidos/:id/estado {estado} (admin).
func (ctl *PedidoController) CambiarEstado(c *fiber.Ctx) error {
	id, err := ctl.paramID(c)
	if err != nil {
		return err
	}
	var cuerpo struct {
		Estado string `json:"estado"`
	}
	if err := c.BodyParser(&cuerpo); err != nil {
		return ctl.errorJSON(c, fiber.StatusBadRequest, "Cuerpo de petición inválido", nil)
	}
	nuevo := strings.ToLower(strings.TrimSpace(cuerpo.Estado))
	if !domain.EstadoValido(nuevo) {
		return ctl.errorJSON(c, fiber.StatusBadRequest, "Estado de pedido desconocido", nuevo)
	}

	pedido, err := ctl.repo.CambiarEstadoPedido(c.Context(), id, nuevo)
	if err != nil {
		var errValidacion *repository.ErrValidacion
		if errors.As(err, &errValidacion) {
			return ctl.errorJSON(c, fiber.StatusBadRequest, errValidacion.Mensaje, nil)
		}
		return ctl.errorRepo(c, err, "Pedido no encontrado")
	}

	if nuevo == domain.EstadoCancelado {
		// Volvió stock y cambiaron las ventas: invalidamos catálogo y árboles.
		ctl.cache.Eliminar(c.Context(), cache.ClaveProductos)
		ctl.cache.EliminarPatron(c.Context(), cache.PatronRed)
	}
	return c.JSON(pedido)
}
