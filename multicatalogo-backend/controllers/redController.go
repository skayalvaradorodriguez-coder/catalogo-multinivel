package controllers

import (
	"strings"

	"multicatalogo-backend/cache"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"

	"github.com/gofiber/fiber/v2"
)

// RedController maneja la red multinivel (Cache-Aside, TTL 10 min).
type RedController struct {
	base
	repo  repository.DBRepository
	cache cache.CacheRepository
}

// GetRed: GET /api/red (admin) — árbol completo. Clave red:tree:global.
func (ctl *RedController) GetRed(c *fiber.Ctx) error {
	return ctl.responder(c, cache.ClaveRed("global"), func() (*models.Referido, error) {
		return ctl.repo.ObtenerRed(c.Context())
	})
}

// GetMiRed: GET /api/red/mia — subárbol del usuario del token. Clave red:tree:{email}.
func (ctl *RedController) GetMiRed(c *fiber.Ctx) error {
	claims := usuarioToken(c)
	usuario, err := ctl.repo.BuscarUsuarioPorID(c.Context(), claims.UsuarioID())
	if err != nil {
		return ctl.errorRepo(c, err, "Usuario no encontrado")
	}
	if usuario.ReferidoID == nil {
		return ctl.errorJSON(c, fiber.StatusNotFound, "Tu cuenta todavía no pertenece a la red", nil)
	}
	return ctl.responder(c, cache.ClaveRed(strings.ToLower(usuario.Email)), func() (*models.Referido, error) {
		return ctl.repo.ObtenerSubRed(c.Context(), *usuario.ReferidoID)
	})
}

func (ctl *RedController) responder(c *fiber.Ctx, clave string, consultar func() (*models.Referido, error)) error {
	var red models.Referido
	if ctl.cache.Obtener(c.Context(), clave, &red) {
		c.Set("X-Cache", "HIT")
		return c.JSON(red)
	}
	arbol, err := consultar()
	if err != nil {
		return ctl.errorRepo(c, err, "La red de referidos está vacía")
	}
	ctl.cache.Guardar(c.Context(), clave, arbol, cache.TTLRed)
	c.Set("X-Cache", "MISS")
	return c.JSON(arbol)
}
