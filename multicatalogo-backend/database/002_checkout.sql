-- Fase 4: migración para una base que YA tiene los datos de las fases 2 y 3.
-- Se puede ejecutar varias veces sin romper nada (usa IF NOT EXISTS).
--   psql -U postgres -d multicatalogo -f multicatalogo-backend/database/fase4.sql

-- Stock simulado de cada producto (se resta en cada compra).
ALTER TABLE productos ADD COLUMN IF NOT EXISTS stock INT NOT NULL DEFAULT 50 CHECK (stock >= 0);

-- Vincula cada nodo de la red con una cuenta de usuario y guarda su saldo de comisiones.
ALTER TABLE referidos ADD COLUMN IF NOT EXISTS usuario_id INT UNIQUE REFERENCES usuarios (id) ON DELETE SET NULL;
ALTER TABLE referidos ADD COLUMN IF NOT EXISTS comisiones NUMERIC(12, 2) NOT NULL DEFAULT 0;

-- El administrador es la raíz de la red ("Tú") y el cliente es Diana Paz (nivel 3),
-- así una compra del cliente paga comisiones a sus 3 niveles ascendentes:
-- Carlos Ruiz (10%), Ana García (5%) y el administrador (2%).
UPDATE referidos SET usuario_id = 1 WHERE id = 0;
UPDATE referidos SET usuario_id = 2 WHERE id = 7;

-- Cabecera del pedido.
CREATE TABLE IF NOT EXISTS pedidos (
    id             SERIAL PRIMARY KEY,
    usuario_id     INT            NOT NULL REFERENCES usuarios (id),
    cliente_nombre VARCHAR(150)   NOT NULL,
    cliente_email  VARCHAR(150)   NOT NULL,
    direccion      TEXT           NOT NULL,
    ciudad         VARCHAR(100)   NOT NULL,
    telefono       VARCHAR(30)    NOT NULL,
    metodo_pago    VARCHAR(50)    NOT NULL,
    total          NUMERIC(12, 2) NOT NULL DEFAULT 0,
    creado_en      TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

-- Detalle del pedido. La clave foránea a productos rechaza ids inexistentes.
CREATE TABLE IF NOT EXISTS pedido_items (
    id              SERIAL PRIMARY KEY,
    pedido_id       INT            NOT NULL REFERENCES pedidos (id) ON DELETE CASCADE,
    producto_id     INT            NOT NULL REFERENCES productos (id),
    cantidad        INT            NOT NULL CHECK (cantidad > 0),
    precio_unitario NUMERIC(10, 2) NOT NULL
);

-- Historial de comisiones pagadas por la goroutine (una fila por nivel ascendente).
CREATE TABLE IF NOT EXISTS comisiones (
    id          SERIAL PRIMARY KEY,
    pedido_id   INT            NOT NULL REFERENCES pedidos (id) ON DELETE CASCADE,
    referido_id INT            NOT NULL REFERENCES referidos (id) ON DELETE CASCADE,
    nivel       INT            NOT NULL CHECK (nivel BETWEEN 1 AND 3),
    monto       NUMERIC(12, 2) NOT NULL,
    creado_en   TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pedido_items_pedido ON pedido_items (pedido_id);
CREATE INDEX IF NOT EXISTS idx_comisiones_pedido ON comisiones (pedido_id);
