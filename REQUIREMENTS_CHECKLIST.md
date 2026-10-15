# REQUIREMENTS_CHECKLIST — Catálogo Multinivel

## Estado General
- **Hito actual:** Unidad 2 — cerrada (post Fase 4)
- **Última tarea completada:** Fase 4 — DI, checkout ACID, goroutine de comisiones, invalidación de caché
- **Siguiente tarea pendiente:** Según la siguiente fase del profesor
- **Actualizado:** 2026-10-15

---

## Hito 0 — Entorno y base
- [x] Proyecto React + Vite + Tailwind configurado
- [x] Backend Go + Fiber en puerto 3000
- [x] CORS apuntando a localhost:5173
- [x] Frontend consume API_URL desde config.js

## Hito 1 — Frontend (Unidad 1) — COMPLETADO
- [x] Login con roles admin / cliente
- [x] ProtectedRoute y AdminRoute
- [x] Dashboard (solo admin)
- [x] Mi Red multinivel (solo admin)
- [x] Storefront / Catálogo de productos
- [x] Detalle de producto por ID
- [x] Carrito (CartContext + localStorage por usuario)
- [x] Checkout y Confirmación
- [x] Layout + Navbar + Sidebar
- [x] servicios/productosService.js (mock listo para swap a fetch)

## Hito 2 — Backend Fase 1: Arquitectura y API robusta — COMPLETADO
- [x] Estructura APIError (status, message, error, details) + NewAPIError
- [x] Validación Login: email/password no vacíos → HTTP 400 APIError
- [x] GetProductoPorID: strconv error → HTTP 400 APIError
- [x] GetProductoPorID: ID inexistente → HTTP 404 APIError
- [x] healthController + GET /api/health → 200 ok
- [x] Rutas registradas en routes.go sin romper firmas existentes
- [x] Respuestas de éxito de login/productos intactas para el frontend

## Hito 3 — Backend Fase 2: Repositorio y persistencia PostgreSQL — COMPLETADO
- [x] `docker-compose.yml` con PostgreSQL 16 (base `multicatalogo`)
- [x] `config/db.go`: pool pgx, variables desde `.env`, cierre al apagar
- [x] `database/init.sql`: tablas usuarios, productos, referidos + datos semilla
- [x] Capa `repository/repository.go` con SQL (sin datos en memoria)
- [x] Controllers usan el repositorio (firmas HTTP intactas)
- [x] Modelos con tags json/db (Producto ampliado, Usuario, Referido)
- [x] Endpoint GET /api/red (árbol armado desde parent_id)

## Hito 3b — Backend Fase 3: Caché con Redis — COMPLETADO
- [x] `docker-compose.yml`: servicio `redis:alpine` (6379) y `backend` que depende de db y redis
- [x] `config/redis.go`: cliente go-redis v9 desde REDIS_URL o REDIS_ADDR (.env)
- [x] Cache-Aside en GetProductos: clave `productos:all`, TTL 5 min
- [x] Cache-Aside en GetRed: clave `red:tree:{email}`, TTL 10 min
- [x] Cabecera `X-Cache: HIT|MISS` para verlo en la pestaña Network
- [x] Si Redis no está disponible, la API sigue respondiendo desde PostgreSQL

## Hito 3c — Backend Fase 4: Checkout transaccional — COMPLETADO
- [x] Inyección de dependencias: interfaces `DBRepository` y `CacheRepository`, controladores como structs, armado en `main.go`
- [x] Sin variables globales de conexión (`config.ConectarDB` / `ConectarRedis` devuelven la conexión)
- [x] `POST /api/checkout` recibe el JSON de `Checkout.jsx` (cliente, método de pago, items)
- [x] Transacción ACID: `pedidos` + `pedido_items` + resta de stock; ROLLBACK ante error (400 datos / 500 infraestructura)
- [x] Goroutine de comisiones: 10% / 5% / 2% a los niveles ascendentes 1, 2 y 3 (tabla `comisiones`)
- [x] Invalidación de caché: `red:tree:{email}` del comprador y sus ascendentes, y `productos:all`
- [x] `database/fase4.sql` (migración) y `init.sql` actualizado
- [x] Simulación de fallas: `SIMULAR_FALLA_CHECKOUT` y `CHECKOUT_PAUSA_SEGUNDOS` en `.env`

## Hito 4 — Integración Frontend ↔ Backend real
- [x] productosService.js: reemplazar mock por fetch a /api/productos
- [ ] Login.jsx: consumir errores estructurados (data.error / data.message)
- [x] DetalleProducto: usar GET /api/productos/:id
- [x] MiRed.jsx y Dashboard.jsx: consumir GET /api/red?email= (hook useRed)
- [ ] Manejo de estados de carga y error en UI

## Hito 5 — Calidad y cierre
- [ ] Tests de endpoints críticos (health, login vacío, id inválido)
- [ ] Documentación mínima de API
- [ ] .windsurfrules + PROJECT_MAP + este checklist mantenidos al día

---

## Notas técnicas (no marcar)
- Credenciales: admin@upse.edu.ec / 123456 | cliente@upse.edu.ec / 123456
- Prohibido cambiar firmas: POST /api/login, GET /api/productos, estructura JSON de éxito
- Errores siempre con models.NewAPIError(...)