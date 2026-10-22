// Cliente HTTP único de la aplicación. Todos los servicios llaman a la API a través de apiFetch:
// - agrega "Authorization: Bearer <token>" si hay sesión (salvo en peticiones públicas),
// - convierte las respuestas de error (APIError del backend) en un ApiError con status,
// - si el token venció (401), cierra la sesión,
// - sin conexión responde con un mensaje claro en vez de un error genérico.
import { API_URL } from "../config";

const CLAVE_SESION = "multicatalogo_sesion";

// ---- Sesión (usuario + JWT) en localStorage --------------------------------------------
export const leerSesion = () => {
  try {
    const guardada = localStorage.getItem(CLAVE_SESION);
    return guardada ? JSON.parse(guardada) : null;
  } catch {
    return null;
  }
};

export const guardarSesion = (sesion) => {
  try {
    localStorage.setItem(CLAVE_SESION, JSON.stringify(sesion));
  } catch {
    // Sin localStorage (modo privado estricto) la sesión dura solo mientras la pestaña esté abierta.
  }
};

export const borrarSesion = () => {
  try {
    localStorage.removeItem(CLAVE_SESION);
  } catch {
    // nada que borrar
  }
  borrarCacheDatos();
};

// Borra las respuestas de la API guardadas por el service worker (cachés "multicatalogo-api-*").
// Se llama al cerrar sesión y cuando el token deja de ser válido: el siguiente usuario del
// equipo no debe heredar nada de la sesión anterior.
export const borrarCacheDatos = async () => {
  if (typeof caches === "undefined") return;
  try {
    const nombres = await caches.keys();
    await Promise.all(nombres.filter((n) => n.startsWith("multicatalogo-api-")).map((n) => caches.delete(n)));
  } catch {
    // Cache Storage no disponible (por ejemplo, en un contexto no seguro)
  }
};

// Evento que escucha AuthContext para cerrar la sesión cuando el backend rechaza el token.
export const EVENTO_SESION_EXPIRADA = "multicatalogo:sesion-expirada";

export const MENSAJE_SIN_CONEXION = "Sin conexión. Necesitas internet para esta acción; inténtalo cuando vuelvas a estar en línea.";
export const MENSAJE_SECCION_SIN_CONEXION = "Necesitas conexión para ver esta sección.";

const estaSinConexion = () => typeof navigator !== "undefined" && navigator.onLine === false;

// ---- Errores ---------------------------------------------------------------------------
export class ApiError extends Error {
  constructor(status, message, details = null) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.details = details;
  }

  // true si el error es por falta de conexión (status 0).
  get sinConexion() {
    return this.status === 0;
  }

  // Mensaje para mostrar: incluye los errores por campo si el backend los envió.
  get mensajeCompleto() {
    if (this.details && typeof this.details === "object") {
      return Object.values(this.details).join(". ");
    }
    return this.message;
  }
}

// ---- Petición ---------------------------------------------------------------------------
// publico: true → no envía el token. Lo usan las lecturas del catálogo para que el service
// worker pueda guardarlas en caché (una respuesta con Authorization nunca se guarda).
export const apiFetch = async (ruta, { method = "GET", body, query, publico = false } = {}) => {
  const url = new URL(`${API_URL}${ruta}`, window.location.origin);
  Object.entries(query ?? {}).forEach(([clave, valor]) => {
    if (valor !== undefined && valor !== null && valor !== "") url.searchParams.set(clave, valor);
  });

  // Sin conexión solo tiene sentido intentar las lecturas públicas: el service worker puede
  // responderlas desde su caché. Todo lo demás (privado o escritura) se corta aquí mismo.
  const lecturaPublica = publico && method === "GET";
  if (estaSinConexion() && !lecturaPublica) {
    throw new ApiError(0, method === "GET" ? MENSAJE_SECCION_SIN_CONEXION : MENSAJE_SIN_CONEXION);
  }

  const headers = { Accept: "application/json" };
  if (body !== undefined) headers["Content-Type"] = "application/json";
  const token = publico ? null : leerSesion()?.token;
  if (token) headers.Authorization = `Bearer ${token}`;

  let response;
  try {
    response = await fetch(url, {
      method,
      headers,
      body: body !== undefined ? JSON.stringify(body) : undefined,
    });
  } catch {
    if (estaSinConexion()) {
      throw new ApiError(0, lecturaPublica ? "Sin conexión y sin datos guardados todavía." : MENSAJE_SIN_CONEXION);
    }
    throw new ApiError(0, "No se pudo conectar con el servidor. Verifica que el backend esté corriendo.");
  }

  const data = await response.json().catch(() => null);
  if (!response.ok) {
    if (response.status === 401 && token) {
      borrarSesion();
      window.dispatchEvent(new Event(EVENTO_SESION_EXPIRADA));
    }
    throw new ApiError(response.status, data?.message ?? `Error ${response.status}`, data?.details ?? null);
  }
  return data;
};
