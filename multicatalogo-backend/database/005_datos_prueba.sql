-- 005_datos_prueba.sql — Datos de prueba adicionales para la demostración.
-- Agrega productos, distribuidores con cuenta, ~30 pedidos de los últimos 60 días con su
-- detalle, el ledger de comisiones (10 % / 5 % / 2 %) y dos liquidaciones pagadas.
--
-- Se ejecuta DESPUÉS de init.sql (o sobre una base con datos). No borra nada.
-- Todas las cuentas nuevas tienen la contraseña: 123456
--
--   Neon:  pegar todo en el SQL Editor y pulsar Run
--   Local: psql -U postgres -d multicatalogo -f 005_datos_prueba.sql

BEGIN;

-- ---------------------------------------------------------------------------
-- 1. Productos nuevos (si ya existe uno con el mismo nombre, se omite)
-- ---------------------------------------------------------------------------
INSERT INTO productos (nombre, descripcion, precio, categoria, img, galeria, stock)
SELECT v.nombre, v.descripcion, v.precio, v.categoria,
       'https://picsum.photos/seed/' || v.semilla || '/600',
       ARRAY['https://picsum.photos/seed/' || v.semilla || '1/600',
             'https://picsum.photos/seed/' || v.semilla || '2/600',
             'https://picsum.photos/seed/' || v.semilla || '3/600'],
       v.stock
FROM (VALUES
    ('Protector Solar FPS 50', 'Protector solar facial de amplio espectro, textura invisible y resistente al agua. Ideal para uso diario en la costa.', 24.90, 'Protector', 'solar', 80),
    ('Gel Limpiador Facial', 'Gel limpiador suave con aloe vera y té verde. Elimina impurezas y maquillaje sin resecar la piel.', 18.50, 'Limpiador', 'limpiador', 90),
    ('Agua Micelar Calmante', 'Agua micelar con manzanilla para pieles sensibles. Desmaquilla y calma en un solo paso.', 15.00, 'Limpiador', 'micelar', 70),
    ('Crema Corporal Coco', 'Crema corporal con aceite de coco y manteca de cacao. Hidratación intensa y aroma tropical.', 21.00, 'Crema', 'coco', 60),
    ('Bálsamo Labial Reparador', 'Bálsamo con cera de abejas y vitamina E. Repara labios agrietados y protege del sol.', 8.50, 'Labios', 'labial', 120),
    ('Suero Niacinamida 10%', 'Suero que reduce poros visibles y controla el brillo. Apto para piel mixta y grasa.', 39.00, 'Suero', 'niacinamida', 55),
    ('Kit Viaje Esencial', 'Kit en tamaño de viaje: limpiador, tónico, crema y protector solar en un neceser.', 59.90, 'Kit', 'viaje', 40)
) AS v(nombre, descripcion, precio, categoria, semilla, stock)
WHERE NOT EXISTS (SELECT 1 FROM productos p WHERE p.nombre = v.nombre);

-- ---------------------------------------------------------------------------
-- 2. Nuevos distribuidores en la red (el nivel se calcula a partir del patrocinador).
--    Se insertan en orden para que un nuevo nodo pueda colgar de otro nuevo.
-- ---------------------------------------------------------------------------
DO $$
DECLARE
    nuevo RECORD;
BEGIN
    FOR nuevo IN
        SELECT * FROM (VALUES
            (1, 'Daniel Pozo',      'DANIEL',   'TU'),
            (2, 'Valeria Mora',     'VALERIA',  'ANA'),
            (3, 'Jorge Tomalá',     'JORGE',    'LUIS'),
            (4, 'Andrés Villón',    'ANDRES',   'MARTA'),
            (5, 'Karla Reyes',      'KARLA',    'JORGE'),
            (6, 'Paola Suárez',     'PAOLA',    'CARLOS'),
            (7, 'Gabriela Rosales', 'GABRIELA', 'SOFIA'),
            (8, 'Kevin Yagual',     'KEVIN',    'DIANA'),
            (9, 'Lucía Tigrero',    'LUCIA',    'DANIEL')
        ) AS t(orden, nombre, codigo, padre)
        ORDER BY orden
    LOOP
        INSERT INTO referidos (nombre, nivel, ventas, parent_id, codigo)
        SELECT nuevo.nombre, p.nivel + 1, 0, p.id, nuevo.codigo
        FROM referidos p
        WHERE p.codigo = nuevo.padre
        ON CONFLICT (codigo) DO NOTHING;
    END LOOP;
END $$;

-- ---------------------------------------------------------------------------
-- 3. Cuentas de cliente para los distribuidores que todavía no tienen una
--    (contraseña 123456, hash bcrypt compatible con el backend).
-- ---------------------------------------------------------------------------
INSERT INTO usuarios (email, password, rol, nombre, referido_id, creado_en, activo)
SELECT v.email, crypt('123456', gen_salt('bf', 10)), 'cliente', r.nombre, r.id,
       NOW() - (v.dias || ' days')::interval, v.activo
FROM (VALUES
    ('ana.garcia@demo.ec',       'ANA',      58, true),
    ('marta.sanchez@demo.ec',    'MARTA',    55, true),
    ('carlos.ruiz@demo.ec',      'CARLOS',   50, true),
    ('sofia.leon@demo.ec',       'SOFIA',    47, true),
    ('marco.diaz@demo.ec',       'MARCO',    44, true),
    ('diana.paz@demo.ec',        'DIANA',    40, true),
    ('daniel.pozo@demo.ec',      'DANIEL',   36, true),
    ('valeria.mora@demo.ec',     'VALERIA',  33, true),
    ('jorge.tomala@demo.ec',     'JORGE',    30, true),
    ('andres.villon@demo.ec',    'ANDRES',   27, true),
    ('karla.reyes@demo.ec',      'KARLA',    22, true),
    ('paola.suarez@demo.ec',     'PAOLA',    18, true),
    ('gabriela.rosales@demo.ec', 'GABRIELA', 14, true),
    ('kevin.yagual@demo.ec',     'KEVIN',     9, true),
    ('lucia.tigrero@demo.ec',    'LUCIA',     5, false)   -- cuenta desactivada (para probar el back-office)
) AS v(email, codigo, dias, activo)
JOIN referidos r ON r.codigo = v.codigo
WHERE NOT EXISTS (SELECT 1 FROM usuarios u WHERE u.referido_id = r.id)
ON CONFLICT (email) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 4. Pedidos de los últimos 60 días con detalle, stock, comisiones y ventas.
--    setseed hace que el resultado sea siempre el mismo.
-- ---------------------------------------------------------------------------
DO $$
DECLARE
    tasas      NUMERIC[] := ARRAY[10, 5, 2];
    estados    TEXT[]    := ARRAY['entregado', 'entregado', 'entregado', 'entregado', 'enviado',
                                  'enviado', 'preparando', 'pendiente', 'pendiente', 'cancelado'];
    metodos    TEXT[]    := ARRAY['Tarjeta de crédito', 'Transferencia bancaria', 'Efectivo contra entrega'];
    ciudades   TEXT[]    := ARRAY['La Libertad', 'Santa Elena', 'Salinas', 'Guayaquil', 'Ballenita', 'Montañita'];
    calles     TEXT[]    := ARRAY['Av. 9 de Octubre', 'Calle Guayaquil', 'Av. Eleodoro Solórzano',
                                  'Barrio Las Acacias', 'Av. Carlos Espinoza', 'Cdla. Puerto Lucía'];
    compradores INT[];
    u          RECORD;
    p          RECORD;
    v_pedido   INT;
    v_estado   TEXT;
    v_fecha    TIMESTAMPTZ;
    v_total    NUMERIC(12, 2);
    v_cant     INT;
    v_nodo     INT;
    v_nivel    INT;
    i          INT;
BEGIN
    PERFORM setseed(0.42);

    SELECT array_agg(id ORDER BY id) INTO compradores
    FROM usuarios
    WHERE rol = 'cliente' AND activo AND referido_id IS NOT NULL;

    IF compradores IS NULL THEN
        RAISE EXCEPTION 'No hay clientes activos con nodo en la red';
    END IF;

    FOR i IN 1..30 LOOP
        v_nodo := compradores[1 + floor(random() * array_length(compradores, 1))::int];
        SELECT * INTO u FROM usuarios WHERE id = v_nodo;

        v_estado := estados[1 + floor(random() * array_length(estados, 1))::int];
        v_fecha  := NOW() - ((60 - i * 2) || ' days')::interval - (floor(random() * 12) || ' hours')::interval;

        INSERT INTO pedidos (usuario_id, cliente_nombre, cliente_email, direccion, ciudad, telefono,
                             metodo_pago, total, estado, creado_en)
        VALUES (u.id, u.nombre, u.email,
                calles[1 + floor(random() * array_length(calles, 1))::int] || ' #' || (100 + floor(random() * 900))::int,
                ciudades[1 + floor(random() * array_length(ciudades, 1))::int],
                '09' || lpad((floor(random() * 100000000))::int::text, 8, '0'),
                metodos[1 + floor(random() * array_length(metodos, 1))::int],
                0, v_estado, v_fecha)
        RETURNING id INTO v_pedido;

        v_total := 0;
        v_cant := 1 + floor(random() * 3)::int;   -- cuántos productos distintos lleva el pedido
        FOR p IN
            SELECT id, precio FROM productos WHERE activo ORDER BY random() LIMIT v_cant
        LOOP
            v_cant := 1 + floor(random() * 3)::int;
            INSERT INTO pedido_items (pedido_id, producto_id, cantidad, precio_unitario)
            VALUES (v_pedido, p.id, v_cant, p.precio);
            v_total := v_total + p.precio * v_cant;
            IF v_estado <> 'cancelado' THEN
                UPDATE productos SET stock = GREATEST(stock - v_cant, 0) WHERE id = p.id;
            END IF;
        END LOOP;

        UPDATE pedidos SET total = v_total WHERE id = v_pedido;

        -- Ledger: hasta 3 patrocinadores sobre el nodo del comprador (igual que el checkout).
        SELECT parent_id INTO v_nodo FROM referidos WHERE id = u.referido_id;
        v_nivel := 1;
        WHILE v_nodo IS NOT NULL AND v_nivel <= 3 LOOP
            INSERT INTO comisiones (pedido_id, referido_id, nivel, porcentaje, monto, estado, creado_en)
            VALUES (v_pedido, v_nodo, v_nivel, tasas[v_nivel], round(v_total * tasas[v_nivel] / 100, 2),
                    CASE WHEN v_estado = 'cancelado' THEN 'revertida' ELSE 'pendiente' END, v_fecha);
            SELECT parent_id INTO v_nodo FROM referidos WHERE id = v_nodo;
            v_nivel := v_nivel + 1;
        END LOOP;

        IF v_estado <> 'cancelado' THEN
            UPDATE referidos SET ventas = ventas + v_total WHERE id = u.referido_id;
        END IF;
    END LOOP;
END $$;

-- ---------------------------------------------------------------------------
-- 5. Dos liquidaciones pagadas: los 2 distribuidores con más comisiones de pedidos
--    entregados hace más de 20 días cobran esas comisiones.
-- ---------------------------------------------------------------------------
DO $$
DECLARE
    d       RECORD;
    v_liq   INT;
BEGIN
    FOR d IN
        SELECT c.referido_id, SUM(c.monto) AS total
        FROM comisiones c JOIN pedidos p ON p.id = c.pedido_id
        WHERE c.estado = 'pendiente' AND c.liquidacion_id IS NULL AND c.monto > 0
          AND p.estado = 'entregado' AND p.creado_en < NOW() - INTERVAL '20 days'
          AND c.referido_id <> (SELECT id FROM referidos WHERE parent_id IS NULL LIMIT 1)
        GROUP BY c.referido_id
        ORDER BY SUM(c.monto) DESC
        LIMIT 2
    LOOP
        INSERT INTO liquidaciones (referido_id, total, estado, creado_en, pagada_en)
        VALUES (d.referido_id, d.total, 'pagada', NOW() - INTERVAL '15 days', NOW() - INTERVAL '15 days')
        RETURNING id INTO v_liq;

        UPDATE comisiones c SET estado = 'pagada', liquidacion_id = v_liq
        FROM pedidos p
        WHERE p.id = c.pedido_id AND c.referido_id = d.referido_id
          AND c.estado = 'pendiente' AND c.liquidacion_id IS NULL AND c.monto > 0
          AND p.estado = 'entregado' AND p.creado_en < NOW() - INTERVAL '20 days';
    END LOOP;
END $$;

COMMIT;

-- Resumen de lo que hay ahora en la base
SELECT 'productos' AS tabla, COUNT(*) FROM productos
UNION ALL SELECT 'usuarios', COUNT(*) FROM usuarios
UNION ALL SELECT 'distribuidores (red)', COUNT(*) FROM referidos
UNION ALL SELECT 'pedidos', COUNT(*) FROM pedidos
UNION ALL SELECT 'comisiones', COUNT(*) FROM comisiones
UNION ALL SELECT 'liquidaciones', COUNT(*) FROM liquidaciones;
