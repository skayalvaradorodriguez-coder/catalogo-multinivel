package config

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// ConectarRedis crea el cliente leyendo las variables de entorno.
// Devuelve nil si Redis no está disponible: la API sigue funcionando, solo que sin caché.
//   - REDIS_URL  (opcional, tiene prioridad): redis://usuario:clave@host:6379/0
//   - REDIS_ADDR (por defecto localhost:6379), REDIS_PASSWORD y REDIS_DB
//
// Debe llamarse después de CargarEnv.
func ConectarRedis() *redis.Client {
	var opciones *redis.Options

	if uri := getEnv("REDIS_URL", ""); uri != "" {
		parsed, err := redis.ParseURL(uri)
		if err != nil {
			log.Printf("Aviso: REDIS_URL inválida (%v); se continúa sin caché", err)
			return nil
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
		return nil
	}

	log.Printf("Conectado a Redis en %s", opciones.Addr)
	return cliente
}
