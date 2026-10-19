package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"multicatalogo-backend/domain"
	"multicatalogo-backend/models"

	"github.com/jackc/pgx/v5"
)

func numeroPedido(id int) string { return fmt.Sprintf("MC-%06d", id) }

// CrearPedido guarda el pedido, su detalle y el ledger de comisiones en UNA transacción ACID.
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

	pedido := &models.Pedido{
		UsuarioID: nuevo.UsuarioID, Cliente: nuevo.Cliente, MetodoPago: nuevo.MetodoPago,
		Estado: domain.EstadoPendiente, Items: []models.ItemPedido{},
	}

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

	// Falla simulada para la clase (SIMULAR_FALLA_CHECKOUT=true en el .env).
	if r.opciones.SimularFalla {
		return nil, errors.New("falla simulada: caída del sistema antes de procesar los items")
	}

	// 2. Detalle del pedido, restando el stock.
	for _, item := range nuevo.Items {
		if item.Cantidad <= 0 {
			return nil, ErrCantidadInvalida
		}
		linea := models.ItemPedido{ProductoID: item.ProductoID, Cantidad: item.Cantidad}

		// Si el producto no existe, la clave foránea producto_id rechaza el INSERT (código 23503).
		err = tx.QueryRow(ctx, `
			INSERT INTO pedido_items (pedido_id, producto_id, cantidad, precio_unitario)
			VALUES ($1, $2, $3, COALESCE((SELECT precio FROM productos WHERE id = $2), 0))
			RETURNING precio_unitario::float8`,
			pedido.ID, item.ProductoID, item.Cantidad,
		).Scan(&linea.Precio)
		if codigoPG(err) == "23503" {
			return nil, errValidacion("El producto con id %d no existe", item.ProductoID)
		}
		if err != nil {
			return nil, err
		}

		// FOR UPDATE: dos compras simultáneas no pueden vender el mismo stock.
		var stock int
		var activo bool
		err = tx.QueryRow(ctx,
			"SELECT nombre, stock, activo FROM productos WHERE id = $1 FOR UPDATE", item.ProductoID,
		).Scan(&linea.Nombre, &stock, &activo)
		if err != nil {
			return nil, err
		}
		if !activo {
			return nil, errValidacion("El producto %s ya no está disponible", linea.Nombre)
		}
		if stock < item.Cantidad {
			return nil, errValidacion("Stock insuficiente para %s: quedan %d unidades", linea.Nombre, stock)
		}
		if _, err = tx.Exec(ctx,
			"UPDATE productos SET stock = stock - $1 WHERE id = $2", item.Cantidad, item.ProductoID); err != nil {
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

	// 4. Ledger de comisiones: nace con el pedido, dentro de la misma transacción.
	if err = insertarComisiones(ctx, tx, pedido.ID, nuevo.UsuarioID, pedido.Total); err != nil {
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

	pedido.Numero = numeroPedido(pedido.ID)
	log.Printf("COMMIT: pedido %s guardado (total $%.2f)", pedido.Numero, pedido.Total)
	return pedido, nil
}

// insertarComisiones busca hasta 3 patrocinadores sobre el nodo del comprador
// (usuarios.referido_id → parent_id) y escribe una fila del ledger por cada uno.
func insertarComisiones(ctx context.Context, tx pgx.Tx, pedidoID, usuarioID int, total float64) error {
	filas, err := tx.Query(ctx, `
		WITH RECURSIVE upline AS (
			SELECT r.parent_id AS id, 1 AS distancia
			FROM usuarios u JOIN referidos r ON r.id = u.referido_id
			WHERE u.id = $1 AND r.parent_id IS NOT NULL
			UNION ALL
			SELECT r.parent_id, up.distancia + 1
			FROM upline up JOIN referidos r ON r.id = up.id
			WHERE r.parent_id IS NOT NULL AND up.distancia < $2
		)
		SELECT id FROM upline ORDER BY distancia`, usuarioID, len(domain.TasasMLM))
	if err != nil {
		return err
	}
	upline, err := pgx.CollectRows(filas, pgx.RowTo[int])
	if err != nil {
		return err
	}

	// Regla de negocio pura (sin I/O): cuánto le toca a cada nivel.
	reparto := domain.RepartirComisiones(total, domain.TasasMLM)
	for i, referidoID := range upline {
		c := reparto[i]
		if _, err := tx.Exec(ctx, `
			INSERT INTO comisiones (pedido_id, referido_id, nivel, porcentaje, monto, estado)
			VALUES ($1, $2, $3, $4, $5, 'pendiente')`,
			pedidoID, referidoID, c.Nivel, c.Porcentaje, c.Monto); err != nil {
			return err
		}
	}
	return nil
}

// SumarVentas actualiza el contador de ventas del nodo del comprador.
// Lo llama la goroutine después del COMMIT; el ledger ya quedó escrito en la transacción.
func (r *PostgresRepository) SumarVentas(ctx context.Context, usuarioID int, monto float64) error {
	_, err := r.db.Exec(ctx, `
		UPDATE referidos SET ventas = ventas + $2
		WHERE id = (SELECT referido_id FROM usuarios WHERE id = $1)`, usuarioID, monto)
	return err
}

// ListarPedidos devuelve pedidos paginados (de un usuario o todos) con filtro por estado.
func (r *PostgresRepository) ListarPedidos(ctx context.Context, f models.FiltroPedidos) (models.Pagina[models.PedidoResumen], error) {
	page, limit, offset := paginar(f.Page, f.Limit)
	pagina := models.Pagina[models.PedidoResumen]{Data: []models.PedidoResumen{}, Page: page, Limit: limit}
	const where = "WHERE ($1::int IS NULL OR p.usuario_id = $1) AND ($2 = '' OR p.estado = $2)"

	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM pedidos p "+where, f.UsuarioID, f.Estado).
		Scan(&pagina.Total); err != nil {
		return pagina, err
	}

	filas, err := r.db.Query(ctx, `
		SELECT p.id, p.total::float8, p.estado, p.cliente_nombre, p.cliente_email,
		       COALESCE((SELECT SUM(cantidad) FROM pedido_items WHERE pedido_id = p.id), 0), p.creado_en
		FROM pedidos p `+where+`
		ORDER BY p.id DESC LIMIT $3 OFFSET $4`, f.UsuarioID, f.Estado, limit, offset)
	if err != nil {
		return pagina, err
	}
	defer filas.Close()
	for filas.Next() {
		var p models.PedidoResumen
		if err := filas.Scan(&p.ID, &p.Total, &p.Estado, &p.ClienteNombre, &p.ClienteEmail, &p.Unidades, &p.CreadoEn); err != nil {
			return pagina, err
		}
		p.Numero = numeroPedido(p.ID)
		p.Siguientes = domain.SiguientesEstados(p.Estado)
		pagina.Data = append(pagina.Data, p)
	}
	return pagina, filas.Err()
}

// ObtenerPedido devuelve cabecera + ítems (con el nombre del producto) o ErrNoEncontrado.
func (r *PostgresRepository) ObtenerPedido(ctx context.Context, id int) (*models.Pedido, error) {
	p := &models.Pedido{Items: []models.ItemPedido{}}
	err := r.db.QueryRow(ctx, `
		SELECT id, usuario_id, total::float8, estado, metodo_pago, cliente_nombre, cliente_email,
		       direccion, ciudad, telefono, creado_en
		FROM pedidos WHERE id = $1`, id,
	).Scan(&p.ID, &p.UsuarioID, &p.Total, &p.Estado, &p.MetodoPago, &p.Cliente.Nombre, &p.Cliente.Email,
		&p.Cliente.Direccion, &p.Cliente.Ciudad, &p.Cliente.Telefono, &p.CreadoEn)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoEncontrado
	}
	if err != nil {
		return nil, err
	}

	filas, err := r.db.Query(ctx, `
		SELECT i.producto_id, pr.nombre, i.cantidad, i.precio_unitario::float8
		FROM pedido_items i JOIN productos pr ON pr.id = i.producto_id
		WHERE i.pedido_id = $1 ORDER BY i.id`, id)
	if err != nil {
		return nil, err
	}
	defer filas.Close()
	for filas.Next() {
		var item models.ItemPedido
		if err := filas.Scan(&item.ProductoID, &item.Nombre, &item.Cantidad, &item.Precio); err != nil {
			return nil, err
		}
		p.Items = append(p.Items, item)
	}
	if err := filas.Err(); err != nil {
		return nil, err
	}

	p.Numero = numeroPedido(p.ID)
	p.Siguientes = domain.SiguientesEstados(p.Estado)
	return p, nil
}

// CambiarEstadoPedido aplica una transición válida. Cancelar es el espejo ACID del checkout:
// en la misma transacción devuelve el stock, revierte las comisiones y descuenta las ventas.
func (r *PostgresRepository) CambiarEstadoPedido(ctx context.Context, id int, nuevo string) (*models.Pedido, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	var actual string
	var usuarioID int
	var total float64
	err = tx.QueryRow(ctx,
		"SELECT estado, usuario_id, total::float8 FROM pedidos WHERE id = $1 FOR UPDATE", id,
	).Scan(&actual, &usuarioID, &total)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoEncontrado
	}
	if err != nil {
		return nil, err
	}
	if !domain.TransicionValida(actual, nuevo) {
		return nil, errValidacion("No se puede pasar un pedido de %q a %q", actual, nuevo)
	}

	if _, err = tx.Exec(ctx, "UPDATE pedidos SET estado = $2 WHERE id = $1", id, nuevo); err != nil {
		return nil, err
	}

	if nuevo == domain.EstadoCancelado {
		// Devolver el stock de cada ítem.
		if _, err = tx.Exec(ctx, `
			UPDATE productos p SET stock = p.stock + i.cantidad
			FROM pedido_items i WHERE i.pedido_id = $1 AND i.producto_id = p.id`, id); err != nil {
			return nil, err
		}
		// Comisiones aún no pagadas: se anulan.
		if _, err = tx.Exec(ctx, `
			UPDATE comisiones SET estado = 'revertida'
			WHERE pedido_id = $1 AND estado IN ('pendiente', 'aprobada') AND monto > 0`, id); err != nil {
			return nil, err
		}
		// Comisiones ya pagadas: clawback, un cargo negativo que se descuenta en la próxima liquidación.
		if _, err = tx.Exec(ctx, `
			INSERT INTO comisiones (pedido_id, referido_id, nivel, porcentaje, monto, estado)
			SELECT pedido_id, referido_id, nivel, porcentaje, -monto, 'pendiente'
			FROM comisiones WHERE pedido_id = $1 AND estado = 'pagada' AND monto > 0`, id); err != nil {
			return nil, err
		}
		// El contador de ventas del comprador vuelve a su valor anterior.
		if _, err = tx.Exec(ctx, `
			UPDATE referidos SET ventas = GREATEST(ventas - $2, 0)
			WHERE id = (SELECT referido_id FROM usuarios WHERE id = $1)`, usuarioID, total); err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	log.Printf("Pedido %s: %s → %s", numeroPedido(id), actual, nuevo)
	return r.ObtenerPedido(ctx, id)
}
