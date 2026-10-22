// Registro de clientes y administración de cuentas.
import { apiFetch } from "./http";

export const iniciarSesion = (email, password) =>
  apiFetch("/api/login", { method: "POST", body: { email, password }, publico: true });

export const registrarse = ({ nombre, email, password, ref }) =>
  apiFetch("/api/register", { method: "POST", body: { nombre, email, password, ref }, publico: true });

// ---- Administración ----
export const getUsuarios = ({ page, limit, q }) => apiFetch("/api/usuarios", { query: { page, limit, q } });

export const actualizarUsuario = (id, cambios) => apiFetch(`/api/usuarios/${id}`, { method: "PATCH", body: cambios });
