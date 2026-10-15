// Package repository es la capa de acceso a datos: aquí vive TODO el SQL.
// Los controladores dependen de la interfaz DBRepository, no de PostgreSQL directamente.
package repository

import (
	"context"
	"errors"
	"time"

	"multicatalogo-backend/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBRepository es el contrato de la capa de datos que se inyecta en los controladores.
type DBRepository interface {
	BuscarUsuarioPorCredenciales(ctx context.Context, email, password string) (*models.Usuario, error)
	BuscarUsuarioPorEmail(ctx context.Context, email string) (*models.Usuario, error)
	ObtenerProductos(ctx context.Context) ([]models.Producto, error)
	ObtenerProductoPorID(ctx context.Context, id int) (*models.Producto, error)
	ObtenerRed(ctx context.Context) (*models.Referido, error)
	// CrearPedido guarda el pedido y su detalle en UNA transacción ACID.
	CrearPedido(ctx context.Context, pedido models.NuevoPedido) (*models.Pedido, error)
	// CalcularComisiones suma la venta al comprador y paga comisiones a sus 3 niveles ascendentes.
	// Devuelve los emails cuyo árbol cambió, para invalidar su caché.
	CalcularComisiones(ctx context.Context, pedidoID, usuarioID int, total float64) ([]string, error)
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

const columnasProducto = "id, nombre, descripcion, precio::float8, categoria, img, galeria, stock"

// BuscarUsuarioPorCredenciales devuelve el usuario que coincide con email y password.
// Si no existe devuelve (nil, nil); el error queda reservado para fallos de la base de datos.
func (r *PostgresRepository) BuscarUsuarioPorCredenciales(ctx context.Context, email, password string) (*models.Usuario, error) {
	return r.buscarUsuario(ctx,
		"SELECT id, email, password, rol FROM usuarios WHERE email = $1 AND password = $2", email, password)
}

// BuscarUsuarioPorEmail devuelve el usuario con ese email, o (nil, nil) si no existe.
func (r *PostgresRepository) BuscarUsuarioPorEmail(ctx context.Context, email string) (*models.Usuario, error) {
	return r.buscarUsuario(ctx, "SELECT id, email, password, rol FROM usuarios WHERE email = $1", email)
}

func (r *PostgresRepository) buscarUsuario(ctx context.Context, sql string, args ...interface{}) (*models.Usuario, error) {
	var u models.Usuario
	err := r.db.QueryRow(ctx, sql, args...).Scan(&u.ID, &u.Email, &u.Password, &u.Rol)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// ObtenerProductos devuelve todo el catálogo ordenado por id.
func (r *PostgresRepository) ObtenerProductos(ctx context.Context) ([]models.Producto, error) {
	filas, err := r.db.Query(ctx, "SELECT "+columnasProducto+" FROM productos ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer filas.Close()

	// Slice no nulo: si la tabla está vacía el JSON es [] y no null.
	productos := []models.Producto{}
	for filas.Next() {
		var p models.Producto
		if err := filas.Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Categoria, &p.Img, &p.Galeria, &p.Stock); err != nil {
			return nil, err
		}
		productos = append(productos, p)
	}
	return productos, filas.Err()
}

// ObtenerProductoPorID devuelve un producto, o (nil, nil) si el id no existe.
func (r *PostgresRepository) ObtenerProductoPorID(ctx context.Context, id int) (*models.Producto, error) {
	var p models.Producto
	err := r.db.QueryRow(ctx,
		"SELECT "+columnasProducto+" FROM productos WHERE id = $1", id,
	).Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Categoria, &p.Img, &p.Galeria, &p.Stock)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// ObtenerRed lee la tabla referidos como lista plana y reconstruye el árbol en memoria.
// Devuelve el nodo raíz (parent_id NULL), o (nil, nil) si la tabla está vacía.
func (r *PostgresRepository) ObtenerRed(ctx context.Context) (*models.Referido, error) {
	filas, err := r.db.Query(ctx,
		"SELECT id, nombre, nivel, ventas::float8, comisiones::float8, parent_id FROM referidos ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer filas.Close()

	nodos := map[int]*models.Referido{} // id -> nodo
	padres := map[int]*int{}            // id -> parent_id (nil en la raíz)
	var orden []int                     // conserva el orden por id

	for filas.Next() {
		nodo := &models.Referido{}
		var parentID *int
		if err := filas.Scan(&nodo.ID, &nodo.Nombre, &nodo.Nivel, &nodo.Ventas, &nodo.Comisiones, &parentID); err != nil {
			return nil, err
		}
		nodos[nodo.ID] = nodo
		padres[nodo.ID] = parentID
		orden = append(orden, nodo.ID)
	}
	if err := filas.Err(); err != nil {
		return nil, err
	}

	// Segunda pasada: colgamos cada nodo de su padre.
	var raiz *models.Referido
	for _, id := range orden {
		parentID := padres[id]
		if parentID == nil {
			if raiz == nil {
				raiz = nodos[id]
			}
			continue
		}
		if padre, existe := nodos[*parentID]; existe {
			padre.Hijos = append(padre.Hijos, nodos[id])
		}
	}
	return raiz, nil
}
