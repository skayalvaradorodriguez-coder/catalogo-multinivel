# MultiCatálogo — Catálogo Multinivel

Tienda de cosméticos con red de ventas multinivel (MLM).
Frontend en React + Vite + Tailwind (JavaScript), backend en Go + Fiber, datos en PostgreSQL 16 y caché en Redis.

## Requisitos

- Node.js 20+
- Go 1.26+
- PostgreSQL 16 o superior (local) y Redis (en Windows: Memurai)

## Configuración inicial (una sola vez)

1. Copia `multicatalogo-backend/.env.example` como `multicatalogo-backend/.env` y ajusta
   `DB_PASSWORD` y `JWT_SECRET` (mínimo 16 caracteres).
2. Base de datos:
   - **Base nueva:** crea la base `multicatalogo` y ejecuta `multicatalogo-backend/database/init.sql`.
   - **Base de una versión anterior del proyecto:** ejecuta en orden
     `002_checkout.sql`, `003_seguridad.sql` y `004_backoffice.sql` (se pueden repetir sin problema).
3. Dependencias: `cd multicatalogo-backend && go mod tidy`, y en la raíz `npm install`.

## Cómo ejecutar

**Backend (puerto 3000)** — PostgreSQL y Redis deben estar corriendo:
```bash
cd multicatalogo-backend
go run .
```

**Frontend (puerto 5173)**, en otra terminal:
```bash
npm run dev
```

Abrir http://localhost:5173

| Rol | Email | Contraseña |
|-----|-------|------------|
| Administrador | admin@upse.edu.ec | 123456 |
| Cliente | cliente@upse.edu.ec | 123456 |

También puedes crear cuentas nuevas en `/registro` (con un código de referido, por ejemplo `/registro?ref=LUIS`).

## Pruebas

```bash
cd multicatalogo-backend && go test ./...
npm run lint
```

Prueba de humo de la API (con el backend corriendo, en Windows):
```powershell
cd multicatalogo-backend
powershell -ExecutionPolicy Bypass -File scripts\prueba-api.ps1 -Email tu@correo.com
```

## Con Docker (opcional)

```bash
cd multicatalogo-backend
docker compose up -d --build
```
Levanta PostgreSQL, Redis y el backend. El volumen nuevo ejecuta `init.sql` automáticamente.

## Documentación

- Mapa del proyecto y endpoints: [`PROJECT_MAP.md`](PROJECT_MAP.md)
- Estado de tareas: [`REQUIREMENTS_CHECKLIST.md`](REQUIREMENTS_CHECKLIST.md)
- Plan del back-office: [`PLAN_IMPLEMENTACION_U2.md`](PLAN_IMPLEMENTACION_U2.md)
- Análisis funcional: [`ANALISIS_FUNCIONAL_U2.md`](ANALISIS_FUNCIONAL_U2.md)
