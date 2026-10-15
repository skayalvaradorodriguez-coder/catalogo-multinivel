package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"multicatalogo-backend/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ErrValidacion indica un problema con los datos del pedido (producto inexistente,
// stock insuficiente...). El controlador lo traduce a HTTP 400; cualquier otro error es un 500.
type ErrValidacion struct {
	Mensaje string
}

func (e *ErrValidacion) Error() string { return e.Mensaje }

// TasaComision por nivel ascendente: el patrocinador directo cobra 10%, el siguiente 5% y el tercero 2%.
var TasaComision = map[int]float64{1: 0.10, 2: 0.05, 3: 0.02}

// CrearPedido guarda el pedido y su detalle dentro de una transacción ACID.
// Si cualquier paso falla, el defer hace ROLLBACK y no queda ningún "pedido fantasma".
func (r *PostgresRepository) CrearPedido(ctx context.Context, nuevo models.NuevoPedido) (*models.Pedido, error) {
	// BEGIN
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	// Si salimos de la función sin COMMIT, se revierte todo. Después de un COMMIT no hace nada.
	defer func() {
		if err := tx.Rollback(context.Background()); err == nil {
			log.Println("ROLLBACK: la transacción del pedido se revirtió")
		}
	}()

	pedido := &models.Pedido{Cliente: nuevo.Cliente, MetodoPago: nuevo.MetodoPago, Items: []models.ItemPedido{}}

	// 1. Registro principal del pedido (el total se calcula al final con los precios de la base).
	err = tx.QueryRow(ctx, `
		INSERT INTO pedidos (usuario_id, cliente_nombre, cliente_email, direccion, ciudad, telefono, metodo_pago, total)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 0)
		RETURNING id, creado_en`,
		nuevo.UsuarioID, nuevo.Cliente.Nombre, nuevo.Cliente.Email, nuevo.Cliente.Direccion,
		nuevo.Cliente.Ciudad, nuevo.Cliente.Telefono, nuevo.MetodoPago,
	).Scan(&pedido.ID, &pedido.CreadoEn)
	if err != nil {
		return nil, err
	}

	// Falla simulada para la clase (SIMULAR_FALLA_CHECKOUT=true en el .env):
	// el pedido ya se insertó, pero como no hay COMMIT, el ROLLBACK lo borra.
	if r.opciones.SimularFalla {
		return nil, errors.New("falla simulada: caída del sistema antes de procesar los items")
	}

	// 2. Detalle del pedido, restando el stock.
	for _, item := range nuevo.Items {
		var linea models.ItemPedido
		linea.ProductoID = item.ProductoID
		linea.Cantidad = item.Cantidad

		// El precio se copia desde productos. Si el producto no existe, la clave foránea
		// producto_id hace que PostgreSQL rechace el INSERT (código 23503).
		err = tx.QueryRow(ctx, `
			INSERT INTO pedido_items (pedido_id, producto_id, cantidad, precio_unitario)
			VALUES ($1, $2, $3, COALESCE((SELECT precio FROM productos WHERE id = $2), 0))
			RETURNING precio_unitario::float8`,
			pedido.ID, item.ProductoID, item.Cantidad,
		).Scan(&linea.Precio)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23503" {
				return nil, &ErrValidacion{Mensaje: fmt.Sprintf("El producto con id %d no existe", item.ProductoID)}
			}
			return nil, err
		}

		// Bloqueamos la fila (FOR UPDATE) para que dos compras simultáneas no vendan el mismo stock.
		var stock int
		err = tx.QueryRow(ctx,
			"SELECT nombre, stock FROM productos WHERE id = $1 FOR UPDATE", item.ProductoID,
		).Scan(&linea.Nombre, &stock)
		if err != nil {
			return nil, err
		}
		if stock < item.Cantidad {
			return nil, &ErrValidacion{Mensaje: fmt.Sprintf(
				"Stock insuficiente para %s: quedan %d unidades", linea.Nombre, stock)}
		}

		if _, err = tx.Exec(ctx,
			"UPDATE productos SET stock = stock - $1 WHERE id = $2", item.Cantidad, item.ProductoID,
		); err != nil {
			return nil, err
		}

		pedido.Items = append(pedido.Items, linea)
	}

	// 3. Total calculado por PostgreSQL a partir del detalle.
	err = tx.QueryRow(ctx, `
		UPDATE pedidos
		SET total = (SELECT COALESCE(SUM(cantidad * precio_unitario), 0) FROM pedido_items WHERE pedido_id = $1)
		WHERE id = $1
		RETURNING total::float8`, pedido.ID,
	).Scan(&pedido.Total)
	if err != nil {
		return nil, err
	}

	// Pausa opcional (CHECKOUT_PAUSA_SEGUNDOS en el .env) para detener PostgreSQL antes del COMMIT.
	if r.opciones.Pausa > 0 {
		log.Printf("Pedido %d listo: esperando %s antes del COMMIT...", pedido.ID, r.opciones.Pausa)
		time.Sleep(r.opciones.Pausa)
	}

	// COMMIT: a partir de aquí los cambios son permanentes.
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	pedido.Numero = fmt.Sprintf("MC-%06d", pedido.ID)
	log.Printf("COMMIT: pedido %s guardado (total $%.2f)", pedido.Numero, pedido.Total)
	return pedido, nil
}

// CalcularComisiones se ejecuta en una goroutine después del COMMIT del pedido.
// En su propia transacción: suma la venta al nodo del comprador en la red y paga
// comisiones a sus referidos ascendentes de nivel 1, 2 y 3.
func (r *PostgresRepository) CalcularComisiones(ctx context.Context, pedidoID, usuarioID int, total float64) ([]string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	// Nodo del comprador en la red multinivel.
	var nodoID int
	err = tx.QueryRow(ctx, "SELECT id FROM referidos WHERE usuario_id = $1", usuarioID).Scan(&nodoID)
	if errors.Is(err, pgx.ErrNoRows) {
		log.Printf("Comisiones: el usuario %d no pertenece a la red, no se pagan comisiones", usuarioID)
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if _, err = tx.Exec(ctx, "UPDATE referidos SET ventas = ventas + $1 WHERE id = $2", total, nodoID); err != nil {
		return nil, err
	}

	// Subimos por el árbol con una consulta recursiva (CTE) hasta 3 niveles.
	filas, err := tx.Query(ctx, `
		WITH RECURSIVE ascendentes AS (
			SELECT parent_id AS id, 1 AS distancia
			FROM referidos WHERE id = $1 AND parent_id IS NOT NULL
			UNION ALL
			SELECT r.parent_id, a.distancia + 1
			FROM ascendentes a JOIN referidos r ON r.id = a.id
			WHERE r.parent_id IS NOT NULL AND a.distancia < 3
		)
		SELECT a.id, a.distancia, u.email
		FROM ascendentes a
		JOIN referidos r ON r.id = a.id
		LEFT JOIN usuarios u ON u.id = r.usuario_id`, nodoID)
	if err != nil {
		return nil, err
	}

	type ascendente struct {
		id, distancia int
		email         *string
	}
	var lista []ascendente
	for filas.Next() {
		var a ascendente
		if err := filas.Scan(&a.id, &a.distancia, &a.email); err != nil {
			filas.Close()
			return nil, err
		}
		lista = append(lista, a)
	}
	filas.Close()
	if err := filas.Err(); err != nil {
		return nil, err
	}

	// Emails cuyo árbol cambió: el comprador y los ascendentes que tienen cuenta.
	var compradorEmail string
	if err = tx.QueryRow(ctx, "SELECT email FROM usuarios WHERE id = $1", usuarioID).Scan(&compradorEmail); err != nil {
		return nil, err
	}
	afectados := []string{compradorEmail}

	for _, a := range lista {
		monto := total * TasaComision[a.distancia]
		if _, err = tx.Exec(ctx,
			"UPDATE referidos SET comisiones = comisiones + $1 WHERE id = $2", monto, a.id); err != nil {
			return nil, err
		}
		if _, err = tx.Exec(ctx,
			"INSERT INTO comisiones (pedido_id, referido_id, nivel, monto) VALUES ($1, $2, $3, $4)",
			pedidoID, a.id, a.distancia, monto); err != nil {
			return nil, err
		}
		if a.email != nil {
			afectados = append(afectados, *a.email)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return afectados, nil
}
