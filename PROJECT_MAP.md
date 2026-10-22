# PROJECT MAP — MultiCatálogo (catalogo-multinivel)

> Mapa arquitectónico del repositorio. Leer ESTE archivo antes de explorar código.
> Proyecto didáctico "MultiCatálogo": tienda de cosméticos con red de ventas
> multinivel (MLM). Asignatura: Framework Programación Web (UPSE, 2026-02).

## Visión general

```
React SPA (:5173) ──apiFetch + JWT──► Go/Fiber API (:3000, env PORT) ──pgx────► PostgreSQL 16 (:5432)
                                            │
                                            └──go-redis──► Redis (:6379) — caché de lecturas
```

## Documentos raíz (meta, no código)

| Archivo | Contenido |
|---|---|
| `PROJECT_MAP.md` | Este archivo. Índice de navegación del proyecto. |
| `REQUIREMENTS_CHECKLIST.md` | Estado por Unidad/Tema del sílabo (U1 SPA → U2 Go/Fiber → U3 → U4) + back-office. |
| `PLAN_IMPLEMENTACION_U2.md` | Plan del profesor (fases 0-7 del back-office) con notas de cómo quedó en este repo. |
| `PLAN_PWA.md` | Plan para convertir el frontend en PWA (fases PWA-0 a PWA-7): manifiesto, service worker, caché, offline, instalación, HTTPS. |
| `ANALISIS_FUNCIONAL_U2.md` | Análisis de brecha funcional del profesor + estado de cada punto en este repo. |
| `README.md` | Cómo instalar, migrar la base y correr el proyecto. |
| `.windsurfrules` | Reglas del agente IA (ahorro de tokens, navegación, seguimiento). |

## Frontend (raíz del repo) — SPA

Stack: React 19 + JavaScript + Vite 8 + Tailwind CSS 4 + React Router 7.
Comandos: `npm run dev` (:5173) | `npm run build` | `npm run lint`.
Config: `.env` con `VITE_API_URL` (ver `.env.example`; por defecto `http://localhost:3000`).

| Ruta (`src/`) | Responsabilidad |
|---|---|
| `App.jsx` | Router; guardas `ProtectedRoute`/`AdminRoute`; `CartBoundary` (carrito por usuario). |
| `config.js` | `API_URL` desde `import.meta.env.VITE_API_URL`. |
| `services/http.js` | `apiFetch` único: Bearer automático, `ApiError{status,message,details}`, sesión en localStorage, 401 → cierra sesión. |
| `services/*Service.js` | `productos` (público + admin), `red`, `checkout`, `pedidos`, `comisiones` (+ liquidaciones), `reportes`, `usuarios` (login/registro/admin). |
| `hooks/useApi.js` | Llamada a la API con `{data, isLoading, error, recargar}`. |
| `hooks/useRed.js` | Red del usuario (`GET /api/red/mia`). |
| `context/AuthContext.jsx` | Sesión `{usuario, token}` persistida (`multicatalogo_sesion`). |
| `context/CartContext.jsx` | Carrito por usuario en localStorage (`multicatalogo_carrito_<email>`). |
| `data/productos.js` | Lista `categorias` (los productos vienen de la API). |
| `data/red.js` | Funciones puras: comisiones estimadas 10/5/2 %, totales, nivel alcanzado. |
| `data/formato.js` | `dinero()`, `fecha()`, estados de pedido. |
| `components/Login.jsx` `Registro.jsx` | Login (APIError) y registro público con `?ref=` prellenado. |
| `components/Storefront.jsx` `Catalogo.jsx` `DetalleProducto.jsx` | Tienda, catálogo con filtros, detalle + lightbox. |
| `components/Carrito.jsx` `Checkout.jsx` `Confirmacion.jsx` | Compra → `POST /api/checkout` (identidad por token). |
| `components/MisPedidos.jsx` `DetallePedido.jsx` | `/mis-pedidos`: historial paginado + detalle expandible. |
| `components/MiRed.jsx` `LinkReferido.jsx` `MisComisiones.jsx` | `/mi-red` (todos): subárbol propio, link de invitación con copiar, ledger. |
| `components/Dashboard.jsx` | `/` (admin): KPIs de `/api/reportes/*` con filtro de fechas. |
| `components/admin/AdminProductos.jsx` `FormularioProducto.jsx` | `/admin/productos`: tabla paginada, búsqueda, alta/edición, activar/desactivar. |
| `components/admin/AdminPedidos.jsx` | `/admin/pedidos`: filtro por estado + cambio de estado. |
| `components/admin/AdminUsuarios.jsx` | `/admin/usuarios`: rol y activación de cuentas. |
| `components/admin/AdminLiquidaciones.jsx` | `/admin/liquidaciones`: deuda por distribuidor + liquidar + historial. |
| `components/Paginacion.jsx` `EstadoPedido.jsx` `MensajeError.jsx` | UI compartida. |
| `components/Layout.jsx` `Navbar.jsx` `Sidebar.jsx` | Shell: menú por rol, perfil con mis pedidos / mi red / cerrar sesión. |

## multicatalogo-backend — API REST

Stack: Go 1.26 + Fiber v2 + pgx v5 (pool) + godotenv + golang-jwt v5 + bcrypt + PostgreSQL 16 + Redis (go-redis v9).
Comandos: `go run .` (:3000, `PORT`) | `go build ./...` | `go test ./...` | `docker compose up -d`.
Capas: `routes` (+ `middleware`) → `controllers` → `repository` / `cache` (interfaces) → PostgreSQL / Redis.

| Ruta | Responsabilidad |
|---|---|
| `main.go` | Entry point: `CargarConfig` → `ConectarDB` + `ConectarRedis` → `NuevoPostgresRepository` + `NuevaRedisCache` → `NewControllers` → `SetupRoutes`; CORS 5173-5175; apagado ordenado. |
| `config/config.go` | `Config{Puerto, Entorno, JWTSecret, JWTTTL, SimularFallaCheckout, PausaCheckout}` desde `.env`. |
| `config/db.go` `config/redis.go` | Pool pgx y cliente Redis (nil si Redis no responde). |
| `seguridad/` | `HashPassword`/`VerificarPassword` (bcrypt) y `GenerarToken`/`ValidarToken` (JWT HS256, 2 h). |
| `middleware/auth.go` | `RequireAuth(secret)`, `RequireAdmin()`, `UsuarioActual(c)` (claims en `c.Locals`). |
| `domain/` | Reglas puras: `RepartirComisiones(total, tasas)`, `TransicionValida(actual, nuevo)`. Con tests. |
| `validacion/` | Validación central de entrada (`Requerido`, `Email`, `MinLongitud`, `Positivo`...). |
| `notificaciones/mailer.go` | Interfaz `Mailer` + `LogMailer` (simula correos en el log). |
| `models/` | `models.go` (Usuario, Producto, Referido, `Pagina[T]`, APIError), `pedido.go`, `backoffice.go` (Comision, Liquidacion, reportes). |
| `cache/cache.go` | Interfaz `CacheRepository` + `RedisCache` (Obtener/Guardar/Eliminar/EliminarPatron/Ping, nil-safe). |
| `repository/repository.go` | Interfaz `DBRepository`, errores (`ErrNoEncontrado`, `ErrDuplicado`, `ErrCantidadInvalida`, `ErrValidacion`), paginación. |
| `repository/usuarios.go` | Login, registro transaccional con código de referido, administración. |
| `repository/productos.go` | Catálogo público (activos, `?q=`), admin paginado, crear/editar/desactivar. |
| `repository/red.go` | Árbol global y subárbol del usuario (CTE recursivo, nivel relativo). |
| `repository/pedidos.go` | `CrearPedido` ACID (+ ledger), `SumarVentas`, listados, detalle, `CambiarEstadoPedido` (cancelación con stock + clawback). |
| `repository/comisiones.go` | Estado de cuenta y `Liquidar` (FOR UPDATE, idempotente). |
| `repository/reportes.go` | Resumen, top productos, top distribuidores, deuda de comisiones. |
| `controllers/controllers.go` | `Controllers` + `NewControllers(repo, cache, mailer, cfg)` + respuestas de error comunes. |
| `controllers/*Controller.go` | health, auth, usuario, prod, red, checkout, pedido, comision, liquidacion, reporte. |
| `routes/routes.go` | `SetupRoutes(app, ctl, jwtSecret)` con `RequireAuth`/`RequireAdmin` por ruta. |
| `database/init.sql` | Instalación limpia: todas las tablas + seed (passwords con bcrypt). |
| `database/002_checkout.sql` `003_seguridad.sql` `004_backoffice.sql` | Migraciones idempotentes para una base existente (aplicar en orden). |
| `docker-compose.yml` `Dockerfile` | Postgres 16 + Redis + backend con healthchecks. |
| `.env` (gitignored) / `.env.example` | `PORT`, `APP_ENV`, `JWT_SECRET`, `DB_*`, `REDIS_*`, simulación de fallas. |

## Endpoints

| Método | Ruta | Acceso |
|---|---|---|
| GET | `/api/health` | público |
| POST | `/api/login` · `/api/register` | público |
| GET | `/api/productos[?q=]` · `/api/productos/:id` | público |
| GET | `/api/red/mia` | token |
| POST | `/api/checkout` | token |
| GET | `/api/pedidos/mios` · `/api/pedidos/:id` (dueño o admin) | token |
| GET | `/api/comisiones/mias` · `/api/liquidaciones/mias` | token |
| GET | `/api/red` | admin |
| GET | `/api/admin/productos?page&limit&q` | admin |
| POST/PUT/DELETE | `/api/productos[/:id]` | admin |
| GET | `/api/pedidos?page&limit&estado` | admin |
| PATCH | `/api/pedidos/:id/estado` | admin |
| GET/PATCH | `/api/usuarios[/:id]` | admin |
| POST/GET | `/api/liquidaciones` | admin |
| GET | `/api/reportes/resumen?desde&hasta` · `top-productos` · `top-referidos` · `comisiones-pendientes` | admin |

## Datos y contratos clave

- Errores API: `{"status":N,"message":"...","error":"...","details"?}`; `details` solo con `APP_ENV=dev`.
- Login / registro OK: `{"token","email","rol","nombre"}`. Cuentas: `admin@upse.edu.ec` / `cliente@upse.edu.ec` — `123456`.
- Checkout: `POST /api/checkout` `{cliente:{nombre,email,direccion,ciudad,telefono}, metodoPago, items:[{id,cantidad}]}` → `201` pedido con `numero`, `total`, `items` (precios del servidor). El comprador sale del token.
- Listados: `{"data":[...],"page","limit","total"}`.
- Red: raíz id 0 "Tú" (código `TU`); admin → nodo 0, cliente → nodo 2 "Luis Poveda" (`LUIS`). Registro sin código cuelga de la raíz.
- Comisiones: 10/5/2 % a los 3 patrocinadores sobre el total, en `comisiones` dentro de la transacción del pedido. Estados `pendiente → pagada` (liquidación) o `revertida` (cancelación). Si se cancela un pedido ya liquidado, se registra un cargo negativo (clawback).
- Estados del pedido: `pendiente → preparando → enviado → entregado`; `cancelado` desde pendiente o preparando.

## Notas operativas

- Caché: `productos:all` (5 min), `red:tree:{email}` y `red:tree:global` (10 min). Checkout, cancelación y registro borran `red:tree:*`; las escrituras de productos borran `productos:all`. Sin Redis, la API consulta PostgreSQL y `/api/health` responde `degraded`.
- Goroutine tras el COMMIT: suma `ventas` del nodo del comprador, invalida caché y "envía" el correo (log). El ledger ya quedó escrito en la transacción.
- Demostración ACID: `SIMULAR_FALLA_CHECKOUT=true` o `CHECKOUT_PAUSA_SEGUNDOS=N` en `.env` (reiniciar el backend).
