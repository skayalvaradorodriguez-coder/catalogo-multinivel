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
| **Hito actual** | **Unidades 1 y 2 completadas** + back-office (`PLAN_IMPLEMENTACION_U2.md`) + **PWA completa** (`PLAN_PWA.md`, PWA-0 a PWA-7, probada en Android) |
| **Última tarea completada** | PWA: manifiesto, íconos, service worker, caché del catálogo, modo sin conexión, aviso de actualización, botón instalar, limpieza de caché al salir, proxy `/api` (2026-10-22) |
| **Siguiente tarea pendiente** | U3-T1: scaffold FastAPI (capa SaaS) cuando el profesor la publique · opcional: probar la PWA en un iPhone |
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

# PWA — APLICACIÓN WEB PROGRESIVA (PLAN_PWA.md)

> Convierte el frontend en una app **instalable** que abre **sin conexión** y avisa de
> versiones nuevas. Herramienta: `vite-plugin-pwa` 2.x (Workbox). El backend no cambia.
> Regla de privacidad: nunca se cachean datos con token ni peticiones de escritura.

## PWA-0 · Identidad visual y metadatos

- [x] Logo propio `public/logo.svg` (reemplaza el `favicon.svg` de Vite)
- [x] Íconos PNG 192, 512, maskable 512, apple-touch 180 y `favicon.ico` (`@vite-pwa/assets-generator`)
- [x] `index.html`: `lang="es"`, título "MultiCatálogo", `description`, `theme-color`, etiquetas iOS
- [x] Borrar assets de plantilla sin uso (`react.svg`, `vite.svg`, `icons.svg`)

## PWA-1 · Manifiesto web

- [x] `vite-plugin-pwa` instalado y registrado en `vite.config.js`
- [x] Manifiesto: nombre, `lang: es`, `start_url`/`scope` "/", `display: standalone`, colores, categorías
- [x] Íconos 192/512/maskable en el manifiesto
- [x] Atajos: Catálogo, Mis Pedidos, Mi Red
- [x] Capturas vertical y horizontal para el diálogo de instalación
- [x] Verificado: DevTools → Application → Manifest sin errores

## PWA-2 · Service worker y app shell

- [x] `generateSW` con `registerType: "prompt"` y precache de la build
- [x] `navigateFallback: /index.html` (rutas de React Router sin red) y `navigateFallbackDenylist: /api/`
- [x] `cleanupOutdatedCaches` y `devOptions` para probar en desarrollo
- [x] Verificado: con DevTools en Offline, `/catalogo` abre tras recargar

## PWA-3 · Estrategias de caché

- [x] `GET /api/productos[/:id]` → StaleWhileRevalidate (50 entradas, 24 h)
- [x] Imágenes de productos → CacheFirst (200 entradas, 30 días)
- [x] Login, registro, checkout, escrituras y rutas privadas/admin → NetworkOnly
- [x] Cachés con prefijo `multicatalogo-`
- [x] Verificado: catálogo e imágenes visibles offline; Cache Storage sin pedidos, red, comisiones ni admin

## PWA-4 · Experiencia sin conexión

- [x] Hook `useOnline` + `BannerSinConexion` en el `Layout`
- [x] `apiFetch` devuelve "Sin conexión..." sin intentar la petición
- [x] Checkout bloqueado offline (el carrito se conserva)
- [x] Mensaje "Necesitas conexión" en Mis Pedidos, Mi Red, admin, login y registro
- [x] Verificado: Offline → banner, catálogo y carrito funcionan; Online → banner desaparece

## PWA-5 · Actualizaciones e instalación

- [x] `AvisoActualizacion` ("Hay una nueva versión — Actualizar") con `useRegisterSW`
- [x] Revisión de actualizaciones cada 60 min
- [x] Botón "Instalar app" (`beforeinstallprompt`), oculto si ya está instalada
- [x] Instrucciones para iPhone (Compartir → Agregar a pantalla de inicio)
- [x] `safe-area-inset` en Navbar/Sidebar para el modo instalado
- [x] Verificado: nueva build → aviso → carga la versión nueva; la app se instala en escritorio

## PWA-6 · Seguridad, privacidad y HTTPS

- [x] Cerrar sesión / 401 borra las cachés de API (`caches.delete`)
- [x] Ninguna regla de caché coincide con peticiones con `Authorization`
- [x] Proxy de Vite `/api → 127.0.0.1:3000`: mismo origen, sin CORS ni contenido mixto
- [x] HTTPS para el celular con túnel (`npx cloudflared tunnel --url http://localhost:4173`) documentado en `README.md`
- [x] Verificado: tras cerrar sesión Cache Storage queda sin datos; la app se instala en el celular por HTTPS

## PWA-7 · Verificación final y documentación

- [x] Application: manifiesto OK, service worker activo, app instalable
- [x] Lighthouse ≥ 90 en Performance, Accessibility, Best Practices y SEO (build de producción)
- [x] Prueba real en Android (Chrome) por túnel HTTPS: instalación con diálogo enriquecido (descripción + captura), ícono en la pantalla de inicio, app abierta en modo avión con banner "Sin conexión" y catálogo con fotos (2026-10-22)
- [ ] Prueba en iPhone (Safari) — opcional, si se dispone de uno
- [x] Prueba de privacidad con dos usuarios en el mismo navegador
- [x] `npm run lint` 0/0, `npm run build` sin advertencias; PROJECT_MAP y checklist al día
- [x] Capturas de evidencia en el celular (diálogo de instalación, ícono instalado, modo avión, catálogo)

## NUBE · Publicación gratuita con un solo link (ver DESPLIEGUE.md)

- [x] Backend entrega la build del frontend (`STATIC_DIR`): SPA con fallback a `index.html`, 404 JSON en `/api/*`, encabezados de caché para `sw.js`, manifiesto y `/assets`
- [x] Conexión por `DATABASE_URL` con SSL (Neon); en local sigue funcionando con `DB_*`
- [x] `Dockerfile` único en la raíz (Node → Go → Alpine) y `render.yaml` (Blueprint: web Free + Key Value Free)
- [x] Prueba local del modo producción: un servidor en :3000 con app + API, base nueva cargada solo con `init.sql`, login, compra, caché HIT, service worker activo
- [ ] Base creada en Neon y `init.sql` ejecutado en su SQL Editor
- [ ] Blueprint desplegado en Render y link `https://….onrender.com` verificado (login, compra, admin, instalar en el celular)
- [ ] Link enviado al ingeniero

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
