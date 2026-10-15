-- Fase 2: estructura relacional y datos semilla de Multicatálogo.
-- Docker lo ejecuta automáticamente la primera vez que se crea el volumen.

DROP TABLE IF EXISTS comisiones;
DROP TABLE IF EXISTS pedido_items;
DROP TABLE IF EXISTS pedidos;
DROP TABLE IF EXISTS referidos;
DROP TABLE IF EXISTS productos;
DROP TABLE IF EXISTS usuarios;

-- ---------------------------------------------------------------------------
-- Usuarios (password en texto plano TEMPORALMENTE; se cifrará en otra fase).
-- ---------------------------------------------------------------------------
CREATE TABLE usuarios (
    id       SERIAL PRIMARY KEY,
    email    VARCHAR(150) NOT NULL UNIQUE,
    password VARCHAR(150) NOT NULL,
    rol      VARCHAR(20)  NOT NULL CHECK (rol IN ('admin', 'cliente'))
);

-- ---------------------------------------------------------------------------
-- Productos del catálogo. "galeria" es un arreglo de URLs de imágenes.
-- ---------------------------------------------------------------------------
CREATE TABLE productos (
    id          SERIAL PRIMARY KEY,
    nombre      VARCHAR(150)   NOT NULL,
    descripcion TEXT           NOT NULL DEFAULT '',
    precio      NUMERIC(10, 2) NOT NULL CHECK (precio >= 0),
    categoria   VARCHAR(50)    NOT NULL DEFAULT '',
    img         TEXT           NOT NULL DEFAULT '',
    galeria     TEXT[]         NOT NULL DEFAULT '{}',
    stock       INT            NOT NULL DEFAULT 50 CHECK (stock >= 0)
);

-- ---------------------------------------------------------------------------
-- Red multinivel. parent_id apunta a la misma tabla para formar el árbol;
-- la raíz ("Tú") es la única fila con parent_id NULL.
-- ---------------------------------------------------------------------------
CREATE TABLE referidos (
    id        SERIAL PRIMARY KEY,
    nombre    VARCHAR(150)   NOT NULL,
    nivel     INT            NOT NULL CHECK (nivel >= 0),
    ventas    NUMERIC(12, 2) NOT NULL DEFAULT 0,
    parent_id INT REFERENCES referidos (id) ON DELETE CASCADE,
    -- Fase 4: cuenta de usuario del nodo y saldo de comisiones acumulado.
    usuario_id INT UNIQUE REFERENCES usuarios (id) ON DELETE SET NULL,
    comisiones NUMERIC(12, 2) NOT NULL DEFAULT 0
);

CREATE INDEX idx_referidos_parent ON referidos (parent_id);

-- ---------------------------------------------------------------------------
-- Datos semilla
-- ---------------------------------------------------------------------------
INSERT INTO usuarios (id, email, password, rol) VALUES
    (1, 'admin@upse.edu.ec',   '123456', 'admin'),
    (2, 'cliente@upse.edu.ec', '123456', 'cliente');

INSERT INTO productos (id, nombre, descripcion, precio, categoria, img, galeria) VALUES
    (1, 'Serum Revitalizante',
        'Serum concentrado con vitamina C y ácido hialurónico. Ilumina la piel, reduce las manchas y aporta hidratación profunda desde la primera aplicación.',
        45.00, 'Suero', 'https://picsum.photos/seed/serum/600',
        ARRAY['https://picsum.photos/seed/serum1/600', 'https://picsum.photos/seed/serum2/600', 'https://picsum.photos/seed/serum3/600']),
    (2, 'Crema Hidratante Pro',
        'Crema facial de textura ligera con niacinamida y manteca de karité. Hidratación de 24 horas y barrera cutánea fortalecida para todo tipo de piel.',
        32.50, 'Crema', 'https://picsum.photos/seed/crema/600',
        ARRAY['https://picsum.photos/seed/crema1/600', 'https://picsum.photos/seed/crema2/600', 'https://picsum.photos/seed/crema3/600']),
    (3, 'Tónico Purificante',
        'Tónico sin alcohol con hamamelis y agua de rosas. Limpia poros, equilibra el pH y prepara la piel para el resto de la rutina diaria.',
        28.00, 'Tónico', 'https://picsum.photos/seed/tonico/600',
        ARRAY['https://picsum.photos/seed/tonico1/600', 'https://picsum.photos/seed/tonico2/600', 'https://picsum.photos/seed/tonico3/600']),
    (4, 'Mascarilla Nocturna',
        'Mascarilla de noche con colágeno y vitamina E. Repara la piel mientras duermes y recupera la luminosidad al despertar.',
        50.00, 'Mascarilla', 'https://picsum.photos/seed/mascarilla/600',
        ARRAY['https://picsum.photos/seed/mascarilla1/600', 'https://picsum.photos/seed/mascarilla2/600', 'https://picsum.photos/seed/mascarilla3/600']),
    (5, 'Suero Anti-Edad Retinol',
        'Suero de retinol encapsulado que suaviza líneas de expresión y mejora la firmeza. Uso nocturno recomendado.',
        58.00, 'Suero', 'https://picsum.photos/seed/retinol/600',
        ARRAY['https://picsum.photos/seed/retinol1/600', 'https://picsum.photos/seed/retinol2/600', 'https://picsum.photos/seed/retinol3/600']),
    (6, 'Crema Contorno de Ojos',
        'Contorno de ojos con cafeína y péptidos. Reduce bolsas y ojeras, hidrata la zona más delicada del rostro.',
        26.00, 'Crema', 'https://picsum.photos/seed/ojos/600',
        ARRAY['https://picsum.photos/seed/ojos1/600', 'https://picsum.photos/seed/ojos2/600', 'https://picsum.photos/seed/ojos3/600']),
    (7, 'Tónico Exfoliante AHA-BHA',
        'Exfoliación química suave con ácidos AHA y BHA. Renueva la textura de la piel y desobstruye los poros.',
        34.00, 'Tónico', 'https://picsum.photos/seed/exfoliante/600',
        ARRAY['https://picsum.photos/seed/exfoliante1/600', 'https://picsum.photos/seed/exfoliante2/600', 'https://picsum.photos/seed/exfoliante3/600']),
    (8, 'Kit Rutina Completa',
        'Kit integral con suero, crema, tónico y mascarilla. La rutina perfecta para comenzar: limpieza, tratamiento e hidratación.',
        129.00, 'Kit', 'https://picsum.photos/seed/kit/600',
        ARRAY['https://picsum.photos/seed/kit1/600', 'https://picsum.photos/seed/kit2/600', 'https://picsum.photos/seed/kit3/600']);

-- Los padres se insertan antes que sus hijos para respetar la clave foránea.
INSERT INTO referidos (id, nombre, nivel, ventas, parent_id) VALUES
    (0, 'Tú',            0, 2400, NULL),
    (1, 'Ana García',    1, 1200, 0),
    (2, 'Luis Poveda',   1,  850, 0),
    (3, 'Marta Sánchez', 1,  430, 0),
    (4, 'Carlos Ruiz',   2,  500, 1),
    (5, 'Sofía León',    2,  430, 1),
    (6, 'Marco Díaz',    2,  380, 2),
    (7, 'Diana Paz',     3,  300, 4);

-- Fase 4: el administrador es la raíz ("Tú") y el cliente es Diana Paz (nivel 3).
UPDATE referidos SET usuario_id = 1 WHERE id = 0;
UPDATE referidos SET usuario_id = 2 WHERE id = 7;

-- Como insertamos ids explícitos, adelantamos las secuencias para los próximos INSERT.
SELECT setval(pg_get_serial_sequence('usuarios', 'id'),  (SELECT MAX(id) FROM usuarios));
SELECT setval(pg_get_serial_sequence('productos', 'id'), (SELECT MAX(id) FROM productos));
SELECT setval(pg_get_serial_sequence('referidos', 'id'), (SELECT MAX(id) FROM referidos));

-- ---------------------------------------------------------------------------
-- Fase 4: pedidos, detalle y comisiones
-- ---------------------------------------------------------------------------
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
