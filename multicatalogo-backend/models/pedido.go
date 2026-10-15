package models

import "time"

// CheckoutRequest es el JSON que envía Checkout.jsx al confirmar la compra.
type CheckoutRequest struct {
	UsuarioEmail string            `json:"usuarioEmail"`
	Cliente      DatosCliente      `json:"cliente"`
	MetodoPago   string            `json:"metodoPago"`
	Items        []ItemCheckoutReq `json:"items"`
}

// DatosCliente son los datos de envío del formulario.
type DatosCliente struct {
	Nombre    string `json:"nombre"`
	Email     string `json:"email"`
	Direccion string `json:"direccion"`
	Ciudad    string `json:"ciudad"`
	Telefono  string `json:"telefono"`
}

// ItemCheckoutReq es un producto del carrito. El precio NO se recibe del cliente:
// se toma de la base de datos para que nadie pueda alterar el total.
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

// Pedido es el pedido guardado que se devuelve al frontend.
type Pedido struct {
	ID         int          `json:"id"`
	Numero     string       `json:"numero"`
	Total      float64      `json:"total"`
	MetodoPago string       `json:"metodoPago"`
	Cliente    DatosCliente `json:"cliente"`
	Items      []ItemPedido `json:"items"`
	CreadoEn   time.Time    `json:"creadoEn"`
}

// ItemPedido es una línea del detalle con el precio tomado de PostgreSQL.
type ItemPedido struct {
	ProductoID int     `json:"productoId"`
	Nombre     string  `json:"nombre"`
	Cantidad   int     `json:"cantidad"`
	Precio     float64 `json:"precio"`
}
