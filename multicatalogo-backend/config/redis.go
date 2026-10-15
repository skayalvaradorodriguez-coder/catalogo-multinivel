package config

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// Redis es el cliente de caché. Queda en nil si Redis no está disponible:
// en ese caso la API sigue funcionando, solo que todas las lecturas van a PostgreSQL.
var Redis *redis.Client

// ConectarRedis crea el cliente leyendo las variables de entorno.
//   - REDIS_URL  (opcional, tiene prioridad): redis://usuario:clave@host:6379/0
//   - REDIS_ADDR (por defecto localhost:6379), REDIS_PASSWORD y REDIS_DB
//
// Debe llamarse después de ConectarDB, que es quien carga el archivo .env.
func ConectarRedis() {
	var opciones *redis.Options

	if uri := getEnv("REDIS_URL", ""); uri != "" {
		parsed, err := redis.ParseURL(uri)
		if err != nil {
			log.Printf("Aviso: REDIS_URL inválida (%v); se continúa sin caché", err)
			return
		}
		opciones = parsed
	} else {
		db, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
		if err != nil {
			db = 0
		}
		opciones = &redis.Options{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       db,
		}
	}

	// Tiempos cortos: si Redis se cae, no queremos que una petición espere segundos.
	opciones.DialTimeout = 2 * time.Second
	opciones.ReadTimeout = 500 * time.Millisecond
	opciones.WriteTimeout = 500 * time.Millisecond

	cliente := redis.NewClient(opciones)

	ctx, cancelar := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelar()

	if err := cliente.Ping(ctx).Err(); err != nil {
		log.Printf("Aviso: no se pudo conectar a Redis en %s (%v); se continúa sin caché", opciones.Addr, err)
		_ = cliente.Close()
		return
	}

	Redis = cliente
	log.Printf("Conectado a Redis en %s", opciones.Addr)
}

// CerrarRedis libera las conexiones del cliente. Se llama al finalizar la aplicación.
func CerrarRedis() {
	if Redis != nil {
		_ = Redis.Close()
		log.Println("Conexión a Redis cerrada")
	}
}
