import process from "node:process";
import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import { VitePWA } from "vite-plugin-pwa";

// Colores de la marca (los mismos del sidebar).
const COLOR_TEMA = "#0f766e";
const COLOR_FONDO = "#f8fafc";

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  // El navegador llama a /api en el MISMO origen que la app y Vite lo reenvía al backend en Go.
  // Así no hay CORS ni "contenido mixto" cuando la app se sirve por HTTPS (túnel o proxy de la U4).
  const proxy = {
    "/api": { target: env.VITE_PROXY_TARGET || "http://127.0.0.1:3000", changeOrigin: true },
  };
  // Dominios de túneles HTTPS para probar la PWA en el celular (cloudflared / ngrok).
  const allowedHosts = [".trycloudflare.com", ".ngrok-free.app", ".ngrok.app"];

  return {
    server: { host: true, proxy, allowedHosts },
    preview: { host: true, proxy, allowedHosts },
    plugins: [
      react(),
      tailwindcss(),
      VitePWA({
        // "prompt": la versión nueva espera a que el usuario pulse "Actualizar" (AvisoActualizacion.jsx).
        registerType: "prompt",
        includeAssets: ["favicon.ico", "logo.svg", "apple-touch-icon-180x180.png", "robots.txt"],
        devOptions: { enabled: true, type: "module", navigateFallback: "index.html" },

        manifest: {
          id: "/",
          name: "MultiCatálogo",
          short_name: "MultiCatálogo",
          description:
            "Tienda de cosméticos con red de ventas multinivel: compra, invita y gana comisiones en tres niveles.",
          lang: "es",
          dir: "ltr",
          start_url: "/",
          scope: "/",
          display: "standalone",
          orientation: "portrait",
          theme_color: COLOR_TEMA,
          background_color: COLOR_FONDO,
          categories: ["shopping", "business"],
          icons: [
            { src: "pwa-64x64.png", sizes: "64x64", type: "image/png" },
            { src: "pwa-192x192.png", sizes: "192x192", type: "image/png" },
            { src: "pwa-512x512.png", sizes: "512x512", type: "image/png" },
            { src: "maskable-icon-512x512.png", sizes: "512x512", type: "image/png", purpose: "maskable" },
          ],
          shortcuts: [
            {
              name: "Catálogo",
              short_name: "Catálogo",
              url: "/catalogo",
              icons: [{ src: "pwa-192x192.png", sizes: "192x192", type: "image/png" }],
            },
            {
              name: "Mis Pedidos",
              short_name: "Pedidos",
              url: "/mis-pedidos",
              icons: [{ src: "pwa-192x192.png", sizes: "192x192", type: "image/png" }],
            },
            {
              name: "Mi Red",
              short_name: "Mi Red",
              url: "/mi-red",
              icons: [{ src: "pwa-192x192.png", sizes: "192x192", type: "image/png" }],
            },
          ],
          screenshots: [
            {
              src: "screenshots/movil-tienda.png",
              sizes: "390x844",
              type: "image/png",
              form_factor: "narrow",
              label: "Tienda MultiCatálogo en el celular",
            },
            {
              src: "screenshots/escritorio-catalogo.png",
              sizes: "1280x800",
              type: "image/png",
              form_factor: "wide",
              label: "Catálogo de productos en el escritorio",
            },
          ],
        },

        workbox: {
          // App shell: todo lo que genera "npm run build" queda disponible sin conexión.
          globPatterns: ["**/*.{js,css,html,svg,png,ico,woff2}"],
          globIgnores: ["screenshots/**"],
          cleanupOutdatedCaches: true,
          // La primera vez, el service worker toma control de la página sin esperar a recargar,
          // así el catálogo empieza a guardarse desde la primera visita. (Las versiones nuevas
          // siguen esperando a que el usuario pulse "Actualizar".)
          clientsClaim: true,
          // Las rutas de React Router (/catalogo, /mis-pedidos...) abren sin red...
          navigateFallback: "/index.html",
          // ...pero las llamadas a la API nunca reciben el HTML de la app.
          navigateFallbackDenylist: [/^\/api\//],

          // Caché en tiempo de ejecución (PWA-3). Lo que no coincide con ninguna regla va
          // directo a la red: login, checkout, escrituras y TODAS las rutas privadas o de admin.
          runtimeCaching: [
            {
              // Catálogo público: se muestra al instante y se actualiza en segundo plano.
              // Solo GET y solo sin token: una respuesta con Authorization nunca se guarda.
              urlPattern: ({ url, request }) =>
                request.method === "GET" &&
                !request.headers.has("Authorization") &&
                /^\/api\/productos(\/\d+)?$/.test(url.pathname),
              handler: "StaleWhileRevalidate",
              options: {
                cacheName: "multicatalogo-api-productos",
                expiration: { maxEntries: 50, maxAgeSeconds: 24 * 60 * 60 },
                cacheableResponse: { statuses: [200] },
              },
            },
            {
              // Fotos de productos (picsum.photos u otro dominio): no cambian, se sirven desde caché.
              urlPattern: ({ request }) => request.destination === "image",
              handler: "CacheFirst",
              options: {
                cacheName: "multicatalogo-imagenes",
                expiration: { maxEntries: 200, maxAgeSeconds: 30 * 24 * 60 * 60 },
                cacheableResponse: { statuses: [0, 200] },
              },
            },
            {
              // Fuentes y hojas de estilo externas.
              urlPattern: ({ request }) => request.destination === "font" || request.destination === "style",
              handler: "CacheFirst",
              options: {
                cacheName: "multicatalogo-estaticos",
                expiration: { maxEntries: 30, maxAgeSeconds: 365 * 24 * 60 * 60 },
                cacheableResponse: { statuses: [0, 200] },
              },
            },
          ],
        },
      }),
    ],
  };
});
