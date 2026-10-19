// Package validacion centraliza las reglas de entrada que antes cada controlador repetía a mano.
package validacion

import (
	"net/mail"
	"strings"
)

// Errores acumula los problemas por campo; se envía al cliente en APIError.details.
type Errores map[string]string

// Requerido marca el campo si el texto está vacío.
func (e Errores) Requerido(campo, valor, mensaje string) {
	if strings.TrimSpace(valor) == "" {
		e[campo] = mensaje
	}
}

// Email marca el campo si el texto no es un correo válido.
func (e Errores) Email(campo, valor string) {
	if _, err := mail.ParseAddress(valor); err != nil || !strings.Contains(valor, "@") {
		e[campo] = "El correo electrónico no es válido"
	}
}

// MinLongitud marca el campo si el texto es más corto que min caracteres.
func (e Errores) MinLongitud(campo, valor string, min int, mensaje string) {
	if len([]rune(valor)) < min {
		e[campo] = mensaje
	}
}

// Positivo marca el campo si el número no es mayor que cero.
func (e Errores) Positivo(campo string, valor float64, mensaje string) {
	if valor <= 0 {
		e[campo] = mensaje
	}
}

// NoNegativo marca el campo si el número es menor que cero.
func (e Errores) NoNegativo(campo string, valor int, mensaje string) {
	if valor < 0 {
		e[campo] = mensaje
	}
}

// Validar devuelve nil si no hubo errores, para usarlo como "if errs := v.Validar(); errs != nil".
func (e Errores) Validar() map[string]string {
	if len(e) == 0 {
		return nil
	}
	return e
}
