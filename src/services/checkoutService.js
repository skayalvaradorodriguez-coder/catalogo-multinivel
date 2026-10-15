// Envía la compra al backend (POST /api/checkout), que la guarda en una transacción ACID.
import { API_URL } from "../config";

export const enviarPedido = async (pedido) => {
  const response = await fetch(`${API_URL}/api/checkout`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(pedido),
  });

  const data = await response.json().catch(() => ({}));
  if (!response.ok) {
    // El backend responde con APIError { status, message, error, details }.
    const detalles =
      data.details && typeof data.details === "object" ? Object.values(data.details).join(". ") : "";
    throw new Error(detalles || data.message || "No se pudo procesar el pedido");
  }
  return data;
};
