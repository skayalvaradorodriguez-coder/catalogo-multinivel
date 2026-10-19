// Package routes registra los endpoints de la API con sus middlewares de seguridad.
package routes

import (
	"multicatalogo-backend/controllers"
	"multicatalogo-backend/middleware"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes recibe la aplicación, los controladores ya inyectados y la clave JWT.
// Regla: toda ruta de escritura lleva RequireAuth; las de administración, además, RequireAdmin.
func SetupRoutes(app *fiber.App, ctl *controllers.Controllers, jwtSecret string) {
	auth := middleware.RequireAuth(jwtSecret)
	admin := middleware.RequireAdmin()
	api := app.Group("/api")

	// Públicas
	api.Get("/health", ctl.Health.Health)
	api.Post("/login", ctl.Auth.Login)
	api.Post("/register", ctl.Auth.Register)
	api.Get("/productos", ctl.Productos.GetProductos)
	api.Get("/productos/:id", ctl.Productos.GetProductoPorID)

	// Usuario autenticado (cualquier rol)
	api.Get("/red/mia", auth, ctl.Red.GetMiRed)
	api.Post("/checkout", auth, ctl.Checkout.Checkout)
	api.Get("/pedidos/mios", auth, ctl.Pedidos.Mios) // antes que /pedidos/:id
	api.Get("/pedidos/:id", auth, ctl.Pedidos.Detalle)
	api.Get("/comisiones/mias", auth, ctl.Comisiones.Mias)
	api.Get("/liquidaciones/mias", auth, ctl.Liquidaciones.Mias)

	// Solo administrador
	api.Get("/red", auth, admin, ctl.Red.GetRed)

	api.Get("/admin/productos", auth, admin, ctl.Productos.ListarAdmin)
	api.Post("/productos", auth, admin, ctl.Productos.Crear)
	api.Put("/productos/:id", auth, admin, ctl.Productos.Actualizar)
	api.Delete("/productos/:id", auth, admin, ctl.Productos.Desactivar)

	api.Get("/pedidos", auth, admin, ctl.Pedidos.Listar)
	api.Patch("/pedidos/:id/estado", auth, admin, ctl.Pedidos.CambiarEstado)

	api.Get("/usuarios", auth, admin, ctl.Usuarios.Listar)
	api.Patch("/usuarios/:id", auth, admin, ctl.Usuarios.Actualizar)

	api.Post("/liquidaciones", auth, admin, ctl.Liquidaciones.Crear)
	api.Get("/liquidaciones", auth, admin, ctl.Liquidaciones.Listar)

	api.Get("/reportes/resumen", auth, admin, ctl.Reportes.Resumen)
	api.Get("/reportes/top-productos", auth, admin, ctl.Reportes.TopProductos)
	api.Get("/reportes/top-referidos", auth, admin, ctl.Reportes.TopReferidos)
	api.Get("/reportes/comisiones-pendientes", auth, admin, ctl.Reportes.ComisionesPendientes)
}
