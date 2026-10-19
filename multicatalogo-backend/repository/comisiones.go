package repository

import (
	"context"
	"errors"

	"multicatalogo-backend/models"

	"github.com/jackc/pgx/v5"
)

// EstadoCuenta devuelve el ledger de un distribuidor y sus totales por estado.
func (r *PostgresRepository) EstadoCuenta(ctx context.Context, referidoID int) (*models.EstadoCuenta, error) {
	cuenta := &models.EstadoCuenta{
		Comisiones: []models.Comision{},
		Totales:    map[string]float64{"pendiente": 0, "pagada": 0, "revertida": 0},
	}
	filas, err := r.db.Query(ctx, `
		SELECT id, pedido_id, referido_id, nivel, porcentaje::float8, monto::float8, estado, liquidacion_id, creado_en
		FROM comisiones WHERE referido_id = $1 ORDER BY id DESC`, referidoID)
	if err != nil {
		return nil, err
	}
	defer filas.Close()
	for filas.Next() {
		var c models.Comision
		if err := filas.Scan(&c.ID, &c.PedidoID, &c.ReferidoID, &c.Nivel, &c.Porcentaje, &c.Monto,
			&c.Estado, &c.LiquidacionID, &c.CreadoEn); err != nil {
			return nil, err
		}
		c.PedidoNumero = numeroPedido(c.PedidoID)
		cuenta.Comisiones = append(cuenta.Comisiones, c)
		cuenta.Totales[c.Estado] += c.Monto
	}
	return cuenta, filas.Err()
}

// Liquidar paga las comisiones pendientes de un distribuidor en una transacción:
// crea la liquidación con el total y marca esas comisiones como pagadas.
// Bloquea el nodo con FOR UPDATE, así dos liquidaciones simultáneas no pagan dos veces.
func (r *PostgresRepository) Liquidar(ctx context.Context, referidoID int) (*models.Liquidacion, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	liq := &models.Liquidacion{ReferidoID: referidoID, Estado: "pagada"}
	err = tx.QueryRow(ctx, "SELECT nombre FROM referidos WHERE id = $1 FOR UPDATE", referidoID).Scan(&liq.ReferidoNombre)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoEncontrado
	}
	if err != nil {
		return nil, err
	}

	err = tx.QueryRow(ctx, `
		SELECT COALESCE(SUM(monto), 0)::float8, COUNT(*)
		FROM comisiones WHERE referido_id = $1 AND estado = 'pendiente'`, referidoID,
	).Scan(&liq.Total, &liq.Comisiones)
	if err != nil {
		return nil, err
	}
	if liq.Total <= 0 {
		return nil, errValidacion("%s no tiene comisiones pendientes por liquidar", liq.ReferidoNombre)
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO liquidaciones (referido_id, total, estado, pagada_en)
		VALUES ($1, $2, 'pagada', NOW()) RETURNING id, creado_en, pagada_en`,
		referidoID, liq.Total).Scan(&liq.ID, &liq.CreadoEn, &liq.PagadaEn)
	if err != nil {
		return nil, err
	}

	if _, err = tx.Exec(ctx, `
		UPDATE comisiones SET estado = 'pagada', liquidacion_id = $2
		WHERE referido_id = $1 AND estado = 'pendiente'`, referidoID, liq.ID); err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return liq, nil
}

// ListarLiquidaciones devuelve el historial de pagos (de un distribuidor o de todos).
func (r *PostgresRepository) ListarLiquidaciones(ctx context.Context, referidoID *int, page, limit int) (models.Pagina[models.Liquidacion], error) {
	page, limit, offset := paginar(page, limit)
	pagina := models.Pagina[models.Liquidacion]{Data: []models.Liquidacion{}, Page: page, Limit: limit}

	if err := r.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM liquidaciones WHERE $1::int IS NULL OR referido_id = $1", referidoID,
	).Scan(&pagina.Total); err != nil {
		return pagina, err
	}

	filas, err := r.db.Query(ctx, `
		SELECT l.id, l.referido_id, r.nombre, l.total::float8, l.estado,
		       (SELECT COUNT(*) FROM comisiones c WHERE c.liquidacion_id = l.id), l.creado_en, l.pagada_en
		FROM liquidaciones l JOIN referidos r ON r.id = l.referido_id
		WHERE $1::int IS NULL OR l.referido_id = $1
		ORDER BY l.id DESC LIMIT $2 OFFSET $3`, referidoID, limit, offset)
	if err != nil {
		return pagina, err
	}
	defer filas.Close()
	for filas.Next() {
		var l models.Liquidacion
		if err := filas.Scan(&l.ID, &l.ReferidoID, &l.ReferidoNombre, &l.Total, &l.Estado,
			&l.Comisiones, &l.CreadoEn, &l.PagadaEn); err != nil {
			return pagina, err
		}
		pagina.Data = append(pagina.Data, l)
	}
	return pagina, filas.Err()
}
