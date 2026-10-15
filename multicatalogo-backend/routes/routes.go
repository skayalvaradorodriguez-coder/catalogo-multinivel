// Declaramos el paquete routes, cuya única responsabilidad es registrar los endpoints de la API.
package routes

import (
	// Importamos Fiber para poder recibir y configurar la instancia principal de la aplicación.
	"github.com/gofiber/fiber/v2"
	// Importamos nuestro paquete de controladores, donde residen las funciones que se ejecutarán en cada ruta.
	"multicatalogo-backend/controllers"
)

// Controladores agrupa los controladores ya construidos con sus dependencias (ver main.go).
type Controladores struct {
	Auth      *controllers.AuthController
	Productos *controllers.ProductoController
	Red       *controllers.RedController
	Checkout  *controllers.CheckoutController
}

// SetupRoutes recibe la aplicación Fiber y los controladores inyectados, y registra las rutas.
func SetupRoutes(app *fiber.App, ctl Controladores) {
	// Inicio de sesión.
	app.Post("/api/login", ctl.Auth.Login)

	// Catálogo completo (con caché en Redis).
	app.Get("/api/productos", ctl.Productos.GetProductos)

	// Producto individual por ID.
	app.Get("/api/productos/:id", ctl.Productos.GetProductoPorID)

	// Red multinivel de referidos (árbol jerárquico, con caché en Redis).
	app.Get("/api/red", ctl.Red.GetRed)

	// Fase 4: compra con transacción ACID y comisiones en segundo plano.
	app.Post("/api/checkout", ctl.Checkout.Checkout)

	// Verificación de estado del servidor.
	app.Get("/api/health", controllers.Health)
}
