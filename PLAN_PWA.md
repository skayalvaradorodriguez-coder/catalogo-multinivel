# PLAN DE IMPLEMENTACIÓN — MultiCatálogo como PWA

> Plan checkeable para convertir el frontend (React 19 + Vite 8, JavaScript) en una
> **Aplicación Web Progresiva**: instalable en el celular y el escritorio, que abre
> sin conexión y avisa cuando hay una versión nueva.
>
> **Regla:** una fase se marca `[x]` solo cuando su criterio de verificación pasa.
> Al cerrar cada fase se actualizan `REQUIREMENTS_CHECKLIST.md` y `PROJECT_MAP.md`.
>
> **Principios (los mismos del checklist):**
> - Seguridad por defecto: **nunca** se guardan en caché respuestas con datos privados
>   (token, pedidos, comisiones) ni peticiones de escritura (login, checkout, admin).
> - Mobile-first: todo lo nuevo (banner, aviso de actualización, botón instalar) a 360 px y con toques ≥ 44 px.
> - El backend en Go **no cambia**: la PWA es una capa del frontend.

---

## Diagnóstico de partida (2026-10-22, rama `u2-backoffice`)

| Requisito PWA | Estado | Detalle |
|---|---|---|
| Diseño responsive / mobile-first | ✅ | Auditado a 375 px, botones ≥ 44 px, `viewport` presente |
| Estado local que sobrevive sin red | ✅ | Sesión JWT y carrito en `localStorage` |
| Cliente HTTP único | ✅ | `services/http.js` (`apiFetch`) — punto natural para detectar "sin conexión" |
| Manifiesto web | ❌ | No existe |
| Íconos 192 / 512 / maskable / apple | ❌ | Solo `favicon.svg` (logo de Vite, no de MultiCatálogo) |
| Service worker + caché | ❌ | No existe |
| Metadatos de `index.html` | ⚠️ | `lang="en"`, título `catalogo-multinivel`, sin `theme-color` ni etiquetas iOS |
| HTTPS | ⚠️ | Solo `localhost` / `127.0.0.1` cuentan como seguros; el celular por la IP de red no |
| Experiencia sin conexión | ❌ | Sin red, cada pantalla muestra un error genérico |

**Herramienta elegida:** `vite-plugin-pwa` 2.x (compatible con Vite 8; genera el manifiesto y el
service worker con Workbox) + `@vite-pwa/assets-generator` para los íconos.

---

## FASE PWA-0 — Identidad visual y metadatos

> La app tiene que presentarse como "MultiCatálogo", no como una plantilla de Vite.

- [x] Logo propio en `public/logo.svg` (cuadrado, fondo sólido, legible a 48 px) — reemplaza el `favicon.svg` de Vite
- [x] Íconos generados con `@vite-pwa/assets-generator`: `pwa-192x192.png`, `pwa-512x512.png`, `maskable-icon-512x512.png`, `apple-touch-icon-180x180.png`, `favicon.ico`
- [x] `index.html`: `lang="es"`, `<title>MultiCatálogo</title>`, `meta description`, `meta theme-color` (teal del sidebar `#0f766e`), `apple-touch-icon`, `apple-mobile-web-app-capable`, `apple-mobile-web-app-title`
- [x] Borrar los assets de plantilla que no se usan (`src/assets/react.svg`, `src/assets/vite.svg`, `public/icons.svg` si no se usa)

**Verificación:** la pestaña del navegador muestra el ícono y el título de MultiCatálogo; los PNG existen con su tamaño exacto.

**Archivos:** `public/logo.svg`, `public/*.png`, `index.html`, `pwa-assets.config.js`

---

## FASE PWA-1 — Manifiesto web

> Lo que el sistema operativo lee para instalar la app: nombre, colores, íconos y cómo abrirse.

- [x] Instalar `vite-plugin-pwa` (dev) y registrarlo en `vite.config.js`
- [x] Manifiesto: `name` "MultiCatálogo", `short_name` "MultiCatálogo", `description`, `lang` "es", `start_url` "/", `scope` "/", `display` "standalone", `orientation` "portrait", `theme_color`, `background_color`, `categories` ["shopping", "business"]
- [x] Íconos del manifiesto (192, 512 y 512 con `purpose: "maskable"`)
- [x] Atajos (`shortcuts`): Catálogo, Mis Pedidos y Mi Red
- [x] Capturas (`screenshots`) para el diálogo de instalación enriquecido: una vertical (móvil) y una horizontal (escritorio)

**Verificación:** DevTools → Application → **Manifest** sin errores ni advertencias; muestra nombre, colores e íconos.

**Archivos:** `vite.config.js`, `package.json`, `public/screenshots/*.png`

---

## FASE PWA-2 — Service worker y caché del "app shell"

> El service worker guarda HTML, JS, CSS e íconos: la app abre aunque no haya internet.

- [x] `VitePWA({ registerType: "prompt", ... })` con estrategia `generateSW` (Workbox)
- [x] Precache de todo lo que genera `npm run build` (`globPatterns: **/*.{js,css,html,svg,png,ico,woff2}`)
- [x] `navigateFallback: "/index.html"` para que las rutas de React Router (`/catalogo`, `/mis-pedidos`...) abran sin red
- [x] `navigateFallbackDenylist: [/^\/api\//]` — las llamadas a la API nunca devuelven el HTML de la app
- [x] `devOptions.enabled: true` para probar el service worker con `npm run dev`
- [x] `cleanupOutdatedCaches: true` para borrar cachés de versiones anteriores

**Verificación:** `npm run build && npm run preview` → DevTools → Application → **Service workers** "activated and running"; marcar **Offline** y recargar `/catalogo` → la app abre (sin datos todavía).

**Archivos:** `vite.config.js`, `src/main.jsx` (registro)

---

## FASE PWA-3 — Estrategias de caché para la API e imágenes

> Qué datos se pueden ver sin conexión y cuáles **nunca** se guardan.

| Petición | Estrategia | Motivo |
|---|---|---|
| `GET /api/productos`, `GET /api/productos/:id` | **StaleWhileRevalidate** (máx. 50 entradas, 24 h) | Catálogo público: se muestra al instante y se actualiza en segundo plano |
| Imágenes `picsum.photos` y `img` de productos | **CacheFirst** (máx. 200 entradas, 30 días, solo respuestas 0/200) | Las fotos no cambian; sin esto el catálogo sin red se ve sin imágenes |
| Fuentes / estáticos externos | **CacheFirst** | Igual que imágenes |
| `POST /api/login`, `/api/register`, `/api/checkout`, todo `PUT/PATCH/DELETE` | **NetworkOnly** | Escrituras: jamás desde caché |
| `/api/red/mia`, `/api/pedidos*`, `/api/comisiones*`, `/api/liquidaciones*`, `/api/usuarios*`, `/api/reportes*`, `/api/admin/*` | **NetworkOnly** | Datos privados con token: en un equipo compartido otro usuario podría verlos |
| `GET /api/health` | **NetworkOnly** | Debe reflejar el estado real |

- [x] `workbox.runtimeCaching` con las reglas de la tabla (por `urlPattern` y método)
- [x] La regla de la API funciona tanto con `VITE_API_URL` absoluta como con la ruta relativa `/api` (ver PWA-6)
- [x] Nombres de caché con prefijo `multicatalogo-` para poder borrarlos selectivamente

**Verificación:** con el catálogo visitado una vez, en modo **Offline** `/catalogo` y `/producto/1` muestran productos e imágenes; Application → **Cache Storage** no contiene ninguna URL de pedidos, red, comisiones ni admin.

**Archivos:** `vite.config.js`

---

## FASE PWA-4 — Experiencia sin conexión

> Sin red la app no "se rompe": dice qué pasa y qué se puede hacer.

- [x] Hook `src/hooks/useOnline.js` (`navigator.onLine` + eventos `online`/`offline`)
- [x] Componente `BannerSinConexion.jsx` en el `Layout`: franja fija "Sin conexión — mostrando datos guardados" (`role="status"`)
- [x] `apiFetch`: si `navigator.onLine === false`, lanzar `ApiError(0, "Sin conexión. Revisa tu internet e inténtalo de nuevo.")` sin intentar el `fetch`
- [x] `Checkout`: botón "Confirmar Pedido" deshabilitado sin conexión, con el motivo visible; el carrito se conserva
- [x] Vistas privadas (Mis Pedidos, Mi Red, admin): mensaje "Necesitas conexión para ver esta sección" en lugar del error genérico
- [x] Login y Registro: aviso de que se necesita conexión

**Verificación:** con DevTools en **Offline**: el banner aparece; el catálogo y el carrito funcionan; el checkout no deja confirmar; Mis Pedidos muestra el mensaje de conexión. Al volver **Online** el banner desaparece solo.

**Archivos:** `src/hooks/useOnline.js`, `src/components/BannerSinConexion.jsx`, `src/components/Layout.jsx`, `src/services/http.js`, `src/components/Checkout.jsx`

---

## FASE PWA-5 — Actualizaciones e instalación

> El usuario nunca se queda atrapado en una versión vieja, y sabe cómo instalar la app.

- [x] `AvisoActualizacion.jsx` con `useRegisterSW` (`virtual:pwa-register/react`): "Hay una nueva versión — Actualizar" y "Lista para usar sin conexión"
- [x] Comprobación periódica de actualizaciones (cada 60 min mientras la app está abierta)
- [x] Hook `useInstalarPWA.js` con el evento `beforeinstallprompt`: botón **"Instalar app"** en el menú de perfil (Android / Chrome / Edge); se oculta si ya está instalada (`display-mode: standalone`)
- [x] iPhone/iPad (sin `beforeinstallprompt`): instrucciones "Compartir → Agregar a pantalla de inicio" detectando Safari iOS
- [x] Ajustes en modo instalado: `env(safe-area-inset-*)` en Navbar/Sidebar para el notch

**Verificación:** cambiar un texto, `npm run build`, recargar → aparece el aviso y al pulsar "Actualizar" carga la versión nueva. En Chrome de escritorio el botón instala la app y abre en ventana propia sin barra de direcciones.

**Archivos:** `src/components/AvisoActualizacion.jsx`, `src/hooks/useInstalarPWA.js`, `src/components/Navbar.jsx`, `src/components/Layout.jsx`

---

## FASE PWA-6 — Seguridad, privacidad y HTTPS

> Un service worker solo funciona en contexto seguro, y una caché mal pensada filtra datos.

- [x] Al **cerrar sesión** (y cuando `apiFetch` recibe 401): borrar las cachés de datos `multicatalogo-api-*` (`caches.delete`) además de `localStorage`
- [x] Revisar que ninguna regla de caché coincida con peticiones que llevan `Authorization`
- [x] **Proxy de Vite** `/api → http://127.0.0.1:3000` (dev y preview) y `VITE_API_URL` vacía: la app y la API quedan en el **mismo origen**, sin CORS y sin "contenido mixto" cuando se sirva por HTTPS
- [x] HTTPS local para probar en el celular: certificado de confianza con **mkcert** (`server.https`) o túnel HTTPS (**cloudflared** / **ngrok**) hacia `npm run preview`
- [x] Documentar en `README.md` cómo probar la PWA en el celular

**Verificación:** cerrar sesión → Cache Storage sin cachés de API; abrir la app desde el celular por HTTPS → se puede instalar y funciona contra el backend de la computadora.

**Archivos:** `vite.config.js`, `src/services/http.js`, `src/context/AuthContext.jsx`, `.env.example`, `README.md`

> En la Unidad 4 este HTTPS lo dará el proxy inverso (Caddy/Nginx con SSL); aquí solo se prepara para que la PWA funcione igual en producción.

---

## FASE PWA-7 — Verificación final y documentación

- [x] DevTools → Application: Manifest sin errores, Service worker activo, app **instalable** (ícono de instalar en la barra de direcciones)
- [x] Lighthouse (categorías Performance, Accessibility, Best Practices, SEO) ≥ 90 en la build de producción. *Lighthouse 12 ya no tiene categoría "PWA"; la instalabilidad se verifica en el panel Application.*
- [x] Prueba real en Android (Chrome) por túnel HTTPS: instalar, abrir sin conexión (modo avión), ver catálogo
- [ ] Prueba en iPhone (Safari) — opcional
- [x] Prueba de privacidad: dos usuarios en el mismo navegador no ven datos del otro sin conexión
- [x] `npm run lint` 0/0 y `npm run build` sin advertencias del plugin
- [x] `PROJECT_MAP.md` y `REQUIREMENTS_CHECKLIST.md` actualizados
- [x] Capturas de evidencia en el celular (diálogo de instalación, ícono instalado, modo avión, catálogo)

---

## Estado de fases

| Fase | Entregable | Estado |
|---|---|---|
| PWA-0 | Logo, íconos y metadatos de `index.html` | ✅ 2026-10-22 |
| PWA-1 | Manifiesto con atajos y capturas | ✅ 2026-10-22 |
| PWA-2 | Service worker + app shell offline | ✅ 2026-10-22 |
| PWA-3 | Caché de catálogo e imágenes; datos privados nunca cacheados | ✅ 2026-10-22 |
| PWA-4 | Banner y mensajes sin conexión; checkout bloqueado offline | ✅ 2026-10-22 |
| PWA-5 | Aviso de actualización + botón "Instalar app" | ✅ 2026-10-22 |
| PWA-6 | Limpieza de caché al cerrar sesión, proxy `/api`, HTTPS por túnel | ✅ 2026-10-22 |
| PWA-7 | Verificación automatizada + prueba real en Android | ✅ 2026-10-22 (iPhone opcional) |

> Con PWA-0 a PWA-2 la app ya es **instalable**. PWA-3 y PWA-4 la hacen **útil sin
> conexión**; PWA-5 a PWA-7 la dejan lista para usarla en un celular real.

---

## Cómo quedó implementado (2026-10-22)

**Verificación automatizada** (Chromium con perfil normal, build de producción servida con `npm run preview`):

| Prueba | Resultado |
|---|---|
| `Page.getInstallabilityErrors` de Chrome | Sin errores → la app es instalable |
| Manifiesto (`Page.getAppManifest`) | Sin errores; íconos y capturas responden 200 |
| Service worker | `activated`; aviso "Lista para usar sin conexión" |
| Caché tras navegar como cliente | `multicatalogo-api-productos` (catálogo y detalle), `multicatalogo-imagenes`; **ninguna** URL de red, pedidos, comisiones, admin, login o checkout |
| Servidor web y backend **apagados** | La app abre, banner "Sin conexión", catálogo y detalle visibles, carrito funciona, checkout bloqueado, Mis Pedidos/Mi Red muestran "Necesitas conexión" |
| Cerrar sesión | Se borran las cachés `multicatalogo-api-*` y la sesión de `localStorage` |
| Nueva build publicada | Aparece "Hay una nueva versión"; sigue la versión anterior hasta pulsar "Actualizar" |
| Botón "Instalar app" | Chrome: abre el diálogo y luego se oculta · navegador sin soporte: no aparece · iPhone: muestra los pasos |
| Lighthouse 12 (`/login`, `/registro`) | Performance 99 · Accessibility 100 · Best Practices 100 · SEO 100 |

**Decisiones:**
- `clientsClaim: true`: en la primera visita el service worker controla la página sin recargar, así el catálogo se guarda desde el inicio. Las versiones nuevas siguen esperando a "Actualizar" (no se usa `skipWaiting`).
- Las lecturas del catálogo se piden **sin token** (`apiFetch(..., { publico: true })`); la regla de caché exige que la petición no lleve `Authorization`.
- `VITE_API_URL` vacía + proxy de Vite: la app y la API comparten origen (sirve igual por el túnel HTTPS).
- Se agregó `public/robots.txt` (SEO 100) y se asociaron las etiquetas de los formularios (`htmlFor`).

**Prueba en Android real (2026-10-22)**, por el túnel `npx cloudflared tunnel --protocol http2 --edge-ip-version 4 --url http://127.0.0.1:4173`:
- Chrome mostró el diálogo de instalación enriquecido (nombre, descripción y captura del manifiesto) y la app quedó en la pantalla de inicio con su ícono.
- En **modo avión** la app abrió con la franja "Sin conexión"; el Dashboard de admin mostró "Necesitas conexión para ver esta sección" (datos privados no cacheados) y el catálogo se vio con sus fotos.

**Nota de red:** si el túnel muestra `failed to accept QUIC stream` o `unreachable network`, la red bloquea UDP/IPv6: usar `--protocol http2 --edge-ip-version 4`.

**Opcional:** repetir la prueba en un iPhone (Safari → Compartir → Agregar a pantalla de inicio).
