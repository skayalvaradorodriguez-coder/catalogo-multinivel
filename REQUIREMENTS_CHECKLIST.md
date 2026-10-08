# REQUIREMENTS_CHECKLIST — Catálogo Multinivel

## Estado General
- **Hito actual:** Unidad 2 — Persistencia PostgreSQL (post Fase 2)
- **Última tarea completada:** Fase 2 — PostgreSQL (Docker), config/db.go, init.sql, repository, /api/red
- **Siguiente tarea pendiente:** MiRed.jsx contra /api/red y errores estructurados en Login.jsx
- **Actualizado:** 2026-10-08

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

## Hito 4 — Integración Frontend ↔ Backend real
- [x] productosService.js: reemplazar mock por fetch a /api/productos
- [ ] Login.jsx: consumir errores estructurados (data.error / data.message)
- [x] DetalleProducto: usar GET /api/productos/:id
- [ ] MiRed.jsx: consumir GET /api/red en lugar de data/red.js
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