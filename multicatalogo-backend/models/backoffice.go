package models

import "time"

// Comision es una fila del ledger de comisiones.
type Comision struct {
	ID            int       `json:"id"`
	PedidoID      int       `json:"pedidoId"`
	PedidoNumero  string    `json:"pedidoNumero"`
	ReferidoID    int       `json:"referidoId"`
	Nivel         int       `json:"nivel"`
	Porcentaje    float64   `json:"porcentaje"`
	Monto         float64   `json:"monto"`
	Estado        string    `json:"estado"`
	LiquidacionID *int      `json:"liquidacionId"`
	CreadoEn      time.Time `json:"creadoEn"`
}

// EstadoCuenta es el ledger de un distribuidor con sus totales por estado.
type EstadoCuenta struct {
	Comisiones []Comision         `json:"comisiones"`
	Totales    map[string]float64 `json:"totales"` // pendiente, pagada, revertida
}

// Liquidacion es un pago realizado a un distribuidor.
type Liquidacion struct {
	ID             int        `json:"id"`
	ReferidoID     int        `json:"referidoId"`
	ReferidoNombre string     `json:"referidoNombre"`
	Total          float64    `json:"total"`
	Estado         string     `json:"estado"`
	Comisiones     int        `json:"comisiones"`
	CreadoEn       time.Time  `json:"creadoEn"`
	PagadaEn       *time.Time `json:"pagadaEn"`
}

// ResumenVentas es el KPI principal del Dashboard.
type ResumenVentas struct {
	Desde                *string `json:"desde"`
	Hasta                *string `json:"hasta"`
	TotalVendido         float64 `json:"totalVendido"`
	NumPedidos           int     `json:"numPedidos"`
	TicketPromedio       float64 `json:"ticketPromedio"`
	ComisionesDevengadas float64 `json:"comisionesDevengadas"`
	ComisionesPendientes float64 `json:"comisionesPendientes"`
	ComisionesPagadas    float64 `json:"comisionesPagadas"`
	PedidosCancelados    int     `json:"pedidosCancelados"`
}

// TopProducto es una fila del ranking de productos.
type TopProducto struct {
	ProductoID int     `json:"productoId"`
	Nombre     string  `json:"nombre"`
	Unidades   int     `json:"unidades"`
	Ingresos   float64 `json:"ingresos"`
}

// TopReferido es una fila del ranking de distribuidores.
type TopReferido struct {
	ReferidoID int     `json:"referidoId"`
	Nombre     string  `json:"nombre"`
	Codigo     string  `json:"codigo"`
	Nivel      int     `json:"nivel"`
	Pedidos    int     `json:"pedidos"`
	Ventas     float64 `json:"ventas"`
	Comisiones float64 `json:"comisiones"`
}

// DeudaComisiones es lo que se le debe a un distribuidor.
type DeudaComisiones struct {
	ReferidoID int     `json:"referidoId"`
	Nombre     string  `json:"nombre"`
	Codigo     string  `json:"codigo"`
	Email      *string `json:"email"`
	Comisiones int     `json:"comisiones"`
	Pendiente  float64 `json:"pendiente"`
}
