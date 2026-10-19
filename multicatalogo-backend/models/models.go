// Package models define las estructuras que viajan entre capas y en el JSON de la API.
package models

import "time"

// LoginRequest representa los datos enviados al iniciar sesión.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegistroRequest es el formulario de /registro. Ref es el código de quien invita (opcional).
type RegistroRequest struct {
	Nombre   string `json:"nombre"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Ref      string `json:"ref"`
}

// RespuestaSesion es lo que devuelven login y registro.
type RespuestaSesion struct {
	Token  string `json:"token"`
	Email  string `json:"email"`
	Rol    string `json:"rol"`
	Nombre string `json:"nombre"`
}

// Usuario representa una fila de la tabla usuarios.
// Password lleva json:"-" para que nunca se envíe al cliente.
type Usuario struct {
	ID         int       `json:"id"`
	Email      string    `json:"email"`
	Password   string    `json:"-"`
	Rol        string    `json:"rol"`
	Nombre     string    `json:"nombre"`
	Activo     bool      `json:"activo"`
	ReferidoID *int      `json:"referidoId"`
	Codigo     string    `json:"codigo,omitempty"` // código de invitación de su nodo
	CreadoEn   time.Time `json:"creadoEn"`
}

// CambiosUsuario es el cuerpo de PATCH /api/usuarios/:id (campos opcionales).
type CambiosUsuario struct {
	Rol    *string `json:"rol"`
	Activo *bool   `json:"activo"`
}

// Producto representa un producto del catálogo.
type Producto struct {
	ID          int      `json:"id"`
	Nombre      string   `json:"nombre"`
	Descripcion string   `json:"descripcion"`
	Precio      float64  `json:"precio"`
	Categoria   string   `json:"categoria"`
	Img         string   `json:"img"`
	Galeria     []string `json:"galeria"`
	Stock       int      `json:"stock"`
	Activo      bool     `json:"activo"`
}

// ProductoInput es el cuerpo de POST/PUT /api/productos. En PUT todos los campos son opcionales.
type ProductoInput struct {
	Nombre      *string   `json:"nombre"`
	Descripcion *string   `json:"descripcion"`
	Precio      *float64  `json:"precio"`
	Categoria   *string   `json:"categoria"`
	Img         *string   `json:"img"`
	Galeria     *[]string `json:"galeria"`
	Stock       *int      `json:"stock"`
	Activo      *bool     `json:"activo"`
}

// Referido es un nodo de la red multinivel. Hijos se arma en el repositorio a partir de parent_id.
type Referido struct {
	ID     int         `json:"id"`
	Nombre string      `json:"nombre"`
	Nivel  int         `json:"nivel"`
	Ventas float64     `json:"ventas"`
	Codigo string      `json:"codigo,omitempty"`
	Hijos  []*Referido `json:"hijos,omitempty"`
}

// Pagina es la respuesta estándar de los listados paginados.
type Pagina[T any] struct {
	Data  []T `json:"data"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

// APIError estandariza los mensajes de error HTTP de la API.
// El campo Error repite Message para no romper al frontend anterior.
type APIError struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Error   string      `json:"error"`
	Details interface{} `json:"details,omitempty"`
}

// NewAPIError construye un APIError con Error sincronizado con Message.
func NewAPIError(status int, message string, details interface{}) APIError {
	return APIError{Status: status, Message: message, Error: message, Details: details}
}
