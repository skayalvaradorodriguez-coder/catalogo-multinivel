package models

// LoginRequest representa los datos enviados al iniciar sesión.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Producto representa un producto del catálogo.
type Producto struct {
	ID          int      `json:"id" db:"id"`
	Nombre      string   `json:"nombre" db:"nombre"`
	Descripcion string   `json:"descripcion" db:"descripcion"`
	Precio      float64  `json:"precio" db:"precio"`
	Categoria   string   `json:"categoria" db:"categoria"`
	Img         string   `json:"img" db:"img"`
	Galeria     []string `json:"galeria" db:"galeria"`
}

// Usuario representa una fila de la tabla usuarios.
// Password lleva json:"-" para que nunca se envíe al cliente.
type Usuario struct {
	ID       int    `json:"id" db:"id"`
	Email    string `json:"email" db:"email"`
	Password string `json:"-" db:"password"`
	Rol      string `json:"rol" db:"rol"`
}

// Referido es un nodo de la red multinivel. Hijos se arma en el repositorio
// a partir de parent_id; se omite del JSON cuando el nodo no tiene hijos.
type Referido struct {
	ID     int         `json:"id" db:"id"`
	Nombre string      `json:"nombre" db:"nombre"`
	Nivel  int         `json:"nivel" db:"nivel"`
	Ventas float64     `json:"ventas" db:"ventas"`
	Hijos  []*Referido `json:"hijos,omitempty" db:"-"`
}

// APIError estandariza los mensajes de error HTTP de la API.
// El campo Error repite Message para no romper al frontend actual,
// que lee "data.error" en el login.
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