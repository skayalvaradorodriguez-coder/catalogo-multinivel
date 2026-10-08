// Package repository es la capa de acceso a datos: aquí vive TODO el SQL.
// Los controllers llaman a estas funciones y no conocen la base de datos.
package repository

import (
	"context"
	"errors"

	"multicatalogo-backend/config"
	"multicatalogo-backend/models"

	"github.com/jackc/pgx/v5"
)

const columnasProducto = "id, nombre, descripcion, precio::float8, categoria, img, galeria"

// BuscarUsuarioPorCredenciales devuelve el usuario que coincide con email y password.
// Si no existe devuelve (nil, nil); el error queda reservado para fallos de la base de datos.
func BuscarUsuarioPorCredenciales(ctx context.Context, email, password string) (*models.Usuario, error) {
	var u models.Usuario
	err := config.DB.QueryRow(ctx,
		"SELECT id, email, password, rol FROM usuarios WHERE email = $1 AND password = $2",
		email, password,
	).Scan(&u.ID, &u.Email, &u.Password, &u.Rol)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// ObtenerProductos devuelve todo el catálogo ordenado por id.
func ObtenerProductos(ctx context.Context) ([]models.Producto, error) {
	filas, err := config.DB.Query(ctx, "SELECT "+columnasProducto+" FROM productos ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer filas.Close()

	// Slice no nulo: si la tabla está vacía el JSON es [] y no null.
	productos := []models.Producto{}
	for filas.Next() {
		var p models.Producto
		if err := filas.Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Categoria, &p.Img, &p.Galeria); err != nil {
			return nil, err
		}
		productos = append(productos, p)
	}
	return productos, filas.Err()
}

// ObtenerProductoPorID devuelve un producto, o (nil, nil) si el id no existe.
func ObtenerProductoPorID(ctx context.Context, id int) (*models.Producto, error) {
	var p models.Producto
	err := config.DB.QueryRow(ctx,
		"SELECT "+columnasProducto+" FROM productos WHERE id = $1", id,
	).Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Precio, &p.Categoria, &p.Img, &p.Galeria)

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
func ObtenerRed(ctx context.Context) (*models.Referido, error) {
	filas, err := config.DB.Query(ctx,
		"SELECT id, nombre, nivel, ventas::float8, parent_id FROM referidos ORDER BY id")
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
		if err := filas.Scan(&nodo.ID, &nodo.Nombre, &nodo.Nivel, &nodo.Ventas, &parentID); err != nil {
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
