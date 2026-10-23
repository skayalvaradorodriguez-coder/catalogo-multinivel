# Despliegue gratis en la nube (un solo link)

La app queda publicada en una dirección como **https://multicatalogo.onrender.com**: el ingeniero
la abre en la computadora o en el celular (e incluso la instala como PWA), sin instalar nada.

| Pieza | Servicio gratis | Qué guarda |
|-------|-----------------|------------|
| App (React + API Go en el mismo servidor) | **Render**, servicio web *Free* (Docker) | El código |
| Base de datos PostgreSQL | **Neon**, plan *Free* (no vence, 0.5 GB) | Usuarios, productos, pedidos, comisiones |
| Caché Redis | **Render Key Value**, plan *Free* | Caché del catálogo y de la red |

No piden tarjeta de crédito.

> **Ten en cuenta:** en el plan gratis de Render el servidor **se duerme tras 15 min sin visitas**.
> La primera visita después tarda **~1 minuto** en despertar; luego va normal.
> Ábrelo tú unos minutos antes de que el ingeniero lo revise.

---

## Cómo funciona

```
Navegador ──https──► Render (Docker)
                      ├─ /            → la app React (carpeta web/, PWA)
                      ├─ /api/...     → API Go + Fiber
                      ├─► Neon (PostgreSQL, conexión cifrada SSL)
                      └─► Render Key Value (Redis)
```

- `Dockerfile` (en la raíz): compila el frontend (`npm run build`) y el backend (`go build`) y deja
  una imagen pequeña con los dos.
- `STATIC_DIR=/app/web`: el backend entrega la app además de la API, así todo vive en **un solo link**
  (sin problemas de CORS y con HTTPS, que la PWA necesita).
- `DATABASE_URL`: el backend usa la cadena de conexión de Neon (con `sslmode=require`).
  En tu computadora no cambia nada: si no existe `DATABASE_URL` sigue usando `DB_HOST`, `DB_USER`...
- `render.yaml`: describe todo lo que Render debe crear (Blueprint).

---

## Paso 1 — Subir el código a GitHub

Render publica la rama **main** del repositorio. Sube los cambios y únelos a main
(como siempre, por ramas):

```powershell
cd "C:\PROYECTOS FRAMEWORK\catalogo-multinivel"
git checkout pwa
git add .
git commit -m "Despliegue en la nube: Render + Neon"
git push origin pwa
```

Luego en GitHub: **Compare & pull request → Create pull request → Merge pull request**.

## Paso 2 — Crear la base de datos en Neon

1. Entra a **https://neon.tech** → *Sign up* → **Continuar con GitHub**.
2. *Create project*: nombre `multicatalogo`, versión de Postgres **16 o 17**, región
   **AWS Oeste de EE. UU. 2 (Oregón)** (la misma región que usa Render por defecto). → *Create*.
3. En el panel, botón **Connect** → copia la **Connection string**. Se ve así:
   ```
   postgresql://neondb_owner:XXXXXXXX@ep-algo-123456.us-west-2.aws.neon.tech/neondb?sslmode=require&channel_binding=require
   ```
   Guárdala, es la **DATABASE_URL** (no la compartas: tiene la contraseña).
4. Menú izquierdo **SQL Editor** → abre en tu computadora
   `multicatalogo-backend\database\init.sql`, copia **todo** su contenido, pégalo y pulsa **Run**.
   Debe terminar sin errores (crea las tablas, los 8 productos y los usuarios de prueba).
5. Comprueba en el mismo editor:
   ```sql
   SELECT email, rol FROM usuarios;
   ```
   Deben salir `admin@upse.edu.ec` y `cliente@upse.edu.ec`.

## Paso 3 — Publicar en Render

1. Entra a **https://render.com** → *Get Started* → **GitHub** (autoriza el acceso al repositorio
   `catalogo-multinivel`).
2. Arriba: **New + → Blueprint**.
3. Elige el repositorio `catalogo-multinivel`, rama **main** → Render lee el `render.yaml` y muestra:
   - `multicatalogo` (Web Service, Free)
   - `multicatalogo-cache` (Key Value, Free)
4. Te pide el valor de **DATABASE_URL**: pega la cadena de Neon del paso 2. (`JWT_SECRET` se genera solo.)
5. **Deploy Blueprint**. La primera compilación tarda **5–10 min** (ve el progreso en *Logs*).
   Al final debe aparecer:
   ```
   Conectado a PostgreSQL en ep-....neon.tech/neondb
   Sirviendo el frontend desde /app/web
   ```
6. Arriba del servicio `multicatalogo` está tu link: **https://multicatalogo.onrender.com**
   (si el nombre estaba ocupado, Render agrega letras, p. ej. `multicatalogo-x1y2.onrender.com`).

## Paso 4 — Verificar

| Prueba | Qué debe pasar |
|--------|----------------|
| Abrir `https://TU-LINK.onrender.com/api/health` | Responde JSON con `"status":"ok"` |
| Abrir `https://TU-LINK.onrender.com` | Aparece el login con tu nombre en el menú |
| Entrar con `cliente@upse.edu.ec` / `123456` | Se ve la tienda con los productos |
| Comprar un producto | El pedido aparece en *Mis pedidos* |
| Entrar con `admin@upse.edu.ec` / `123456` | Back-office: pedidos, red, comisiones, reportes |
| En el celular, menú ⋮ → *Instalar app* | Se instala la PWA |

Link para enviar al ingeniero, por ejemplo:

> App MultiCatálogo — Skay Alvarado: https://TU-LINK.onrender.com
> Cliente: cliente@upse.edu.ec / 123456 · Admin: admin@upse.edu.ec / 123456
> (Si tarda en abrir la primera vez, espere ~1 minuto: el servidor gratuito se está despertando.)

## Actualizar la app después

Cada `git push` (o *merge*) a **main** vuelve a desplegar solo (`autoDeploy: true`). Si cambias
la base de datos, ejecuta el script nuevo en el **SQL Editor** de Neon.

## Problemas comunes

| Síntoma | Solución |
|---------|----------|
| La página tarda ~1 min o muestra “Service waking up” | Normal en el plan gratis: espera y recarga |
| En *Logs*: `JWT_SECRET no está definido` | En el servicio → *Environment* → agrega `JWT_SECRET` con un texto de 32+ caracteres |
| En *Logs*: error de conexión a PostgreSQL | Revisa `DATABASE_URL` en *Environment* (cópiala de nuevo de Neon, sin espacios) |
| Login dice “credenciales inválidas” | No se ejecutó `init.sql` en Neon (paso 2.4) |
| Falla la compilación en *Logs* | Asegúrate de que `package-lock.json` y `multicatalogo-backend/go.sum` estén subidos a GitHub |
| Render pide tarjeta | Elegiste un plan pago: el servicio y el Key Value deben decir **Free** |

## Probar la imagen en tu computadora (opcional, requiere Docker)

```powershell
docker build -t multicatalogo .
docker run -p 3000:3000 -e JWT_SECRET=una-clave-de-prueba-larga -e DATABASE_URL="postgresql://postgres:postgres@host.docker.internal:5432/multicatalogo?sslmode=disable" multicatalogo
```
Abre http://localhost:3000 (app y API en el mismo puerto).
