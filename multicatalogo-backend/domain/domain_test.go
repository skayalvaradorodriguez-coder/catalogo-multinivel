package domain

import "testing"

func TestRepartirComisiones(t *testing.T) {
	reparto := RepartirComisiones(264, TasasMLM)
	esperado := []Comision{{1, 10, 26.40}, {2, 5, 13.20}, {3, 2, 5.28}}
	if len(reparto) != len(esperado) {
		t.Fatalf("se esperaban %d comisiones, hay %d", len(esperado), len(reparto))
	}
	for i := range esperado {
		if reparto[i] != esperado[i] {
			t.Errorf("nivel %d: esperado %+v, obtenido %+v", i+1, esperado[i], reparto[i])
		}
	}
}

func TestRepartirComisionesRedondea(t *testing.T) {
	reparto := RepartirComisiones(32.5, TasasMLM)
	if reparto[0].Monto != 3.25 || reparto[1].Monto != 1.63 || reparto[2].Monto != 0.65 {
		t.Errorf("redondeo incorrecto: %+v", reparto)
	}
}

func TestRepartirComisionesTotalCero(t *testing.T) {
	if reparto := RepartirComisiones(0, TasasMLM); len(reparto) != 0 {
		t.Errorf("un total de 0 no debe generar comisiones: %+v", reparto)
	}
}

func TestTransiciones(t *testing.T) {
	casos := []struct {
		actual, nuevo string
		valida        bool
	}{
		{EstadoPendiente, EstadoPreparando, true},
		{EstadoPreparando, EstadoEnviado, true},
		{EstadoEnviado, EstadoEntregado, true},
		{EstadoPendiente, EstadoCancelado, true},
		{EstadoEntregado, EstadoPendiente, false},
		{EstadoEnviado, EstadoCancelado, false},
		{EstadoCancelado, EstadoPendiente, false},
		{EstadoPendiente, EstadoEntregado, false},
	}
	for _, caso := range casos {
		if got := TransicionValida(caso.actual, caso.nuevo); got != caso.valida {
			t.Errorf("%s → %s: esperado %v, obtenido %v", caso.actual, caso.nuevo, caso.valida, got)
		}
	}
}
