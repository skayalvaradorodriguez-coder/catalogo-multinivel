// Package cache implementa las operaciones del patrón Cache-Aside sobre Redis.
// Si Redis no está disponible, Obtener siempre reporta "miss" y Guardar no hace nada,
// así la API sigue respondiendo desde PostgreSQL.
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"multicatalogo-backend/config"

	"github.com/redis/go-redis/v9"
)

// Claves y tiempos de expiración (TTL) definidos en la Fase 3.
const (
	ClaveProductos = "productos:all"
	TTLProductos   = 5 * time.Minute

	PrefijoRed = "red:tree:" // + email del usuario
	TTLRed     = 10 * time.Minute
)

// Obtener busca la clave en Redis y deserializa el JSON en destino.
// Devuelve true solo en un Cache Hit.
func Obtener(ctx context.Context, clave string, destino interface{}) bool {
	if config.Redis == nil {
		return false
	}

	datos, err := config.Redis.Get(ctx, clave).Bytes()
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

// Guardar serializa valor a JSON y lo almacena en Redis con el TTL indicado.
func Guardar(ctx context.Context, clave string, valor interface{}, ttl time.Duration) {
	if config.Redis == nil {
		return
	}

	datos, err := json.Marshal(valor)
	if err != nil {
		log.Printf("Aviso: no se pudo serializar %q: %v", clave, err)
		return
	}

	if err := config.Redis.Set(ctx, clave, datos, ttl).Err(); err != nil {
		log.Printf("Aviso: no se pudo guardar %q en Redis: %v", clave, err)
	}
}
