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

## App instalable (PWA)

El frontend es una PWA: se instala como app y el catálogo se puede consultar sin conexión.
El service worker solo se activa con la build de producción o en `npm run dev`, y siempre en
`localhost` o con HTTPS.

**En la computadora** (con el backend corriendo):
```bash
npm run build
npm run preview
```
Abre http://localhost:4173 → menú de perfil → **Instalar app** (o el ícono de instalar en la barra de direcciones).

**En el celular** (necesita HTTPS): con `npm run preview` corriendo, en otra terminal:
```bash
npx cloudflared tunnel --url http://localhost:4173
```
Si el túnel repite errores `QUIC` o `unreachable network`, tu red bloquea UDP/IPv6; usa:
```bash
npx cloudflared tunnel --protocol http2 --edge-ip-version 4 --url http://127.0.0.1:4173
```
Abre en el celular la dirección `https://...trycloudflare.com` que aparece. En Android: menú ⋮ →
*Instalar app*. En iPhone (Safari): *Compartir → Agregar a pantalla de inicio*.
Para que funcione por el túnel, `VITE_API_URL` debe estar **vacía** en el `.env` de la raíz
(la app llama a `/api` y Vite lo reenvía al backend).

**Probar sin conexión:** DevTools → Application → Service workers → *Offline*, o apaga el backend:
el catálogo visitado y el carrito siguen funcionando; comprar y las vistas privadas piden conexión.

**Íconos:** se generan desde `public/logo.svg` con `npm run generar-iconos`.

## Con Docker (opcional)

```bash
cd multicatalogo-backend
docker compose up -d --build
```
Levanta PostgreSQL, Redis y el backend. El volumen nuevo ejecuta `init.sql` automáticamente.

## En la nube (gratis, un solo link)

La app se publica en **Render** (app + API en un solo servidor Docker) con la base de datos en
**Neon** y Redis en **Render Key Value**, todo en planes gratuitos. Paso a paso en
[`DESPLIEGUE.md`](DESPLIEGUE.md).

- `Dockerfile` (raíz): compila frontend y backend en una sola imagen.
- `render.yaml`: Blueprint con el servicio web y el Redis.
- Variables nuevas del backend: `DATABASE_URL` (cadena de Neon, tiene prioridad sobre `DB_*`) y
  `STATIC_DIR` (carpeta de la build del frontend que el backend entrega en `/`).

## Documentación

- Despliegue en la nube: [`DESPLIEGUE.md`](DESPLIEGUE.md)
- Mapa del proyecto y endpoints: [`PROJECT_MAP.md`](PROJECT_MAP.md)
- Estado de tareas: [`REQUIREMENTS_CHECKLIST.md`](REQUIREMENTS_CHECKLIST.md)
- Plan del back-office: [`PLAN_IMPLEMENTACION_U2.md`](PLAN_IMPLEMENTACION_U2.md)
- Análisis funcional: [`ANALISIS_FUNCIONAL_U2.md`](ANALISIS_FUNCIONAL_U2.md)
