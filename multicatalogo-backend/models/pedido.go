package models

import "time"

// CheckoutRequest es el JSON que envía Checkout.jsx. La identidad del comprador
// NO viaja en el cuerpo: se toma del token JWT.
type CheckoutRequest struct {
	Cliente    DatosCliente      `json:"cliente"`
	MetodoPago string            `json:"metodoPago"`
	Items      []ItemCheckoutReq `json:"items"`
}

// DatosCliente son los datos de envío del formulario.
type DatosCliente struct {
	Nombre    string `json:"nombre"`
	Email     string `json:"email"`
	Direccion string `json:"direccion"`
	Ciudad    string `json:"ciudad"`
	Telefono  string `json:"telefono"`
}

// ItemCheckoutReq es un producto del carrito. El precio NO se recibe del cliente.
type ItemCheckoutReq struct {
	ProductoID int `json:"id"`
	Cantidad   int `json:"cantidad"`
}

// NuevoPedido es lo que el controlador entrega al repositorio ya validado.
type NuevoPedido struct {
	UsuarioID  int
	Cliente    DatosCliente
	MetodoPago string
	Items      []ItemCheckoutReq
}

// Pedido es la cabecera + detalle de un pedido.
type Pedido struct {
	ID         int          `json:"id"`
	Numero     string       `json:"numero"`
	UsuarioID  int          `json:"usuarioId"`
	Total      float64      `json:"total"`
	Estado     string       `json:"estado"`
	MetodoPago string       `json:"metodoPago"`
	Cliente    DatosCliente `json:"cliente"`
	Items      []ItemPedido `json:"items"`
	CreadoEn   time.Time    `json:"creadoEn"`
	// Siguientes son los estados a los que puede pasar (solo informativo para el admin).
	Siguientes []string `json:"siguientes,omitempty"`
}

// PedidoResumen es una fila de los listados de pedidos.
type PedidoResumen struct {
	ID            int       `json:"id"`
	Numero        string    `json:"numero"`
	Total         float64   `json:"total"`
	Estado        string    `json:"estado"`
	ClienteNombre string    `json:"clienteNombre"`
	ClienteEmail  string    `json:"clienteEmail"`
	Unidades      int       `json:"unidades"`
	CreadoEn      time.Time `json:"creadoEn"`
	Siguientes    []string  `json:"siguientes"`
}

// ItemPedido es una línea del detalle con el precio tomado de PostgreSQL.
type ItemPedido struct {
	ProductoID int     `json:"productoId"`
	Nombre     string  `json:"nombre"`
	Cantidad   int     `json:"cantidad"`
	Precio     float64 `json:"precio"`
}

// FiltroPedidos parametriza los listados paginados.
type FiltroPedidos struct {
	UsuarioID *int   // nil = todos (admin)
	Estado    string // "" = todos
	Page      int
	Limit     int
}
