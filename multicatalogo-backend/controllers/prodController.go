// Seguimos dentro del paquete controllers, ya que maneja la lógica de otra sección del negocio.
package controllers

import (
	"strconv"

	// Importamos Fiber para manejar la respuesta HTTP.
	"github.com/gofiber/fiber/v2"
	// Importamos la caché (Redis), los modelos y el repositorio, que es quien consulta PostgreSQL.
	"multicatalogo-backend/cache"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// GetProductos es la función controladora encargada de devolver el catálogo de artículos.
// Usa el patrón Cache-Aside: primero Redis y, solo si no está, PostgreSQL.
func GetProductos(c *fiber.Ctx) error {
	ctx := c.Context()

	// 1. Intentamos leer el catálogo desde Redis.
	var productos []models.Producto
	if cache.Obtener(ctx, cache.ClaveProductos, &productos) {
		// 2. Cache Hit: respondemos sin tocar PostgreSQL.
		c.Set("X-Cache", "HIT")
		return c.JSON(productos)
	}

	// 3. Cache Miss: consultamos PostgreSQL y guardamos el resultado por 5 minutos.
	productos, err := repository.ObtenerProductos(ctx)
	if err != nil {
		return errorInterno(c, err)
	}
	cache.Guardar(ctx, cache.ClaveProductos, productos, cache.TTLProductos)

	c.Set("X-Cache", "MISS")
	return c.JSON(productos)
}

// GetProductoPorID devuelve un producto según el parámetro :id de la URL.
func GetProductoPorID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			models.NewAPIError(fiber.StatusBadRequest, "El id debe ser un número entero", c.Params("id")))
	}

	producto, err := repository.ObtenerProductoPorID(c.Context(), id)
	if err != nil {
		return errorInterno(c, err)
	}
	if producto != nil {
		return c.JSON(producto)
	}

	return c.Status(fiber.StatusNotFound).JSON(
		models.NewAPIError(fiber.StatusNotFound, "Producto no encontrado", id))
}