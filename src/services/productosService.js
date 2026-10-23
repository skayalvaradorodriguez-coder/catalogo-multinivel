// Catálogo público y administración de productos.
import { apiFetch } from "./http";

// Lecturas públicas (sin token): el service worker las guarda para verlas sin conexión.
export const getProductos = (q) => apiFetch("/api/productos", { query: { q }, publico: true });

// Devuelve undefined si el producto no existe o está desactivado (404).
export const getProductoById = async (id) => {
  try {
    return await apiFetch(`/api/productos/${id}`, { publico: true });
  } catch (err) {
    if (err.status === 404) return undefined;
    throw err;
  }
};

// ---- Administración (requiere rol admin) ----
export const getProductosAdmin = ({ page, limit, q }) =>
  apiFetch("/api/admin/productos", { query: { page, limit, q } });

export const crearProducto = (producto) => apiFetch("/api/productos", { method: "POST", body: producto });

export const actualizarProducto = (id, cambios) =>
  apiFetch(`/api/productos/${id}`, { method: "PUT", body: cambios });

export const desactivarProducto = (id) => apiFetch(`/api/productos/${id}`, { method: "DELETE" });
