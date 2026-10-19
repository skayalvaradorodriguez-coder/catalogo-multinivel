package controllers

import (
	"errors"
	"strings"
	"time"

	"multicatalogo-backend/cache"
	"multicatalogo-backend/models"
	"multicatalogo-backend/notificaciones"
	"multicatalogo-backend/repository"
	"multicatalogo-backend/seguridad"
	"multicatalogo-backend/validacion"

	"github.com/gofiber/fiber/v2"
)

// AuthController maneja inicio de sesión y registro.
type AuthController struct {
	base
	repo    repository.DBRepository
	cache   cache.CacheRepository
	mailer  notificaciones.Mailer
	secreto string
	ttl     time.Duration
}

// Login valida las credenciales con bcrypt y devuelve un JWT real.
func (ctl *AuthController) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return ctl.errorJSON(c, fiber.StatusBadRequest, "Cuerpo de petición inválido", nil)
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := strings.TrimSpace(req.Password)

	v := validacion.Errores{}
	v.Requerido("email", email, "El email es obligatorio")
	v.Requerido("password", password, "La contraseña es obligatoria")
	if errs := v.Validar(); errs != nil {
		return ctl.errorJSON(c, fiber.StatusBadRequest, "Email y contraseña son obligatorios", errs)
	}

	usuario, err := ctl.repo.BuscarUsuarioPorEmail(c.Context(), email)
	if errors.Is(err, repository.ErrNoEncontrado) || (err == nil && !seguridad.VerificarPassword(usuario.Password, password)) {
		// Mismo mensaje en ambos casos: no revelamos si el email existe.
		return ctl.errorJSON(c, fiber.StatusUnauthorized, "Credenciales incorrectas", nil)
	}
	if err != nil {
		return ctl.errorInterno(c, err)
	}
	if !usuario.Activo {
		return ctl.errorJSON(c, fiber.StatusForbidden, "Tu cuenta está desactivada, comunícate con el administrador", nil)
	}

	return ctl.responderSesion(c, fiber.StatusOK, usuario)
}

// Register crea una cuenta de cliente y su nodo en la red (con código de referido opcional).
func (ctl *AuthController) Register(c *fiber.Ctx) error {
	var req models.RegistroRequest
	if err := c.BodyParser(&req); err != nil {
		return ctl.errorJSON(c, fiber.StatusBadRequest, "Cuerpo de petición inválido", nil)
	}
	req.Nombre = strings.Join(strings.Fields(req.Nombre), " ")
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Ref = strings.TrimSpace(req.Ref)

	v := validacion.Errores{}
	v.Requerido("nombre", req.Nombre, "El nombre es obligatorio")
	v.Email("email", req.Email)
	v.MinLongitud("password", req.Password, 6, "La contraseña debe tener al menos 6 caracteres")
	if errs := v.Validar(); errs != nil {
		return ctl.errorJSON(c, fiber.StatusBadRequest, "Revisa los datos del registro", errs)
	}

	hash, err := seguridad.HashPassword(req.Password)
	if err != nil {
		return ctl.errorInterno(c, err)
	}

	usuario, err := ctl.repo.RegistrarUsuario(c.Context(), req.Nombre, req.Email, hash, req.Ref)
	if errors.Is(err, repository.ErrDuplicado) {
		return ctl.errorJSON(c, fiber.StatusConflict, "Ya existe una cuenta con ese email", nil)
	}
	if err != nil {
		return ctl.errorRepo(c, err, "No se pudo registrar")
	}

	// La red cambió: los árboles en caché ya no son válidos.
	ctl.cache.EliminarPatron(c.Context(), cache.PatronRed)
	_ = ctl.mailer.Enviar(usuario.Email, "Bienvenida a MultiCatálogo",
		"Tu código para invitar a otros es "+usuario.Codigo)

	return ctl.responderSesion(c, fiber.StatusCreated, usuario)
}

func (ctl *AuthController) responderSesion(c *fiber.Ctx, status int, u *models.Usuario) error {
	token, err := seguridad.GenerarToken(ctl.secreto, u.ID, u.Email, u.Rol, ctl.ttl)
	if err != nil {
		return ctl.errorInterno(c, err)
	}
	return c.Status(status).JSON(models.RespuestaSesion{Token: token, Email: u.Email, Rol: u.Rol, Nombre: u.Nombre})
}
