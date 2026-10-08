// Package config centraliza la configuración de infraestructura (base de datos).
package config

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// DB es el pool de conexiones a PostgreSQL que usa la capa repository.
var DB *pgxpool.Pool

// getEnv lee una variable de entorno o devuelve un valor por defecto.
func getEnv(clave, porDefecto string) string {
	if valor := os.Getenv(clave); valor != "" {
		return valor
	}
	return porDefecto
}

// ConectarDB carga el archivo .env, crea el pool y verifica la conexión con un ping.
func ConectarDB() error {
	// Si no existe .env no es un error: se usan las variables del sistema o los valores por defecto.
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: no se encontró el archivo .env, se usan valores por defecto")
	}

	// Armamos la URL con net/url para que usuario y contraseña queden bien escapados.
	dsn := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(getEnv("DB_USER", "postgres"), getEnv("DB_PASSWORD", "postgres")),
		Host:     fmt.Sprintf("%s:%s", getEnv("DB_HOST", "localhost"), getEnv("DB_PORT", "5432")),
		Path:     getEnv("DB_NAME", "multicatalogo"),
		RawQuery: "sslmode=disable",
	}

	cfg, err := pgxpool.ParseConfig(dsn.String())
	if err != nil {
		return fmt.Errorf("configuración de base de datos inválida: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MaxConnIdleTime = 5 * time.Minute

	ctx, cancelar := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelar()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("no se pudo crear el pool de conexiones: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("no se pudo conectar a PostgreSQL (¿está levantado el contenedor?): %w", err)
	}

	DB = pool
	log.Printf("Conectado a PostgreSQL en %s/%s", dsn.Host, getEnv("DB_NAME", "multicatalogo"))
	return nil
}

// CerrarDB libera todas las conexiones del pool. Se llama al finalizar la aplicación.
func CerrarDB() {
	if DB != nil {
		DB.Close()
		log.Println("Conexión a PostgreSQL cerrada")
	}
}
