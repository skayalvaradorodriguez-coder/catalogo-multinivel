# REQUIREMENTS_CHECKLIST — Catálogo Multinivel

## Estado General
- **Hito actual:** Unidad 2 — Backend robusto (post Fase 1)
- **Última tarea completada:** Fase 1 — APIError, validaciones Login, GetProductoPorID, /api/health
- **Siguiente tarea pendiente:** Capa repository + preparación persistencia relacional
- **Actualizado:** 2026-10-02

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

## Hito 3 — Backend Fase 2: Repositorio y persistencia (PENDIENTE)
- [ ] Crear capa `repository/` (interfaz + implementación en memoria)
- [ ] Mover listaProductos() al repositorio
- [ ] Inyectar repositorio en controllers (sin romper firmas HTTP)
- [ ] Preparar modelos para DB relacional (tags sql/json)
- [ ] Endpoint /api/red (red multinivel) si aplica al sílabo
- [ ] Persistencia real (SQLite/Postgres) según guía práctica

## Hito 4 — Integración Frontend ↔ Backend real
- [ ] productosService.js: reemplazar mock por fetch a /api/productos
- [ ] Login.jsx: consumir errores estructurados (data.error / data.message)
- [ ] DetalleProducto: usar GET /api/productos/:id
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