// Consulta y gestión de pedidos.
import { apiFetch } from "./http";

export const getMisPedidos = ({ page, limit }) => apiFetch("/api/pedidos/mios", { query: { page, limit } });

export const getPedido = (id) => apiFetch(`/api/pedidos/${id}`);

// ---- Administración ----
export const getPedidos = ({ page, limit, estado }) => apiFetch("/api/pedidos", { query: { page, limit, estado } });

export const cambiarEstadoPedido = (id, estado) =>
  apiFetch(`/api/pedidos/${id}/estado`, { method: "PATCH", body: { estado } });
