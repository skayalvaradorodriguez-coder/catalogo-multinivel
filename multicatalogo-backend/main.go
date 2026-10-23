// Declaramos que este es el paquete principal (main), es decir, el archivo ejecutable que inicia el programa.
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"multicatalogo-backend/cache"
	"multicatalogo-backend/config"
	"multicatalogo-backend/controllers"
	"multicatalogo-backend/notificaciones"
	"multicatalogo-backend/repository"
	"multicatalogo-backend/routes"
)

// main construye las dependencias UNA sola vez y las inyecta hacia abajo:
// configuración → conexiones → repositorio y caché → controladores → rutas.
func main() {
	config.CargarEnv()
	cfg, err := config.CargarConfig()
	if err != nil {
		log.Fatal(err)
	}

	// 1. Infraestructura. Si PostgreSQL falla, no tiene sentido arrancar.
	pool, err := config.ConectarDB()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		pool.Close()
		log.Println("Conexión a PostgreSQL cerrada")
	}()

	// Redis es opcional: si no está disponible, la caché queda desactivada.
	redisCache := cache.NuevaRedisCache(config.ConectarRedis())
	defer redisCache.Cerrar()

	// 2. Capa de datos. Las opciones permiten simular fallas del checkout en clase (ver .env).
	if cfg.SimularFallaCheckout {
		log.Println("ATENCIÓN: SIMULAR_FALLA_CHECKOUT activo, todos los checkouts harán ROLLBACK")
	}
	repo := repository.NuevoPostgresRepository(pool, repository.OpcionesCheckout{
		SimularFalla: cfg.SimularFallaCheckout,
		Pausa:        cfg.PausaCheckout,
	})

	// 3. Controladores con sus interfaces.
	ctl := controllers.NewControllers(repo, redisCache, notificaciones.LogMailer{}, cfg)

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSOrigins,
		// En desarrollo también se acepta el frontend abierto desde la red local
		// (la dirección "Network" que muestra Vite, p. ej. http://192.168.1.10:5173).
		AllowOriginsFunc: func(origen string) bool { return cfg.EsDev() && config.OrigenLocal(origen) },
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, PATCH, DELETE, OPTIONS",
		ExposeHeaders:    "X-Cache",
	}))

	// 4. Rutas con los controladores ya inyectados y los middlewares de seguridad.
	routes.SetupRoutes(app, ctl, cfg.JWTSecret)

	// En la nube el mismo servidor entrega también el frontend (SPA + PWA).
	if cfg.StaticDir != "" {
		routes.ServirFrontend(app, cfg.StaticDir)
		log.Printf("Sirviendo el frontend desde %s", cfg.StaticDir)
	}

	// Ctrl+C: apagado ordenado para que se ejecuten los defer que cierran Redis y PostgreSQL.
	go func() {
		senal := make(chan os.Signal, 1)
		signal.Notify(senal, os.Interrupt, syscall.SIGTERM)
		<-senal
		log.Println("Apagando servidor...")
		_ = app.Shutdown()
	}()

	log.Printf("Entorno: %s · puerto %s", cfg.Entorno, cfg.Puerto)
	if err := app.Listen(":" + cfg.Puerto); err != nil {
		log.Println("Servidor detenido:", err)
	}

	// Antes de cerrar la base de datos, dejamos terminar las goroutines de comisiones en curso.
	ctl.Checkout.EsperarComisiones()
}
