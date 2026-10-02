package models

// LoginRequest representa los datos enviados al iniciar sesión.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Producto representa un producto del catálogo.
type Producto struct {
	ID     int     `json:"id"`
	Nombre string  `json:"nombre"`
	Precio float64 `json:"precio"`
	Img    string  `json:"img"`
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