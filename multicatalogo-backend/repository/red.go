package repository

import (
	"context"

	"multicatalogo-backend/models"
)

// ObtenerRed devuelve el árbol completo desde la raíz (vista del administrador).
func (r *PostgresRepository) ObtenerRed(ctx context.Context) (*models.Referido, error) {
	return r.armarArbol(ctx, `
		SELECT id, nombre, nivel, ventas::float8, COALESCE(codigo, ''), parent_id
		FROM referidos ORDER BY id`)
}

// ObtenerSubRed devuelve el árbol que cuelga del nodo indicado (la red de un usuario).
// El nivel se recalcula relativo a ese nodo: él es 0, sus directos 1, etc.,
// así las tasas 10/5/2 % del frontend se aplican desde su punto de vista.
func (r *PostgresRepository) ObtenerSubRed(ctx context.Context, referidoID int) (*models.Referido, error) {
	return r.armarArbol(ctx, `
		WITH RECURSIVE subred AS (
			SELECT id, nombre, 0 AS nivel, ventas, codigo, NULL::int AS parent_id
			FROM referidos WHERE id = $1
			UNION ALL
			SELECT r.id, r.nombre, s.nivel + 1, r.ventas, r.codigo, r.parent_id
			FROM referidos r JOIN subred s ON r.parent_id = s.id
		)
		SELECT id, nombre, nivel, ventas::float8, COALESCE(codigo, ''), parent_id
		FROM subred ORDER BY nivel, id`, referidoID)
}

// armarArbol lee una lista plana (id, ..., parent_id) y la convierte en árbol en memoria.
// Devuelve ErrNoEncontrado si no hay filas.
func (r *PostgresRepository) armarArbol(ctx context.Context, sql string, args ...interface{}) (*models.Referido, error) {
	filas, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer filas.Close()

	nodos := map[int]*models.Referido{}
	padres := map[int]*int{}
	var orden []int
	for filas.Next() {
		nodo := &models.Referido{}
		var parentID *int
		if err := filas.Scan(&nodo.ID, &nodo.Nombre, &nodo.Nivel, &nodo.Ventas, &nodo.Codigo, &parentID); err != nil {
			return nil, err
		}
		nodos[nodo.ID] = nodo
		padres[nodo.ID] = parentID
		orden = append(orden, nodo.ID)
	}
	if err := filas.Err(); err != nil {
		return nil, err
	}

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
	if raiz == nil {
		return nil, ErrNoEncontrado
	}
	return raiz, nil
}
