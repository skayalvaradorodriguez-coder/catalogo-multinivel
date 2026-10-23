package config

import (
	"errors"
	"net"
	"net/url"
	"time"
)

// Config reúne la configuración de la aplicación leída del entorno (.env).
type Config struct {
	Puerto    string // PORT (por defecto 3000)
	Entorno   string // APP_ENV: "dev" muestra Details en los errores; cualquier otro valor los oculta
	JWTSecret string // JWT_SECRET: clave para firmar los tokens (obligatoria)
	JWTTTL    time.Duration

	CORSOrigins string // CORS_ORIGINS: orígenes permitidos separados por coma

	// StaticDir: carpeta con la build del frontend (dist). Si se define, el mismo servidor
	// entrega la app y la API en un solo dominio (así se despliega en la nube).
	StaticDir string

	SimularFallaCheckout bool          // SIMULAR_FALLA_CHECKOUT
	PausaCheckout        time.Duration // CHECKOUT_PAUSA_SEGUNDOS
}

// EsDev indica si los errores deben incluir detalles técnicos.
func (c Config) EsDev() bool { return c.Entorno == "dev" }

// CargarConfig lee las variables de entorno. Llamar después de CargarEnv.
func CargarConfig() (Config, error) {
	cfg := Config{
		Puerto:    getEnv("PORT", "3000"),
		Entorno:   getEnv("APP_ENV", "dev"),
		JWTSecret: getEnv("JWT_SECRET", ""),
		JWTTTL:    2 * time.Hour,
		StaticDir: getEnv("STATIC_DIR", ""),
		CORSOrigins: getEnv("CORS_ORIGINS", "http://localhost:5173, http://localhost:5174, http://localhost:5175, "+
			"http://127.0.0.1:5173, http://127.0.0.1:5174, http://127.0.0.1:5175"),
		SimularFallaCheckout: GetEnvBool("SIMULAR_FALLA_CHECKOUT"),
		PausaCheckout:        time.Duration(GetEnvInt("CHECKOUT_PAUSA_SEGUNDOS", 0)) * time.Second,
	}
	if len(cfg.JWTSecret) < 16 {
		return cfg, errors.New("JWT_SECRET no está definido en el .env o tiene menos de 16 caracteres")
	}
	return cfg, nil
}

// OrigenLocal indica si el origen es esta computadora o una IP de red privada (192.168.x.x, 10.x.x.x...).
func OrigenLocal(origen string) bool {
	u, err := url.Parse(origen)
	if err != nil {
		return false
	}
	host := u.Hostname()
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && (ip.IsLoopback() || ip.IsPrivate())
}
