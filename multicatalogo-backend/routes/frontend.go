package routes

import (
	"path/filepath"
	"strings"

	"multicatalogo-backend/models"

	"github.com/gofiber/fiber/v2"
)

// ServirFrontend entrega la build de React (carpeta dist) desde el mismo servidor que la API.
// Se registra DESPUÉS de las rutas /api para que estas tengan prioridad.
func ServirFrontend(app *fiber.App, dir string) {
	// Encabezados de caché: el service worker, el manifiesto y el HTML siempre se revalidan
	// (así la PWA detecta las versiones nuevas); los archivos con hash en el nombre son inmutables.
	app.Use(func(c *fiber.Ctx) error {
		ruta := c.Path()
		switch {
		case strings.HasPrefix(ruta, "/api/"):
		case ruta == "/manifest.webmanifest":
			c.Set(fiber.HeaderCacheControl, "no-cache")
			// Tipo correcto para que el navegador reconozca la PWA como instalable.
			if err := c.Next(); err != nil {
				return err
			}
			c.Set(fiber.HeaderContentType, "application/manifest+json")
			return nil
		case ruta == "/sw.js" || ruta == "/" || ruta == "/index.html" || strings.HasPrefix(ruta, "/workbox-"):
			c.Set(fiber.HeaderCacheControl, "no-cache")
		case strings.HasPrefix(ruta, "/assets/"):
			c.Set(fiber.HeaderCacheControl, "public, max-age=31536000, immutable")
		}
		return c.Next()
	})

	app.Static("/", dir, fiber.Static{Compress: true, Index: "index.html"})

	indice := filepath.Join(dir, "index.html")
	app.Use(func(c *fiber.Ctx) error {
		// Una ruta de la API que no existe responde JSON, nunca la página de la app.
		if strings.HasPrefix(c.Path(), "/api/") {
			return c.Status(fiber.StatusNotFound).JSON(
				models.NewAPIError(fiber.StatusNotFound, "Ruta no encontrada", nil))
		}
		// Un archivo que no existe (imagen, .js...) responde 404 en lugar de la página.
		if filepath.Ext(c.Path()) != "" {
			return c.SendStatus(fiber.StatusNotFound)
		}
		// Rutas de React Router (/catalogo, /mis-pedidos...): se entrega index.html.
		c.Set(fiber.HeaderCacheControl, "no-cache")
		return c.SendFile(indice)
	})
}
