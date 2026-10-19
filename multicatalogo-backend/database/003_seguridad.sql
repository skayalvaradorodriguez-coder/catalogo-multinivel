-- 003_seguridad.sql — U2-T4: passwords con bcrypt + índices de consulta.
-- Idempotente: se puede ejecutar varias veces.
--   psql -U postgres -d multicatalogo -f multicatalogo-backend/database/003_seguridad.sql

-- pgcrypto aporta crypt() y gen_salt('bf') (bcrypt). Go verifica estos hashes
-- con bcrypt.CompareHashAndPassword porque usan el mismo formato $2a$.
CREATE EXTENSION IF NOT EXISTS pgcrypto;

ALTER TABLE usuarios ALTER COLUMN password TYPE VARCHAR(100);

-- Solo se hashean las contraseñas que todavía están en texto plano.
UPDATE usuarios
SET password = crypt(password, gen_salt('bf', 10))
WHERE password NOT LIKE '$2%';

-- Índices para las consultas frecuentes.
CREATE INDEX IF NOT EXISTS idx_referidos_parent    ON referidos (parent_id);
CREATE INDEX IF NOT EXISTS idx_pedidos_email       ON pedidos (cliente_email);
CREATE INDEX IF NOT EXISTS idx_pedido_items_pedido ON pedido_items (pedido_id);
CREATE INDEX IF NOT EXISTS idx_productos_categoria ON productos (categoria);
