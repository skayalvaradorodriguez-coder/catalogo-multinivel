// Cliente HTTP único de la aplicación. Todos los servicios llaman a la API a través de apiFetch:
// - agrega "Authorization: Bearer <token>" automáticamente si hay sesión,
// - convierte las respuestas de error (APIError del backend) en un ApiError con status,
// - si el token venció (401), cierra la sesión.
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
};

// Evento que escucha AuthContext para cerrar la sesión cuando el backend rechaza el token.
export const EVENTO_SESION_EXPIRADA = "multicatalogo:sesion-expirada";

// ---- Errores ---------------------------------------------------------------------------
export class ApiError extends Error {
  constructor(status, message, details = null) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.details = details;
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
export const apiFetch = async (ruta, { method = "GET", body, query } = {}) => {
  const url = new URL(`${API_URL}${ruta}`);
  Object.entries(query ?? {}).forEach(([clave, valor]) => {
    if (valor !== undefined && valor !== null && valor !== "") url.searchParams.set(clave, valor);
  });

  const headers = { Accept: "application/json" };
  if (body !== undefined) headers["Content-Type"] = "application/json";
  const token = leerSesion()?.token;
  if (token) headers.Authorization = `Bearer ${token}`;

  let response;
  try {
    response = await fetch(url, {
      method,
      headers,
      body: body !== undefined ? JSON.stringify(body) : undefined,
    });
  } catch {
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
