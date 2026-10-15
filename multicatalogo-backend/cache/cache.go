// Package cache define el contrato de la caché (CacheRepository) y su implementación con Redis.
// Si Redis no está disponible, Obtener siempre reporta "miss" y Guardar/Eliminar no hacen nada,
// así la API sigue respondiendo desde PostgreSQL.
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// Claves y tiempos de expiración (TTL) definidos en la Fase 3.
const (
	ClaveProductos = "productos:all"
	TTLProductos   = 5 * time.Minute

	PrefijoRed = "red:tree:" // + email del usuario
	TTLRed     = 10 * time.Minute
)

// ClaveRed arma la clave del árbol de un usuario: red:tree:{email}.
func ClaveRed(email string) string {
	return PrefijoRed + email
}

// CacheRepository es el contrato que usan los controladores.
// Gracias a la interfaz, en una prueba se puede inyectar una caché falsa en memoria.
type CacheRepository interface {
	// Obtener deserializa el JSON guardado en destino. Devuelve true solo en un Cache Hit.
	Obtener(ctx context.Context, clave string, destino interface{}) bool
	// Guardar serializa valor a JSON y lo almacena con el TTL indicado.
	Guardar(ctx context.Context, clave string, valor interface{}, ttl time.Duration)
	// Eliminar borra una o varias claves (invalidación de caché).
	Eliminar(ctx context.Context, claves ...string)
}

// RedisCache implementa CacheRepository sobre un cliente de Redis.
type RedisCache struct {
	cliente *redis.Client // nil = sin caché
}

// NuevaRedisCache crea la caché. Acepta un cliente nil (Redis no disponible).
func NuevaRedisCache(cliente *redis.Client) *RedisCache {
	return &RedisCache{cliente: cliente}
}

func (r *RedisCache) Obtener(ctx context.Context, clave string, destino interface{}) bool {
	if r.cliente == nil {
		return false
	}

	datos, err := r.cliente.Get(ctx, clave).Bytes()
	if errors.Is(err, redis.Nil) {
		return false // la clave no existe o ya expiró
	}
	if err != nil {
		log.Printf("Aviso: error leyendo %q de Redis: %v", clave, err)
		return false
	}

	if err := json.Unmarshal(datos, destino); err != nil {
		log.Printf("Aviso: JSON inválido en la clave %q: %v", clave, err)
		return false
	}
	return true
}

func (r *RedisCache) Guardar(ctx context.Context, clave string, valor interface{}, ttl time.Duration) {
	if r.cliente == nil {
		return
	}

	datos, err := json.Marshal(valor)
	if err != nil {
		log.Printf("Aviso: no se pudo serializar %q: %v", clave, err)
		return
	}

	if err := r.cliente.Set(ctx, clave, datos, ttl).Err(); err != nil {
		log.Printf("Aviso: no se pudo guardar %q en Redis: %v", clave, err)
	}
}

func (r *RedisCache) Eliminar(ctx context.Context, claves ...string) {
	if r.cliente == nil || len(claves) == 0 {
		return
	}
	if err := r.cliente.Del(ctx, claves...).Err(); err != nil {
		log.Printf("Aviso: no se pudieron eliminar %v de Redis: %v", claves, err)
	}
}

// Cerrar libera las conexiones del cliente.
func (r *RedisCache) Cerrar() {
	if r.cliente != nil {
		_ = r.cliente.Close()
		log.Println("Conexión a Redis cerrada")
	}
}
