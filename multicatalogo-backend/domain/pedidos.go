package domain

// Estados posibles de un pedido.
const (
	EstadoPendiente  = "pendiente"
	EstadoPreparando = "preparando"
	EstadoEnviado    = "enviado"
	EstadoEntregado  = "entregado"
	EstadoCancelado  = "cancelado"
)

// transiciones define el flujo: pendiente → preparando → enviado → entregado.
// Se puede cancelar mientras el pedido no haya salido (pendiente o preparando).
var transiciones = map[string][]string{
	EstadoPendiente:  {EstadoPreparando, EstadoCancelado},
	EstadoPreparando: {EstadoEnviado, EstadoCancelado},
	EstadoEnviado:    {EstadoEntregado},
}

// EstadoValido indica si el texto es uno de los estados conocidos.
func EstadoValido(estado string) bool {
	switch estado {
	case EstadoPendiente, EstadoPreparando, EstadoEnviado, EstadoEntregado, EstadoCancelado:
		return true
	}
	return false
}

// TransicionValida indica si un pedido puede pasar del estado actual al nuevo.
func TransicionValida(actual, nuevo string) bool {
	for _, permitido := range transiciones[actual] {
		if permitido == nuevo {
			return true
		}
	}
	return false
}

// SiguientesEstados devuelve a qué estados puede pasar un pedido (para la interfaz del admin).
func SiguientesEstados(actual string) []string {
	return append([]string{}, transiciones[actual]...)
}
