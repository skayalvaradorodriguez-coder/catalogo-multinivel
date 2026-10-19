// Reportes del negocio (solo admin).
import { apiFetch } from "./http";

export const getResumen = ({ desde, hasta } = {}) => apiFetch("/api/reportes/resumen", { query: { desde, hasta } });

export const getTopProductos = (limit = 5) => apiFetch("/api/reportes/top-productos", { query: { limit } });

export const getTopReferidos = (limit = 5) => apiFetch("/api/reportes/top-referidos", { query: { limit } });

export const getComisionesPendientes = () => apiFetch("/api/reportes/comisiones-pendientes");
