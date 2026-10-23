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
// Si existe DATABASE_URL (la cadena que dan Neon, Render, Supabase...) se usa esa; si no, se
// arma con DB_HOST, DB_PORT, DB_USER, DB_PASSWORD y DB_NAME.
// Quien la llama es responsable de cerrar el pool (pool.Close()).
func ConectarDB() (*pgxpool.Pool, error) {
	cadena, destino := cadenaConexion()

	cfg, err := pgxpool.ParseConfig(cadena)
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

	log.Printf("Conectado a PostgreSQL en %s", destino)
	return pool, nil
}

// cadenaConexion devuelve la URL de conexión y una descripción sin contraseña para el log.
func cadenaConexion() (string, string) {
	if valor := os.Getenv("DATABASE_URL"); valor != "" {
		u, err := url.Parse(valor)
		if err == nil {
			// Neon agrega channel_binding=require; se quita para no depender de que el driver lo soporte
			// (la conexión sigue cifrada con sslmode=require).
			q := u.Query()
			q.Del("channel_binding")
			if q.Get("sslmode") == "" {
				q.Set("sslmode", "require")
			}
			u.RawQuery = q.Encode()
			return u.String(), u.Host + u.Path
		}
		return valor, "DATABASE_URL"
	}

	// Armamos la URL con net/url para que usuario y contraseña queden bien escapados.
	dsn := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(getEnv("DB_USER", "postgres"), getEnv("DB_PASSWORD", "postgres")),
		Host:     fmt.Sprintf("%s:%s", getEnv("DB_HOST", "localhost"), getEnv("DB_PORT", "5432")),
		Path:     getEnv("DB_NAME", "multicatalogo"),
		RawQuery: "sslmode=" + getEnv("DB_SSLMODE", "disable"),
	}
	return dsn.String(), dsn.Host + "/" + getEnv("DB_NAME", "multicatalogo")
}
