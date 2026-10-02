// Seguimos dentro del paquete controllers, ya que maneja la lógica de otra sección del negocio.
package controllers

import (
	"strconv"

	// Importamos Fiber para manejar la respuesta HTTP.
	"github.com/gofiber/fiber/v2"
	// Importamos nuestro paquete de modelos para usar la estructura Producto.
	"multicatalogo-backend/models"
)

// GetProductos es la función controladora encargada de devolver el catálogo de artículos.
func GetProductos(c *fiber.Ctx) error {
	return c.JSON(listaProductos())
}

// listaProductos devuelve el catálogo estático (se reemplazará por el repositorio en la persistencia).
func listaProductos() []models.Producto {
	return []models.Producto{
		// Agregamos el primer producto con sus respectivos valores para ID, Nombre, Precio e Img.
		{ID: 1, Nombre: "Serum Revitalizante", Precio: 45.00, Img: "https://picsum.photos/seed/serum/150"},
		// Agregamos el segundo producto a la lista.
		{ID: 2, Nombre: "Crema Hidratante Pro", Precio: 32.50, Img: "https://picsum.photos/seed/crema/150"},
		// Agregamos el tercer producto a la lista.
		{ID: 3, Nombre: "Tónico Purificante", Precio: 28.00, Img: "https://picsum.photos/seed/tonico/150"},
		// Agregamos el cuarto producto a la lista.
		{ID: 4, Nombre: "Mascarilla Nocturna", Precio: 50.00, Img: "https://picsum.photos/seed/mascarilla/150"},
	}
}

// GetProductoPorID devuelve un producto según el parámetro :id de la URL.
func GetProductoPorID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			models.NewAPIError(fiber.StatusBadRequest, "El id debe ser un número entero", c.Params("id")))
	}

	for _, p := range listaProductos() {
		if p.ID == id {
			return c.JSON(p)
		}
	}

	return c.Status(fiber.StatusNotFound).JSON(
		models.NewAPIError(fiber.StatusNotFound, "Producto no encontrado", id))
}