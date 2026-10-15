# PROJECT_MAP — Catálogo Multinivel

> Mapa ultra-compacto. Consultar ANTES de abrir cualquier archivo de código.

## Stack
| Capa | Tech | Puerto |
|------|------|--------|
| Frontend | React 19 + Vite 8 + Tailwind 4 + React Router 7 | 5173 |
| Backend | Go + Fiber v2 | 3000 |
| Auth | Fake JWT en memoria (admin/cliente) | — |
| Datos | PostgreSQL 16 en Docker (pgx) | 5432 |
| Caché | Redis (go-redis v9), patrón Cache-Aside | 6379 |

## Estructura raíz
catalogo-multinivel/
├── src/                          # FRONTEND (Unidad 1)
│   ├── App.jsx                   # Rutas + ProtectedRoute + AdminRoute
│   ├── config.js                 # API_URL = http://localhost:3000
│   ├── main.jsx                  # Entry point
│   ├── components/               # UI pages
│   │   ├── Login.jsx
│   │   ├── Dashboard.jsx         # Solo admin
│   │   ├── MiRed.jsx             # Solo admin — red multinivel
│   │   ├── Storefront.jsx        # Tienda
│   │   ├── Catalogo.jsx
│   │   ├── DetalleProducto.jsx
│   │   ├── Carrito.jsx
│   │   ├── Checkout.jsx
│   │   ├── Confirmacion.jsx
│   │   ├── Layout.jsx / Navbar.jsx / Sidebar.jsx
│   ├── hooks/useRed.js           # Carga /api/red?email= del usuario
│   ├── context/
│   │   ├── AuthContext.jsx       # isAuthenticated, user.rol
│   │   └── CartContext.jsx       # Carrito por email (localStorage)
│   ├── services/
│   │   ├── redService.js         # getRed(email)
│   │   ├── checkoutService.js    # enviarPedido → POST /api/checkout
│   │   └── productosService.js   # getProductos / getProductoById (fetch al backend)
│   └── data/
│       ├── productos.js          # Mock productos
│       └── red.js                # Mock red multinivel
├── multicatalogo-backend/        # BACKEND (Unidad 2)
│   ├── main.go                   # ConectarDB + Fiber + CORS → SetupRoutes
│   ├── go.mod
│   ├── docker-compose.yml        # PostgreSQL 16 + Redis + backend
│   ├── Dockerfile                # Imagen del backend
│   ├── .env / .env.example       # DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME
│   ├── config/db.go              # CargarEnv, ConectarDB → *pgxpool.Pool
│   ├── config/redis.go           # ConectarRedis → *redis.Client (nil = sin caché)
│   ├── cache/cache.go            # Interfaz CacheRepository + RedisCache, claves y TTL
│   ├── database/init.sql         # Tablas + datos semilla (incluye Fase 4)
│   ├── database/fase4.sql        # Migración Fase 4: stock, pedidos, pedido_items, comisiones
│   ├── repository/repository.go  # Interfaz DBRepository + PostgresRepository
│   ├── repository/pedidos.go     # CrearPedido (transacción ACID), CalcularComisiones
│   ├── models/models.go          # LoginRequest, Producto, Usuario, Referido, APIError
│   ├── controllers/
│   │   ├── authController.go     # Login + validación campos vacíos
│   │   ├── prodController.go     # GetProductos, GetProductoPorID
│   │   ├── redController.go      # RedController.GetRed
│   │   ├── checkoutController.go # CheckoutController.Checkout + goroutine de comisiones
│   │   ├── errores.go            # errorInterno (500)
│   │   └── healthController.go   # GET /api/health
│   └── routes/routes.go          # Registro de endpoints
├── PROJECT_MAP.md                # ← este archivo
├── REQUIREMENTS_CHECKLIST.md     # Estado de tareas
└── .windsurfrules                # Reglas del agente
text## Endpoints API (no cambiar firmas)
| Método | Ruta | Controller |
|--------|------|------------|
| POST | `/api/login` | authController.Login |
| GET | `/api/productos` | prodController.GetProductos |
| GET | `/api/productos/:id` | prodController.GetProductoPorID |
| GET | `/api/red` | RedController.GetRed |
| POST | `/api/checkout` | CheckoutController.Checkout |
| GET | `/api/health` | healthController.Health |

## Credenciales de prueba
- admin@upse.edu.ec / 123456 → rol: admin
- cliente@upse.edu.ec / 123456 → rol: cliente

## Reglas arquitectónicas
1. main.go arma las dependencias → routes → controllers (structs) → interfaces CacheRepository / DBRepository
2. Respuestas de error = `APIError` (status, message, error, details)
3. Frontend NO debe romperse: mismas firmas JSON de éxito