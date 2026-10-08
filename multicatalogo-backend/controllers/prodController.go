// Seguimos dentro del paquete controllers, ya que maneja la lógica de otra sección del negocio.
package controllers

import (
	"strconv"

	// Importamos Fiber para manejar la respuesta HTTP.
	"github.com/gofiber/fiber/v2"
	// Importamos modelos (APIError) y el repositorio, que es quien consulta PostgreSQL.
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// GetProductos es la función controladora encargada de devolver el catálogo de artículos.
func GetProductos(c *fiber.Ctx) error {
	productos, err := repository.ObtenerProductos(c.Context())
	if err != nil {
		return errorInterno(c, err)
	}
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