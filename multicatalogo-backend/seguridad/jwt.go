package seguridad

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims es el contenido del token: {sub, email, rol, iat, exp}.
type Claims struct {
	Email string `json:"email"`
	Rol   string `json:"rol"`
	jwt.RegisteredClaims
}

// UsuarioID devuelve el id del usuario guardado en "sub".
func (c *Claims) UsuarioID() int {
	id, _ := strconv.Atoi(c.Subject)
	return id
}

// EsAdmin indica si el token pertenece a un administrador.
func (c *Claims) EsAdmin() bool { return c.Rol == "admin" }

// GenerarToken firma un JWT HS256 que vence en ttl.
func GenerarToken(secreto string, usuarioID int, email, rol string, ttl time.Duration) (string, error) {
	ahora := time.Now()
	claims := Claims{
		Email: email,
		Rol:   rol,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(usuarioID),
			IssuedAt:  jwt.NewNumericDate(ahora),
			ExpiresAt: jwt.NewNumericDate(ahora.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secreto))
}

// ValidarToken verifica firma, algoritmo y vencimiento, y devuelve los claims.
func ValidarToken(secreto, token string) (*Claims, error) {
	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secreto), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())
	if err != nil {
		return nil, err
	}
	if !parsed.Valid || claims.UsuarioID() <= 0 {
		return nil, errors.New("token inválido")
	}
	return claims, nil
}
