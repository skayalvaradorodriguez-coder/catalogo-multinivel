// Package domain contiene reglas de negocio puras: sin base de datos ni HTTP, fáciles de probar.
package domain

import "math"

// TasasMLM son los porcentajes por nivel ascendente: 10% al patrocinador directo,
// 5% al siguiente y 2% al tercero.
var TasasMLM = []float64{10, 5, 2}

// Comision es una línea del reparto: nivel (1 = patrocinador directo), porcentaje y monto.
type Comision struct {
	Nivel      int
	Porcentaje float64
	Monto      float64
}

// RepartirComisiones calcula cuánto le toca a cada nivel sobre el total de una venta.
// Los montos se redondean a centavos.
func RepartirComisiones(total float64, tasas []float64) []Comision {
	if total <= 0 {
		return nil
	}
	reparto := make([]Comision, 0, len(tasas))
	for i, tasa := range tasas {
		reparto = append(reparto, Comision{
			Nivel:      i + 1,
			Porcentaje: tasa,
			Monto:      Redondear(total * tasa / 100),
		})
	}
	return reparto
}

// Redondear deja un valor monetario con 2 decimales.
func Redondear(valor float64) float64 {
	return math.Round(valor*100) / 100
}
