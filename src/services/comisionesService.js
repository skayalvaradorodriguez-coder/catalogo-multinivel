// Ledger de comisiones y liquidaciones (pagos a distribuidores).
import { apiFetch } from "./http";

export const getMisComisiones = () => apiFetch("/api/comisiones/mias");

export const getMisLiquidaciones = ({ page, limit }) => apiFetch("/api/liquidaciones/mias", { query: { page, limit } });

// ---- Administración ----
export const getLiquidaciones = ({ page, limit }) => apiFetch("/api/liquidaciones", { query: { page, limit } });

export const liquidar = (referidoId) =>
  apiFetch("/api/liquidaciones", { method: "POST", body: { referido_id: referidoId } });
