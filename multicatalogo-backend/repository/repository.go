// Package repository es la capa de acceso a datos: aquí vive TODO el SQL.
// Los controladores dependen de la interfaz DBRepository, no de PostgreSQL directamente.
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"multicatalogo-backend/models"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Errores de dominio que los controladores traducen a códigos HTTP.
var (
	ErrNoEncontrado     = errors.New("no encontrado")                       // 404 (401 en login)
	ErrDuplicado        = errors.New("ya existe")                           // 409
	ErrCantidadInvalida = errors.New("la cantidad debe ser mayor que cero") // 400
)

// ErrValidacion indica un problema con los datos (producto inexistente, stock insuficiente,
// transición de estado no permitida...). El controlador lo traduce a HTTP 400.
type ErrValidacion struct {
	Mensaje string
}

func (e *ErrValidacion) Error() string { return e.Mensaje }

func errValidacion(formato string, args ...interface{}) error {
	return &ErrValidacion{Mensaje: fmt.Sprintf(formato, args...)}
}

// DBRepository es el contrato de la capa de datos que se inyecta en los controladores.
type DBRepository interface {
	Ping(ctx context.Context) error

	// Usuarios
	BuscarUsuarioPorEmail(ctx context.Context, email string) (*models.Usuario, error)
	BuscarUsuarioPorID(ctx context.Context, id int) (*models.Usuario, error)
	RegistrarUsuario(ctx context.Context, nombre, email, hash, ref string) (*models.Usuario, error)
	ListarUsuarios(ctx context.Context, q string, page, limit int) (models.Pagina[models.Usuario], error)
	ActualizarUsuario(ctx context.Context, id int, cambios models.CambiosUsuario) (*models.Usuario, error)

	// Productos
	ObtenerProductos(ctx context.Context, q string) ([]models.Producto, error)
	ObtenerProductoPorID(ctx context.Context, id int, incluirInactivos bool) (*models.Producto, error)
	ListarProductosAdmin(ctx context.Context, q string, page, limit int) (models.Pagina[models.Producto], error)
	CrearProducto(ctx context.Context, p models.Producto) (*models.Producto, error)
	ActualizarProducto(ctx context.Context, id int, cambios models.ProductoInput) (*models.Producto, error)
	DesactivarProducto(ctx context.Context, id int) error

	// Red multinivel
	ObtenerRed(ctx context.Context) (*models.Referido, error)
	ObtenerSubRed(ctx context.Context, referidoID int) (*models.Referido, error)

	// Pedidos
	CrearPedido(ctx context.Context, pedido models.NuevoPedido) (*models.Pedido, error)
	SumarVentas(ctx context.Context, usuarioID int, monto float64) error
	ListarPedidos(ctx context.Context, filtro models.FiltroPedidos) (models.Pagina[models.PedidoResumen], error)
	ObtenerPedido(ctx context.Context, id int) (*models.Pedido, error)
	CambiarEstadoPedido(ctx context.Context, id int, nuevo string) (*models.Pedido, error)

	// Comisiones y liquidaciones
	EstadoCuenta(ctx context.Context, referidoID int) (*models.EstadoCuenta, error)
	Liquidar(ctx context.Context, referidoID int) (*models.Liquidacion, error)
	ListarLiquidaciones(ctx context.Context, referidoID *int, page, limit int) (models.Pagina[models.Liquidacion], error)

	// Reportes
	ResumenVentas(ctx context.Context, desde, hasta *time.Time) (*models.ResumenVentas, error)
	TopProductos(ctx context.Context, limit int) ([]models.TopProducto, error)
	TopReferidos(ctx context.Context, limit int) ([]models.TopReferido, error)
	ComisionesPendientes(ctx context.Context) ([]models.DeudaComisiones, error)
}

// OpcionesCheckout permite provocar fallas a propósito para la demostración de ACID en clase.
type OpcionesCheckout struct {
	SimularFalla bool          // corta la transacción después de insertar el pedido
	Pausa        time.Duration // espera antes del COMMIT (para detener PostgreSQL a mitad)
}

// PostgresRepository implementa DBRepository con un pool de PostgreSQL.
type PostgresRepository struct {
	db       *pgxpool.Pool
	opciones OpcionesCheckout
}

// NuevoPostgresRepository crea el repositorio con sus dependencias.
func NuevoPostgresRepository(db *pgxpool.Pool, opciones OpcionesCheckout) *PostgresRepository {
	return &PostgresRepository{db: db, opciones: opciones}
}

// Ping verifica que PostgreSQL responda (lo usa /api/health).
func (r *PostgresRepository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

// codigoPG devuelve el código de error de PostgreSQL ("23505" = duplicado, "23503" = FK...).
func codigoPG(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

// paginar normaliza page/limit y calcula el OFFSET.
func paginar(page, limit int) (int, int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit, (page - 1) * limit
}
