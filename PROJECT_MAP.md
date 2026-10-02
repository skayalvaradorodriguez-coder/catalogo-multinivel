# PROJECT_MAP — Catálogo Multinivel

> Mapa ultra-compacto. Consultar ANTES de abrir cualquier archivo de código.

## Stack
| Capa | Tech | Puerto |
|------|------|--------|
| Frontend | React 19 + Vite 8 + Tailwind 4 + React Router 7 | 5173 |
| Backend | Go + Fiber v2 | 3000 |
| Auth | Fake JWT en memoria (admin/cliente) | — |
| Datos | Estáticos en memoria (próx. persistencia) | — |

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
│   ├── context/
│   │   ├── AuthContext.jsx       # isAuthenticated, user.rol
│   │   └── CartContext.jsx       # Carrito por email (localStorage)
│   ├── services/
│   │   └── productosService.js   # getProductos / getProductoById (mock→API)
│   └── data/
│       ├── productos.js          # Mock productos
│       └── red.js                # Mock red multinivel
├── multicatalogo-backend/        # BACKEND (Unidad 2)
│   ├── main.go                   # Fiber + CORS → SetupRoutes
│   ├── go.mod
│   ├── models/models.go          # LoginRequest, Producto, APIError, NewAPIError
│   ├── controllers/
│   │   ├── authController.go     # Login + validación campos vacíos
│   │   ├── prodController.go     # GetProductos, GetProductoPorID
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
| GET | `/api/health` | healthController.Health |

## Credenciales de prueba
- admin@upse.edu.ec / 123456 → rol: admin
- cliente@upse.edu.ec / 123456 → rol: cliente

## Reglas arquitectónicas
1. routes → controllers → models (próx. repository)
2. Respuestas de error = `APIError` (status, message, error, details)
3. Frontend NO debe romperse: mismas firmas JSON de éxito