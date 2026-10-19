# REQUIREMENTS CHECKLIST — MultiCatálogo

> Fuente de verdad del progreso. Leer al inicio de cada sesión; marcar `- [x]`
> y actualizar "Estado General" al completar cualquier tarea.
>
> **Estructura:** el documento se divide en fases por **Unidad/Tema del sílabo**.
> El producto integrador es UNO: MultiCatálogo — una tienda cosmética con red
> de ventas multinivel (MLM): React (front) + Go/Fiber (núcleo transaccional, U2)
> + FastAPI (capa SaaS/multitenant, U3), desplegado contenerizado (U4).
>
> **Principios rectores (aplican a toda fase):**
> - Seguridad por defecto: ningún endpoint nuevo nace sin auth ni validación.
> - Inyección de dependencias: interfaces en los bordes; globals solo en el entry point.
> - Funciones puras para el dominio (comisiones, totales, validadores) — testeables sin I/O.
> - Mobile-first: UI a 360px, touch ≥44px, un campo por fila en formularios.
>
> **Adaptaciones de este repositorio** (decididas con el estudiante):
> - El frontend se mantiene en **JavaScript** (`.jsx`/`.js`), no en TypeScript.
> - Estructura: frontend en la raíz del repo y backend en `multicatalogo-backend/`
>   (equivale a `Unidad1_Frontend/Proyecto_base` y `Unidad2_Backend/multicatalogo-backend`).
> - PostgreSQL local en el puerto **5432** (sin Docker en la máquina de desarrollo);
>   `docker-compose.yml` queda listo para quien use contenedores.

## Estado General

| Ítem | Valor |
|---|---|
| **Hito actual** | **Unidades 1 y 2 completadas** + back-office del plan `PLAN_IMPLEMENTACION_U2.md` (fases 0-7) |
| **Última tarea completada** | Seguridad U2-T4 (bcrypt + JWT + RequireAuth/RequireAdmin + identidad por token) y back-office: registro con referido, CRUD productos, pedidos con estados, ledger de comisiones, liquidaciones, reportes, paginación (2026-10-19) |
| **Siguiente tarea pendiente** | U3-T1: scaffold FastAPI (capa SaaS) cuando el profesor la publique |
| **Bloqueadores** | Ninguno |

## Arquitectura objetivo del semestre

```
React SPA (:5173) ──► Caddy/Nginx (U4: proxy inverso + SSL)
                        ├─ /api/tx/*   → Go/Fiber  :3000  — núcleo transaccional (U2)
                        └─ /api/saas/* → FastAPI   :8000  — capa SaaS/multitenant (U3)
                                            │
                          PostgreSQL 16 (:5432) + Redis (:6379)
```

---

# UNIDAD 1 — FRAMEWORKS FRONTEND (SPA React + Vite + Tailwind)

## U1-T1 · Introducción a frameworks web y arquitectura de soluciones

- [x] Pila definida y workspace estructurado (frontend en la raíz, backend en `multicatalogo-backend/`)

## U1-T2 · Entorno SPA con Vite + Tailwind CSS

- [x] Entorno SPA con Vite + Tailwind CSS
- [x] `package.json` renombrado a `multicatalogo-frontend`
- [x] `VITE_API_URL` en `.env` + `import.meta.env` (`src/config.js`, ver `.env.example`) — cliente único en `services/http.js`

## U1-T3 · Componentes y gestión de estado global

- [x] `AuthContext` (sesión `{email, rol, nombre}`) y `CartContext` (carrito en localStorage por email)
- [x] Persistencia de sesión en `AuthContext` (`multicatalogo_sesion`: usuario + JWT en localStorage)
- [x] Contextos separados (`AuthContextValue.js` / `useAuth.jsx`, `CartContextValue.js` / `useCart.jsx`); `npm run lint` → 0 errores, 0 warnings

## U1-T4 · Consumo de API RESTful y enrutamiento

- [x] `services/productosService.js`, `redService.js` (+ pedidos, comisiones, reportes, usuarios, checkout) + guardas `ProtectedRoute`/`AdminRoute`
- [x] `services/http.js`: `apiFetch` único con `Authorization: Bearer` automático, `ApiError` con status y helpers de sesión; 401 cierra la sesión

## U1-T5 · Interfaces full-screen y flujos MLM (Práctica 03)

- [x] F8 Login con rol + redirección por rol (+ enlace "Crear cuenta" → `/registro`)
- [x] F1 Storefront `/tienda` full-screen | F3 Catálogo con búsqueda/filtros | F2 Detalle `/producto/:id` | F4 Lightbox
- [x] F5 Carrito + checkout + confirmación | F6 Árbol multinivel comisiones 10/5/2 % | F7 Dashboard KPIs
- [x] `Checkout.jsx` → `POST /api/checkout` real (sin simulación) + manejo de `APIError`
- [x] Audit mobile-first: 375px sin overflow en Dashboard/MiRed/Catálogo/Mis pedidos/Liquidaciones; botones de acción ≥44px (`Añadir al Carrito`, steppers de `Carrito`, `Cerrar Sesión`, menú, paginación)

---

# UNIDAD 2 — BACKEND ALTA TRANSACCIONALIDAD (Go/Fiber · "entornos bancarios")

## U2-T1 · Rendimiento, memoria y concurrencia

- [x] Goroutine `procesarComisiones` tras el COMMIT del pedido (contexto propio, no el del request) + `WaitGroup` para no cortarla al apagar

## U2-T2 · Microservicios y APIs robustas (Fiber)

- [x] Capas routes → controllers → repository; endpoints login/productos/red/health
- [x] `APIError{status,message,details}` en todos los errores + validación de entrada (400/404) centralizada en `validacion/`
- [x] Puerto de escucha configurable (`PORT`, defecto 3000)
- [x] `/api/health` real: `HealthController` con ping a Postgres (503 si cae) y Redis (`"degraded"` si cae)

## U2-T3 · Persistencia PostgreSQL + caché Redis

- [x] `docker-compose` Postgres 16 + Redis alpine + volumen + healthchecks (+ servicio backend)
- [x] `init.sql` (usuarios/productos/referidos + seed), repositorio con SQL real, árbol por `parent_id`
- [x] Cache-Aside: `productos:all` TTL 5 min, `red:tree:{email}` TTL 10 min, degradación sin caché
- [x] Índices: `referidos(parent_id)`, `pedidos(cliente_email)`, `pedido_items(pedido_id)`, `productos(categoria)` (migración `003_seguridad.sql` + init.sql)
- [x] Errores propagados: el repo devuelve `error`; `ErrNoEncontrado`→404 (401 en login), `ErrDuplicado`→409, validación→400, resto→500

## U2-T4 · Inyección de dependencias y microservicio seguro

- [x] Interfaces `DBRepository`/`CacheRepository`; `PostgresRepository`/`RedisCache`; structs de controladores; `NewControllers` en `main.go`
- [x] `POST /api/checkout` ACID: `FOR UPDATE` + descuento de stock + `pedidos`/`pedido_items`, ROLLBACK → 400/500, total recalculado en servidor
- [x] Invalidación de caché tras compra (`productos:all`, `red:tree:*`) + migración `002_checkout.sql`
- [x] bcrypt de passwords (pgcrypto `crypt()/gen_salt('bf')` en `003_seguridad.sql` + seed; login con `CompareHashAndPassword`)
- [x] JWT real (`JWT_SECRET` en env, claims `{sub,email,rol,iat,exp}` 2h, HS256) + `middleware.RequireAuth(secret)`/`RequireAdmin()`
- [x] Protegidos `/api/red` y `/api/checkout` con el middleware (401 sin token / token adulterado)
- [x] Checkout toma la identidad del TOKEN (`c.Locals` desde claims) — verificado: un `usuarioEmail` falso en el body NO se persiste ni atribuye comisiones
- [x] Errores sanitizados: `Details` solo con `APP_ENV=dev`; `ErrCantidadInvalida` propio (400 "La cantidad debe ser mayor que cero", distinto de stock)

## U2 · Back-office (PLAN_IMPLEMENTACION_U2.md)

- [x] Fase 0 — Migración `004_backoffice.sql` (usuarios, referidos.codigo, productos.activo, pedidos.estado, comisiones, liquidaciones)
- [x] Fase 1 — `POST /api/register` con `ref`, bcrypt, nodo en la red en transacción; 409/400; login rechaza inactivos (403); vista `/registro`
- [x] Fase 2 — CRUD productos admin (borrado lógico, invalidación de caché) + `/admin/productos`
- [x] Fase 3 — `/api/pedidos/mios`, `/api/pedidos`, detalle, estados, cancelación ACID (stock + comisiones revertidas) + `/mis-pedidos`, `/admin/pedidos`
- [x] Fase 4 — Ledger `comisiones` dentro de la transacción del pedido, `domain.RepartirComisiones` pura, `/api/red/mia`, `/api/comisiones/mias` + Mi Red personal con link
- [x] Fase 5 — Liquidaciones (`POST/GET /api/liquidaciones`, `/mias`) + `/admin/liquidaciones`
- [x] Fase 6 — Reportes `/api/reportes/*` + Dashboard real con filtro por fechas
- [x] Fase 7 — Paginación `{data,page,limit,total}`, `?q=` ILIKE, auditoría de rutas (escritura → `RequireAuth`, admin → `RequireAdmin`)
- [x] Extra del análisis: `GET/PATCH /api/usuarios` + `/admin/usuarios`; `Mailer` inyectado (`LogMailer`) para bienvenida y confirmación de pedido

## Pruebas automáticas

- [x] `go test ./...` — `domain` (reparto 10/5/2 %, redondeo, transiciones de estado) y `seguridad` (JWT válido/adulterado/otra clave/vencido, bcrypt)
- [x] `npm run lint` (0/0) y `npm run build`

---

# ROADMAP (fuera del alcance U1+U2)

- [ ] U3 — FastAPI SaaS/multitenant
- [ ] U4 — Proxy inverso + SSL + despliegue contenerizado + OWASP
- [ ] Post-sílabo — refresh tokens, Outbox persistente para comisiones, emails reales (SMTP), subida de imágenes, pasarela de pago

## Notas técnicas (no marcar)

- Credenciales: admin@upse.edu.ec / 123456 | cliente@upse.edu.ec / 123456 (ahora con hash bcrypt)
- Seed de la red: admin → "Tú" (código `TU`), cliente → "Luis Poveda" (código `LUIS`)
- Errores siempre con `models.APIError`; `Details` solo en `APP_ENV=dev`
- Un token sigue siendo válido hasta 2 h aunque la cuenta se desactive (no hay lista de revocación); el login y el checkout sí rechazan cuentas inactivas
