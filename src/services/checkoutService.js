// Envía la compra al backend (POST /api/checkout), que la guarda en una transacción ACID.
// El comprador se identifica con el token de la sesión.
import { apiFetch } from "./http";

export const enviarPedido = (pedido) => apiFetch("/api/checkout", { method: "POST", body: pedido });
