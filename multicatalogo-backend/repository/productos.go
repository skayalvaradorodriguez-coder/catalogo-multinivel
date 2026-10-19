package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"multicatalogo-backend/models"

	"github.com/jackc/pgx/v5"
)

const columnasProducto = "id, nombre, descripcion, precio::float8, categoria, img, galeria, stock, activo"

func escanearProducto(fila pgx.Row) (*models.Producto, error) {
	var p models.Producto
	err := fila.Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Categoria, &p.Img, &p.Galeria, &p.Stock, &p.Activo)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoEncontrado
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PostgresRepository) listarProductos(ctx context.Context, sql string, args ...interface{}) ([]models.Producto, error) {
	filas, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer filas.Close()

	// Slice no nulo: si no hay filas el JSON es [] y no null.
	productos := []models.Producto{}
	for filas.Next() {
		p, err := escanearProducto(filas)
		if err != nil {
			return nil, err
		}
		productos = append(productos, *p)
	}
	return productos, filas.Err()
}

// ObtenerProductos devuelve el catálogo público (solo activos). Con q hace búsqueda ILIKE.
func (r *PostgresRepository) ObtenerProductos(ctx context.Context, q string) ([]models.Producto, error) {
	if q = strings.TrimSpace(q); q != "" {
		return r.listarProductos(ctx, "SELECT "+columnasProducto+` FROM productos
			WHERE activo AND (nombre ILIKE $1 OR descripcion ILIKE $1 OR categoria ILIKE $1)
			ORDER BY id`, "%"+q+"%")
	}
	return r.listarProductos(ctx, "SELECT "+columnasProducto+" FROM productos WHERE activo ORDER BY id")
}

// ObtenerProductoPorID devuelve un producto o ErrNoEncontrado. Los inactivos solo los ve el admin.
func (r *PostgresRepository) ObtenerProductoPorID(ctx context.Context, id int, incluirInactivos bool) (*models.Producto, error) {
	return escanearProducto(r.db.QueryRow(ctx,
		"SELECT "+columnasProducto+" FROM productos WHERE id = $1 AND (activo OR $2)", id, incluirInactivos))
}

// ListarProductosAdmin incluye inactivos, con paginación y búsqueda ILIKE.
func (r *PostgresRepository) ListarProductosAdmin(ctx context.Context, q string, page, limit int) (models.Pagina[models.Producto], error) {
	page, limit, offset := paginar(page, limit)
	pagina := models.Pagina[models.Producto]{Data: []models.Producto{}, Page: page, Limit: limit}
	filtro := "%" + strings.TrimSpace(q) + "%"
	const where = "WHERE nombre ILIKE $1 OR descripcion ILIKE $1 OR categoria ILIKE $1"

	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM productos "+where, filtro).Scan(&pagina.Total); err != nil {
		return pagina, err
	}
	datos, err := r.listarProductos(ctx,
		"SELECT "+columnasProducto+" FROM productos "+where+" ORDER BY id LIMIT $2 OFFSET $3", filtro, limit, offset)
	if err != nil {
		return pagina, err
	}
	pagina.Data = datos
	return pagina, nil
}

// CrearProducto inserta un producto nuevo.
func (r *PostgresRepository) CrearProducto(ctx context.Context, p models.Producto) (*models.Producto, error) {
	if p.Galeria == nil {
		p.Galeria = []string{}
	}
	return escanearProducto(r.db.QueryRow(ctx, `
		INSERT INTO productos (nombre, descripcion, precio, categoria, img, galeria, stock, activo)
		VALUES ($1, $2, $3, $4, $5, $6, $7, true)
		RETURNING `+columnasProducto,
		p.Nombre, p.Descripcion, p.Precio, p.Categoria, p.Img, p.Galeria, p.Stock))
}

// ActualizarProducto aplica una edición parcial: solo cambia los campos enviados.
func (r *PostgresRepository) ActualizarProducto(ctx context.Context, id int, c models.ProductoInput) (*models.Producto, error) {
	sets := []string{}
	args := []interface{}{id}
	agregar := func(columna string, valor interface{}) {
		args = append(args, valor)
		sets = append(sets, fmt.Sprintf("%s = $%d", columna, len(args)))
	}
	if c.Nombre != nil {
		agregar("nombre", *c.Nombre)
	}
	if c.Descripcion != nil {
		agregar("descripcion", *c.Descripcion)
	}
	if c.Precio != nil {
		agregar("precio", *c.Precio)
	}
	if c.Categoria != nil {
		agregar("categoria", *c.Categoria)
	}
	if c.Img != nil {
		agregar("img", *c.Img)
	}
	if c.Galeria != nil {
		agregar("galeria", *c.Galeria)
	}
	if c.Stock != nil {
		agregar("stock", *c.Stock)
	}
	if c.Activo != nil {
		agregar("activo", *c.Activo)
	}
	if len(sets) == 0 {
		return r.ObtenerProductoPorID(ctx, id, true)
	}
	return escanearProducto(r.db.QueryRow(ctx,
		"UPDATE productos SET "+strings.Join(sets, ", ")+" WHERE id = $1 RETURNING "+columnasProducto, args...))
}

// DesactivarProducto hace el borrado lógico (activo = false).
func (r *PostgresRepository) DesactivarProducto(ctx context.Context, id int) error {
	etiqueta, err := r.db.Exec(ctx, "UPDATE productos SET activo = false WHERE id = $1", id)
	if err != nil {
		return err
	}
	if etiqueta.RowsAffected() == 0 {
		return ErrNoEncontrado
	}
	return nil
}
