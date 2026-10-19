package seguridad

import (
	"strings"
	"testing"
	"time"
)

const secreto = "clave-de-prueba-1234567890"

func TestTokenValido(t *testing.T) {
	token, err := GenerarToken(secreto, 7, "ana@mail.com", "cliente", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := ValidarToken(secreto, token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UsuarioID() != 7 || claims.Email != "ana@mail.com" || claims.Rol != "cliente" || claims.EsAdmin() {
		t.Errorf("claims incorrectos: %+v", claims)
	}
}

func TestTokenAdulterado(t *testing.T) {
	token, _ := GenerarToken(secreto, 7, "ana@mail.com", "cliente", time.Hour)
	partes := strings.Split(token, ".")
	partes[2] = strings.Repeat("A", len(partes[2])) // firma falsa
	if _, err := ValidarToken(secreto, strings.Join(partes, ".")); err == nil {
		t.Error("un token con la firma alterada no debe ser válido")
	}
}

func TestTokenOtraClave(t *testing.T) {
	token, _ := GenerarToken("otra-clave-distinta-123456", 7, "ana@mail.com", "admin", time.Hour)
	if _, err := ValidarToken(secreto, token); err == nil {
		t.Error("un token firmado con otra clave no debe ser válido")
	}
}

func TestTokenVencido(t *testing.T) {
	token, _ := GenerarToken(secreto, 7, "ana@mail.com", "cliente", -time.Minute)
	if _, err := ValidarToken(secreto, token); err == nil {
		t.Error("un token vencido no debe ser válido")
	}
}

func TestPassword(t *testing.T) {
	hash, err := HashPassword("123456")
	if err != nil {
		t.Fatal(err)
	}
	if !VerificarPassword(hash, "123456") || VerificarPassword(hash, "654321") {
		t.Error("la verificación de bcrypt no funciona")
	}

}
