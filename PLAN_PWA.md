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

- [ ] Logo propio en `public/logo.svg` (cuadrado, fondo sólido, legible a 48 px) — reemplaza el `favicon.svg` de Vite
- [ ] Íconos generados con `@vite-pwa/assets-generator`: `pwa-192x192.png`, `pwa-512x512.png`, `maskable-icon-512x512.png`, `apple-touch-icon-180x180.png`, `favicon.ico`
- [ ] `index.html`: `lang="es"`, `<title>MultiCatálogo</title>`, `meta description`, `meta theme-color` (teal del sidebar `#0f766e`), `apple-touch-icon`, `apple-mobile-web-app-capable`, `apple-mobile-web-app-title`
- [ ] Borrar los assets de plantilla que no se usan (`src/assets/react.svg`, `src/assets/vite.svg`, `public/icons.svg` si no se usa)

**Verificación:** la pestaña del navegador muestra el ícono y el título de MultiCatálogo; los PNG existen con su tamaño exacto.

**Archivos:** `public/logo.svg`, `public/*.png`, `index.html`, `pwa-assets.config.js`

---

## FASE PWA-1 — Manifiesto web

> Lo que el sistema operativo lee para instalar la app: nombre, colores, íconos y cómo abrirse.

- [ ] Instalar `vite-plugin-pwa` (dev) y registrarlo en `vite.config.js`
- [ ] Manifiesto: `name` "MultiCatálogo", `short_name` "MultiCatálogo", `description`, `lang` "es", `start_url` "/", `scope` "/", `display` "standalone", `orientation` "portrait", `theme_color`, `background_color`, `categories` ["shopping", "business"]
- [ ] Íconos del manifiesto (192, 512 y 512 con `purpose: "maskable"`)
- [ ] Atajos (`shortcuts`): Catálogo, Mis Pedidos y Mi Red
- [ ] Capturas (`screenshots`) para el diálogo de instalación enriquecido: una vertical (móvil) y una horizontal (escritorio)

**Verificación:** DevTools → Application → **Manifest** sin errores ni advertencias; muestra nombre, colores e íconos.

**Archivos:** `vite.config.js`, `package.json`, `public/screenshots/*.png`

---

## FASE PWA-2 — Service worker y caché del "app shell"

> El service worker guarda HTML, JS, CSS e íconos: la app abre aunque no haya internet.

- [ ] `VitePWA({ registerType: "prompt", ... })` con estrategia `generateSW` (Workbox)
- [ ] Precache de todo lo que genera `npm run build` (`globPatterns: **/*.{js,css,html,svg,png,ico,woff2}`)
- [ ] `navigateFallback: "/index.html"` para que las rutas de React Router (`/catalogo`, `/mis-pedidos`...) abran sin red
- [ ] `navigateFallbackDenylist: [/^\/api\//]` — las llamadas a la API nunca devuelven el HTML de la app
- [ ] `devOptions.enabled: true` para probar el service worker con `npm run dev`
- [ ] `cleanupOutdatedCaches: true` para borrar cachés de versiones anteriores

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

- [ ] `workbox.runtimeCaching` con las reglas de la tabla (por `urlPattern` y método)
- [ ] La regla de la API funciona tanto con `VITE_API_URL` absoluta como con la ruta relativa `/api` (ver PWA-6)
- [ ] Nombres de caché con prefijo `multicatalogo-` para poder borrarlos selectivamente

**Verificación:** con el catálogo visitado una vez, en modo **Offline** `/catalogo` y `/producto/1` muestran productos e imágenes; Application → **Cache Storage** no contiene ninguna URL de pedidos, red, comisiones ni admin.

**Archivos:** `vite.config.js`

---

## FASE PWA-4 — Experiencia sin conexión

> Sin red la app no "se rompe": dice qué pasa y qué se puede hacer.

- [ ] Hook `src/hooks/useOnline.js` (`navigator.onLine` + eventos `online`/`offline`)
- [ ] Componente `BannerSinConexion.jsx` en el `Layout`: franja fija "Sin conexión — mostrando datos guardados" (`role="status"`)
- [ ] `apiFetch`: si `navigator.onLine === false`, lanzar `ApiError(0, "Sin conexión. Revisa tu internet e inténtalo de nuevo.")` sin intentar el `fetch`
- [ ] `Checkout`: botón "Confirmar Pedido" deshabilitado sin conexión, con el motivo visible; el carrito se conserva
- [ ] Vistas privadas (Mis Pedidos, Mi Red, admin): mensaje "Necesitas conexión para ver esta sección" en lugar del error genérico
- [ ] Login y Registro: aviso de que se necesita conexión

**Verificación:** con DevTools en **Offline**: el banner aparece; el catálogo y el carrito funcionan; el checkout no deja confirmar; Mis Pedidos muestra el mensaje de conexión. Al volver **Online** el banner desaparece solo.

**Archivos:** `src/hooks/useOnline.js`, `src/components/BannerSinConexion.jsx`, `src/components/Layout.jsx`, `src/services/http.js`, `src/components/Checkout.jsx`

---

## FASE PWA-5 — Actualizaciones e instalación

> El usuario nunca se queda atrapado en una versión vieja, y sabe cómo instalar la app.

- [ ] `AvisoActualizacion.jsx` con `useRegisterSW` (`virtual:pwa-register/react`): "Hay una nueva versión — Actualizar" y "Lista para usar sin conexión"
- [ ] Comprobación periódica de actualizaciones (cada 60 min mientras la app está abierta)
- [ ] Hook `useInstalarPWA.js` con el evento `beforeinstallprompt`: botón **"Instalar app"** en el menú de perfil (Android / Chrome / Edge); se oculta si ya está instalada (`display-mode: standalone`)
- [ ] iPhone/iPad (sin `beforeinstallprompt`): instrucciones "Compartir → Agregar a pantalla de inicio" detectando Safari iOS
- [ ] Ajustes en modo instalado: `env(safe-area-inset-*)` en Navbar/Sidebar para el notch

**Verificación:** cambiar un texto, `npm run build`, recargar → aparece el aviso y al pulsar "Actualizar" carga la versión nueva. En Chrome de escritorio el botón instala la app y abre en ventana propia sin barra de direcciones.

**Archivos:** `src/components/AvisoActualizacion.jsx`, `src/hooks/useInstalarPWA.js`, `src/components/Navbar.jsx`, `src/components/Layout.jsx`

---

## FASE PWA-6 — Seguridad, privacidad y HTTPS

> Un service worker solo funciona en contexto seguro, y una caché mal pensada filtra datos.

- [ ] Al **cerrar sesión** (y cuando `apiFetch` recibe 401): borrar las cachés de datos `multicatalogo-api-*` (`caches.delete`) además de `localStorage`
- [ ] Revisar que ninguna regla de caché coincida con peticiones que llevan `Authorization`
- [ ] **Proxy de Vite** `/api → http://127.0.0.1:3000` (dev y preview) y `VITE_API_URL` vacía: la app y la API quedan en el **mismo origen**, sin CORS y sin "contenido mixto" cuando se sirva por HTTPS
- [ ] HTTPS local para probar en el celular: certificado de confianza con **mkcert** (`server.https`) o túnel HTTPS (**cloudflared** / **ngrok**) hacia `npm run preview`
- [ ] Documentar en `README.md` cómo probar la PWA en el celular

**Verificación:** cerrar sesión → Cache Storage sin cachés de API; abrir la app desde el celular por HTTPS → se puede instalar y funciona contra el backend de la computadora.

**Archivos:** `vite.config.js`, `src/services/http.js`, `src/context/AuthContext.jsx`, `.env.example`, `README.md`

> En la Unidad 4 este HTTPS lo dará el proxy inverso (Caddy/Nginx con SSL); aquí solo se prepara para que la PWA funcione igual en producción.

---

## FASE PWA-7 — Verificación final y documentación

- [ ] DevTools → Application: Manifest sin errores, Service worker activo, app **instalable** (ícono de instalar en la barra de direcciones)
- [ ] Lighthouse (categorías Performance, Accessibility, Best Practices, SEO) ≥ 90 en la build de producción. *Lighthouse 12 ya no tiene categoría "PWA"; la instalabilidad se verifica en el panel Application.*
- [ ] Prueba real en Android (Chrome) e iPhone (Safari): instalar, abrir sin conexión, ver catálogo, recibir aviso de actualización
- [ ] Prueba de privacidad: dos usuarios en el mismo navegador no ven datos del otro sin conexión
- [ ] `npm run lint` 0/0 y `npm run build` sin advertencias del plugin
- [ ] `PROJECT_MAP.md` y `REQUIREMENTS_CHECKLIST.md` actualizados; capturas de evidencia (instalación, modo offline, aviso de actualización)

---

## Estado de fases

| Fase | Entregable | Estado |
|---|---|---|
| PWA-0 | Logo, íconos y metadatos de `index.html` | ⏳ Pendiente |
| PWA-1 | Manifiesto con atajos y capturas | ⏳ Pendiente |
| PWA-2 | Service worker + app shell offline | ⏳ Pendiente |
| PWA-3 | Caché de catálogo e imágenes; datos privados nunca cacheados | ⏳ Pendiente |
| PWA-4 | Banner y mensajes sin conexión; checkout bloqueado offline | ⏳ Pendiente |
| PWA-5 | Aviso de actualización + botón "Instalar app" | ⏳ Pendiente |
| PWA-6 | Limpieza de caché al cerrar sesión, proxy `/api`, HTTPS local | ⏳ Pendiente |
| PWA-7 | Verificación en dispositivos y documentación | ⏳ Pendiente |

> Con PWA-0 a PWA-2 la app ya es **instalable**. PWA-3 y PWA-4 la hacen **útil sin
> conexión**; PWA-5 a PWA-7 la dejan lista para usarla en un celular real.
