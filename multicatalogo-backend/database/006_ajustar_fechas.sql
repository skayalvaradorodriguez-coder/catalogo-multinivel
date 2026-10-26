-- 006_ajustar_fechas.sql — Mueve las fechas de los datos de prueba (005) al período real
-- del proyecto: desde el 9 de octubre de 2026 (base de datos conectada en la Fase 2) hasta hoy.
-- Solo toca lo generado por 005: pedidos con sus direcciones de prueba, sus comisiones,
-- sus liquidaciones y las cuentas @demo.ec. Tus pedidos reales no cambian.

BEGIN;

DROP TABLE IF EXISTS tmp_fechas;

-- Fecha nueva de cada pedido de prueba: se reparte proporcionalmente entre el inicio y hoy.
CREATE TABLE tmp_fechas AS
WITH demo AS (
    SELECT id, creado_en,
           MIN(creado_en) OVER () AS minimo,
           MAX(creado_en) OVER () AS maximo
    FROM pedidos
    WHERE direccion ~ '^(Av\. 9 de Octubre|Calle Guayaquil|Av\. Eleodoro Solórzano|Barrio Las Acacias|Av\. Carlos Espinoza|Cdla\. Puerto Lucía) #[0-9]{3}$'
), param AS (
    SELECT TIMESTAMPTZ '2026-10-09 09:00:00-05' AS inicio,
           LEAST(NOW() - INTERVAL '1 hour', (SELECT MAX(maximo) FROM demo)) AS fin
)
SELECT d.id,
       param.inicio + make_interval(secs =>
           EXTRACT(EPOCH FROM (d.creado_en - d.minimo))
           * EXTRACT(EPOCH FROM (param.fin - param.inicio))
           / GREATEST(EXTRACT(EPOCH FROM (d.maximo - d.minimo)), 1)) AS nueva
FROM demo d, param;

UPDATE comisiones c SET creado_en = t.nueva
FROM tmp_fechas t WHERE c.pedido_id = t.id;

UPDATE pedidos p SET creado_en = t.nueva
FROM tmp_fechas t WHERE p.id = t.id;

-- Cada liquidación se paga 2 días después del último pedido que incluye.
UPDATE liquidaciones l SET creado_en = x.fecha, pagada_en = x.fecha
FROM (
    SELECT c.liquidacion_id, LEAST(MAX(t.nueva) + INTERVAL '2 days', NOW()) AS fecha
    FROM comisiones c JOIN tmp_fechas t ON t.id = c.pedido_id
    WHERE c.liquidacion_id IS NOT NULL
    GROUP BY c.liquidacion_id
) x
WHERE l.id = x.liquidacion_id;

-- Las cuentas de prueba se crean el 8 y 9 de octubre, antes de su primer pedido.
UPDATE usuarios
SET creado_en = TIMESTAMPTZ '2026-10-08 10:00:00-05' + (id % 15) * INTERVAL '2 hours'
WHERE email LIKE '%@demo.ec';

DROP TABLE tmp_fechas;

COMMIT;

-- Comprobación: rango de fechas que quedó
SELECT 'pedidos' AS tabla, MIN(creado_en)::date AS desde, MAX(creado_en)::date AS hasta, COUNT(*) FROM pedidos
UNION ALL SELECT 'usuarios', MIN(creado_en)::date, MAX(creado_en)::date, COUNT(*) FROM usuarios
UNION ALL SELECT 'liquidaciones', MIN(pagada_en)::date, MAX(pagada_en)::date, COUNT(*) FROM liquidaciones;
