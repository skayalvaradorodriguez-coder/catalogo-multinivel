package controllers

import (
	"strconv"

	"multicatalogo-backend/cache"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"

	"github.com/gofiber/fiber/v2"
)

// ProductoController maneja el catálogo. Recibe el repositorio y la caché por inyección de dependencias.
type ProductoController struct {
	repo  repository.DBRepository
	cache cache.CacheRepository
}

// NuevoProductoController crea el controlador con sus dependencias.
func NuevoProductoController(repo repository.DBRepository, cache cache.CacheRepository) *ProductoController {
	return &ProductoController{repo: repo, cache: cache}
}

// GetProductos devuelve el catálogo usando el patrón Cache-Aside:
// primero Redis y, solo si no está, PostgreSQL.
func (ctl *ProductoController) GetProductos(c *fiber.Ctx) error {
	ctx := c.Context()

	// 1. Intentamos leer el catálogo desde Redis.
	var productos []models.Producto
	if ctl.cache.Obtener(ctx, cache.ClaveProductos, &productos) {
		// 2. Cache Hit: respondemos sin tocar PostgreSQL.
		c.Set("X-Cache", "HIT")
		return c.JSON(productos)
	}

	// 3. Cache Miss: consultamos PostgreSQL y guardamos el resultado por 5 minutos.
	productos, err := ctl.repo.ObtenerProductos(ctx)
	if err != nil {
		return errorInterno(c, err)
	}
	ctl.cache.Guardar(ctx, cache.ClaveProductos, productos, cache.TTLProductos)

	c.Set("X-Cache", "MISS")
	return c.JSON(productos)
}

// GetProductoPorID devuelve un producto según el parámetro :id de la URL.
func (ctl *ProductoController) GetProductoPorID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			models.NewAPIError(fiber.StatusBadRequest, "El id debe ser un número entero", c.Params("id")))
	}

	producto, err := ctl.repo.ObtenerProductoPorID(c.Context(), id)
	if err != nil {
		return errorInterno(c, err)
	}
	if producto != nil {
		return c.JSON(producto)
	}

	return c.Status(fiber.StatusNotFound).JSON(
		models.NewAPIError(fiber.StatusNotFound, "Producto no encontrado", id))
}
