-- 004_backoffice.sql — Fase 0 del plan de implementación: columnas y tablas del back-office.
-- Idempotente: se puede ejecutar varias veces sobre la base existente.
--   psql -U postgres -d multicatalogo -f multicatalogo-backend/database/004_backoffice.sql

-- Usuarios: nombre, borrado lógico, fecha de alta y FK real a su nodo en la red.
ALTER TABLE usuarios ADD COLUMN IF NOT EXISTS nombre      VARCHAR(150) NOT NULL DEFAULT '';
ALTER TABLE usuarios ADD COLUMN IF NOT EXISTS activo      BOOLEAN      NOT NULL DEFAULT true;
ALTER TABLE usuarios ADD COLUMN IF NOT EXISTS referido_id INT UNIQUE REFERENCES referidos (id);
ALTER TABLE usuarios ADD COLUMN IF NOT EXISTS creado_en   TIMESTAMPTZ  NOT NULL DEFAULT NOW();

-- Referidos: código del link de invitación (/registro?ref=CODIGO).
ALTER TABLE referidos ADD COLUMN IF NOT EXISTS codigo VARCHAR(20) UNIQUE;
UPDATE referidos SET codigo = v.codigo
FROM (VALUES (0, 'TU'), (1, 'ANA'), (2, 'LUIS'), (3, 'MARTA'),
             (4, 'CARLOS'), (5, 'SOFIA'), (6, 'MARCO'), (7, 'DIANA')) AS v (id, codigo)
WHERE referidos.id = v.id AND referidos.codigo IS NULL;

-- Productos: borrado lógico (no se borra la fila para no romper pedido_items históricos).
ALTER TABLE productos ADD COLUMN IF NOT EXISTS activo BOOLEAN NOT NULL DEFAULT true;

-- Pedidos: ciclo de vida (pendiente → preparando → enviado → entregado | cancelado).
ALTER TABLE pedidos ADD COLUMN IF NOT EXISTS estado VARCHAR(20) NOT NULL DEFAULT 'pendiente';
ALTER TABLE pedidos ADD COLUMN IF NOT EXISTS usuario_id INT REFERENCES usuarios (id);
DO $$ BEGIN
    ALTER TABLE pedidos ADD CONSTRAINT pedidos_estado_check
        CHECK (estado IN ('pendiente', 'preparando', 'enviado', 'entregado', 'cancelado'));
EXCEPTION WHEN duplicate_object THEN NULL; END $$;
CREATE INDEX IF NOT EXISTS idx_pedidos_usuario ON pedidos (usuario_id);
CREATE INDEX IF NOT EXISTS idx_pedidos_estado  ON pedidos (estado);

-- Liquidaciones: pagos a distribuidores.
CREATE TABLE IF NOT EXISTS liquidaciones (
    id          SERIAL PRIMARY KEY,
    referido_id INT            NOT NULL REFERENCES referidos (id),
    total       NUMERIC(12, 2) NOT NULL CHECK (total > 0),
    estado      VARCHAR(20)    NOT NULL DEFAULT 'pagada' CHECK (estado IN ('pendiente', 'pagada')),
    creado_en   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    pagada_en   TIMESTAMPTZ
);

-- Ledger de comisiones (la tabla ya existe desde 002_checkout.sql; se completa aquí).
CREATE TABLE IF NOT EXISTS comisiones (
    id          SERIAL PRIMARY KEY,
    pedido_id   INT            NOT NULL REFERENCES pedidos (id) ON DELETE CASCADE,
    referido_id INT            NOT NULL REFERENCES referidos (id) ON DELETE CASCADE,
    nivel       INT            NOT NULL CHECK (nivel BETWEEN 1 AND 3),
    monto       NUMERIC(12, 2) NOT NULL,
    creado_en   TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);
ALTER TABLE comisiones ADD COLUMN IF NOT EXISTS porcentaje     NUMERIC(5, 2) NOT NULL DEFAULT 0;
ALTER TABLE comisiones ADD COLUMN IF NOT EXISTS estado         VARCHAR(20)   NOT NULL DEFAULT 'pendiente';
ALTER TABLE comisiones ADD COLUMN IF NOT EXISTS liquidacion_id INT REFERENCES liquidaciones (id);
DO $$ BEGIN
    ALTER TABLE comisiones ADD CONSTRAINT comisiones_estado_check
        CHECK (estado IN ('pendiente', 'aprobada', 'pagada', 'revertida'));
EXCEPTION WHEN duplicate_object THEN NULL; END $$;
UPDATE comisiones SET porcentaje = CASE nivel WHEN 1 THEN 10 WHEN 2 THEN 5 ELSE 2 END WHERE porcentaje = 0;
CREATE INDEX IF NOT EXISTS idx_comisiones_referido_pedido ON comisiones (referido_id, pedido_id);
CREATE INDEX IF NOT EXISTS idx_comisiones_estado          ON comisiones (estado);

-- El vínculo usuario ↔ nodo ahora vive en usuarios.referido_id (FK real).
-- Se retiran las columnas provisionales de la Fase 4 del curso.
ALTER TABLE referidos DROP COLUMN IF EXISTS usuario_id;
ALTER TABLE referidos DROP COLUMN IF EXISTS comisiones;

-- Datos semilla: admin → "Tú" (0), cliente → "Luis Poveda" (2).
UPDATE usuarios SET nombre = 'Administrador', referido_id = 0 WHERE email = 'admin@upse.edu.ec';
UPDATE usuarios SET nombre = 'Cliente Demo',  referido_id = 2 WHERE email = 'cliente@upse.edu.ec';

-- Las secuencias deben ir por delante de los ids sembrados a mano.
SELECT setval(pg_get_serial_sequence('referidos', 'id'), GREATEST((SELECT MAX(id) FROM referidos), 1));
SELECT setval(pg_get_serial_sequence('usuarios', 'id'),  GREATEST((SELECT MAX(id) FROM usuarios), 1));
