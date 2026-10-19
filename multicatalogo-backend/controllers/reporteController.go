package controllers

import (
	"time"

	"multicatalogo-backend/repository"

	"github.com/gofiber/fiber/v2"
)

// ReporteController expone los KPIs del negocio (solo admin).
type ReporteController struct {
	base
	repo repository.DBRepository
}

// Resumen: GET /api/reportes/resumen?desde=AAAA-MM-DD&hasta=AAAA-MM-DD
func (ctl *ReporteController) Resumen(c *fiber.Ctx) error {
	desde, err := leerFecha(c.Query("desde"))
	if err != nil {
		return ctl.errorJSON(c, fiber.StatusBadRequest, "La fecha 'desde' debe tener el formato AAAA-MM-DD", nil)
	}
	hasta, err := leerFecha(c.Query("hasta"))
	if err != nil {
		return ctl.errorJSON(c, fiber.StatusBadRequest, "La fecha 'hasta' debe tener el formato AAAA-MM-DD", nil)
	}
	if desde != nil && hasta != nil && hasta.Before(*desde) {
		return ctl.errorJSON(c, fiber.StatusBadRequest, "'hasta' no puede ser anterior a 'desde'", nil)
	}

	res, err := ctl.repo.ResumenVentas(c.Context(), desde, hasta)
	if err != nil {
		return ctl.errorInterno(c, err)
	}
	if desde != nil {
		d := c.Query("desde")
		res.Desde = &d
	}
	if hasta != nil {
		h := c.Query("hasta")
		res.Hasta = &h
	}
	return c.JSON(res)
}

// TopProductos: GET /api/reportes/top-productos?limit=5
func (ctl *ReporteController) TopProductos(c *fiber.Ctx) error {
	datos, err := ctl.repo.TopProductos(c.Context(), limiteTop(c))
	if err != nil {
		return ctl.errorInterno(c, err)
	}
	return c.JSON(datos)
}

// TopReferidos: GET /api/reportes/top-referidos?limit=5
func (ctl *ReporteController) TopReferidos(c *fiber.Ctx) error {
	datos, err := ctl.repo.TopReferidos(c.Context(), limiteTop(c))
	if err != nil {
		return ctl.errorInterno(c, err)
	}
	return c.JSON(datos)
}

// ComisionesPendientes: GET /api/reportes/comisiones-pendientes — deuda por distribuidor.
func (ctl *ReporteController) ComisionesPendientes(c *fiber.Ctx) error {
	datos, err := ctl.repo.ComisionesPendientes(c.Context())
	if err != nil {
		return ctl.errorInterno(c, err)
	}
	return c.JSON(datos)
}

func leerFecha(texto string) (*time.Time, error) {
	if texto == "" {
		return nil, nil
	}
	fecha, err := time.ParseInLocation("2006-01-02", texto, time.Local)
	if err != nil {
		return nil, err
	}
	return &fecha, nil
}

func limiteTop(c *fiber.Ctx) int {
	limite := c.QueryInt("limit", 5)
	if limite < 1 || limite > 50 {
		return 5
	}
	return limite
}
