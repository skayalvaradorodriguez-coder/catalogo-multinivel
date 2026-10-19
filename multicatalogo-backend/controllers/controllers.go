// Package controllers traduce HTTP ↔ dominio. Cada controlador es un struct que recibe
// sus dependencias (interfaces) por constructor: nada de variables globales.
package controllers

import (
	"errors"
	"log"
	"strconv"

	"multicatalogo-backend/cache"
	"multicatalogo-backend/config"
	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"
	"multicatalogo-backend/notificaciones"
	"multicatalogo-backend/repository"
	"multicatalogo-backend/seguridad"

	"github.com/gofiber/fiber/v2"
)

// Controllers agrupa todos los controladores ya construidos.
type Controllers struct {
	Health        *HealthController
	Auth          *AuthController
	Usuarios      *UsuarioController
	Productos     *ProductoController
	Red           *RedController
	Checkout      *CheckoutController
	Pedidos       *PedidoController
	Comisiones    *ComisionController
	Liquidaciones *LiquidacionController
	Reportes      *ReporteController
}

// NewControllers es el punto de ensamblado de la inyección de dependencias.
func NewControllers(repo repository.DBRepository, cache cache.CacheRepository, mailer notificaciones.Mailer, cfg config.Config) *Controllers {
	b := base{dev: cfg.EsDev()}
	return &Controllers{
		Health:        &HealthController{base: b, repo: repo, cache: cache},
		Auth:          &AuthController{base: b, repo: repo, cache: cache, mailer: mailer, secreto: cfg.JWTSecret, ttl: cfg.JWTTTL},
		Usuarios:      &UsuarioController{base: b, repo: repo},
		Productos:     &ProductoController{base: b, repo: repo, cache: cache},
		Red:           &RedController{base: b, repo: repo, cache: cache},
		Checkout:      &CheckoutController{base: b, repo: repo, cache: cache, mailer: mailer},
		Pedidos:       &PedidoController{base: b, repo: repo, cache: cache},
		Comisiones:    &ComisionController{base: b, repo: repo},
		Liquidaciones: &LiquidacionController{base: b, repo: repo},
		Reportes:      &ReporteController{base: b, repo: repo},
	}
}

// base reúne las respuestas de error comunes. Details solo se envía con APP_ENV=dev.
type base struct {
	dev bool
}

func (b base) errorJSON(c *fiber.Ctx, status int, mensaje string, detalles interface{}) error {
	if !b.dev {
		detalles = nil
	}
	return c.Status(status).JSON(models.NewAPIError(status, mensaje, detalles))
}

// errorInterno registra el error real en consola y responde un 500 genérico.
func (b base) errorInterno(c *fiber.Ctx, err error) error {
	log.Printf("Error interno en %s %s: %v", c.Method(), c.Path(), err)
	var detalles interface{}
	if b.dev {
		detalles = err.Error()
	}
	return c.Status(fiber.StatusInternalServerError).JSON(
		models.NewAPIError(fiber.StatusInternalServerError, "Error interno del servidor", detalles))
}

// errorRepo traduce los errores del repositorio a HTTP: no encontrado → 404,
// duplicado → 409, validación → 400, el resto → 500.
func (b base) errorRepo(c *fiber.Ctx, err error, noEncontrado string) error {
	var errValidacion *repository.ErrValidacion
	switch {
	case errors.Is(err, repository.ErrNoEncontrado):
		return b.errorJSON(c, fiber.StatusNotFound, noEncontrado, nil)
	case errors.Is(err, repository.ErrDuplicado):
		return b.errorJSON(c, fiber.StatusConflict, "El registro ya existe", nil)
	case errors.Is(err, repository.ErrCantidadInvalida):
		return b.errorJSON(c, fiber.StatusBadRequest, "La cantidad debe ser mayor que cero", nil)
	case errors.As(err, &errValidacion):
		return b.errorJSON(c, fiber.StatusBadRequest, errValidacion.Mensaje, nil)
	}
	return b.errorInterno(c, err)
}

// paramID lee un :id numérico de la URL.
func (b base) paramID(c *fiber.Ctx) (int, error) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return 0, b.errorJSON(c, fiber.StatusBadRequest, "El id debe ser un número entero", c.Params("id"))
	}
	return id, nil
}

// paginacion lee ?page y ?limit (por defecto 1 y 10).
func paginacion(c *fiber.Ctx) (int, int) {
	return c.QueryInt("page", 1), c.QueryInt("limit", 10)
}

// usuarioToken devuelve los claims del JWT (la identidad SIEMPRE sale del token).
func usuarioToken(c *fiber.Ctx) *seguridad.Claims {
	return middleware.UsuarioActual(c)
}
