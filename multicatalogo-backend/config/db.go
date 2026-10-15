// Package config crea las conexiones de infraestructura (PostgreSQL y Redis).
// No guarda variables globales: main.go recibe las conexiones y las inyecta.
package config

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// CargarEnv lee el archivo .env. Si no existe se usan las variables del sistema o los valores por defecto.
func CargarEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: no se encontró el archivo .env, se usan valores por defecto")
	}
}

// getEnv lee una variable de entorno o devuelve un valor por defecto.
func getEnv(clave, porDefecto string) string {
	if valor := os.Getenv(clave); valor != "" {
		return valor
	}
	return porDefecto
}

// GetEnvBool lee una variable de entorno booleana (true/1/si).
func GetEnvBool(clave string) bool {
	valor, err := strconv.ParseBool(getEnv(clave, "false"))
	return err == nil && valor
}

// GetEnvInt lee una variable de entorno entera o devuelve el valor por defecto.
func GetEnvInt(clave string, porDefecto int) int {
	valor, err := strconv.Atoi(getEnv(clave, ""))
	if err != nil {
		return porDefecto
	}
	return valor
}

// ConectarDB crea el pool de conexiones a PostgreSQL y verifica la conexión con un ping.
// Quien la llama es responsable de cerrar el pool (pool.Close()).
func ConectarDB() (*pgxpool.Pool, error) {
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
		return nil, fmt.Errorf("configuración de base de datos inválida: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MaxConnIdleTime = 5 * time.Minute

	ctx, cancelar := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelar()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("no se pudo crear el pool de conexiones: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("no se pudo conectar a PostgreSQL (¿está levantado el servidor?): %w", err)
	}

	log.Printf("Conectado a PostgreSQL en %s/%s", dsn.Host, getEnv("DB_NAME", "multicatalogo"))
	return pool, nil
}
