package repository

import (
	"context"
	"time"

	"multicatalogo-backend/domain"
	"multicatalogo-backend/models"

	"github.com/jackc/pgx/v5"
)

// ResumenVentas calcula los KPIs del periodo [desde, hasta]. Los pedidos cancelados no cuentan como venta.
func (r *PostgresRepository) ResumenVentas(ctx context.Context, desde, hasta *time.Time) (*models.ResumenVentas, error) {
	res := &models.ResumenVentas{}
	// hasta es inclusivo: se compara contra el día siguiente.
	var hastaExclusivo *time.Time
	if hasta != nil {
		h := hasta.AddDate(0, 0, 1)
		hastaExclusivo = &h
	}
	const periodo = "($1::timestamptz IS NULL OR p.creado_en >= $1) AND ($2::timestamptz IS NULL OR p.creado_en < $2)"

	err := r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(p.total) FILTER (WHERE p.estado <> 'cancelado'), 0)::float8,
		       COUNT(*) FILTER (WHERE p.estado <> 'cancelado'),
		       COUNT(*) FILTER (WHERE p.estado = 'cancelado')
		FROM pedidos p WHERE `+periodo, desde, hastaExclusivo,
	).Scan(&res.TotalVendido, &res.NumPedidos, &res.PedidosCancelados)
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(c.monto) FILTER (WHERE c.estado <> 'revertida'), 0)::float8,
		       COALESCE(SUM(c.monto) FILTER (WHERE c.estado = 'pendiente'), 0)::float8,
		       COALESCE(SUM(c.monto) FILTER (WHERE c.estado = 'pagada'), 0)::float8
		FROM comisiones c JOIN pedidos p ON p.id = c.pedido_id WHERE `+periodo, desde, hastaExclusivo,
	).Scan(&res.ComisionesDevengadas, &res.ComisionesPendientes, &res.ComisionesPagadas)
	if err != nil {
		return nil, err
	}

	if res.NumPedidos > 0 {
		res.TicketPromedio = domain.Redondear(res.TotalVendido / float64(res.NumPedidos))
	}
	return res, nil
}

// TopProductos ordena los productos por ingresos (cantidad × precio) de pedidos no cancelados.
func (r *PostgresRepository) TopProductos(ctx context.Context, limit int) ([]models.TopProducto, error) {
	filas, err := r.db.Query(ctx, `
		SELECT pr.id, pr.nombre, SUM(i.cantidad)::int, SUM(i.cantidad * i.precio_unitario)::float8
		FROM pedido_items i
		JOIN pedidos p   ON p.id = i.pedido_id AND p.estado <> 'cancelado'
		JOIN productos pr ON pr.id = i.producto_id
		GROUP BY pr.id, pr.nombre
		ORDER BY 4 DESC, 3 DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(filas, func(f pgx.CollectableRow) (models.TopProducto, error) {
		var t models.TopProducto
		err := f.Scan(&t.ProductoID, &t.Nombre, &t.Unidades, &t.Ingresos)
		return t, err
	})
}

// TopReferidos ordena a los distribuidores por las ventas reales de sus cuentas
// y muestra las comisiones que han ganado en el ledger.
func (r *PostgresRepository) TopReferidos(ctx context.Context, limit int) ([]models.TopReferido, error) {
	filas, err := r.db.Query(ctx, `
		SELECT r.id, r.nombre, COALESCE(r.codigo, ''), r.nivel,
		       COUNT(p.id) FILTER (WHERE p.estado <> 'cancelado'),
		       COALESCE(SUM(p.total) FILTER (WHERE p.estado <> 'cancelado'), 0)::float8,
		       COALESCE((SELECT SUM(c.monto) FROM comisiones c
		                 WHERE c.referido_id = r.id AND c.estado <> 'revertida'), 0)::float8
		FROM referidos r
		LEFT JOIN usuarios u ON u.referido_id = r.id
		LEFT JOIN pedidos p  ON p.usuario_id = u.id
		GROUP BY r.id
		HAVING COUNT(p.id) > 0 OR EXISTS (SELECT 1 FROM comisiones c WHERE c.referido_id = r.id)
		ORDER BY 6 DESC, 7 DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(filas, func(f pgx.CollectableRow) (models.TopReferido, error) {
		var t models.TopReferido
		err := f.Scan(&t.ReferidoID, &t.Nombre, &t.Codigo, &t.Nivel, &t.Pedidos, &t.Ventas, &t.Comisiones)
		return t, err
	})
}

// ComisionesPendientes es la deuda por distribuidor (lo que falta liquidar).
func (r *PostgresRepository) ComisionesPendientes(ctx context.Context) ([]models.DeudaComisiones, error) {
	filas, err := r.db.Query(ctx, `
		SELECT r.id, r.nombre, COALESCE(r.codigo, ''), u.email, COUNT(c.id), SUM(c.monto)::float8
		FROM comisiones c
		JOIN referidos r ON r.id = c.referido_id
		LEFT JOIN usuarios u ON u.referido_id = r.id
		WHERE c.estado = 'pendiente'
		GROUP BY r.id, r.nombre, r.codigo, u.email
		HAVING SUM(c.monto) > 0
		ORDER BY 6 DESC`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(filas, func(f pgx.CollectableRow) (models.DeudaComisiones, error) {
		var d models.DeudaComisiones
		err := f.Scan(&d.ReferidoID, &d.Nombre, &d.Codigo, &d.Email, &d.Comisiones, &d.Pendiente)
		return d, err
	})
}
