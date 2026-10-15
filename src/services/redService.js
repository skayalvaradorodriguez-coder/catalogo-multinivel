// Servicio de la red multinivel: pide el árbol de referidos al backend.
// El backend lo arma desde PostgreSQL y lo guarda en Redis por usuario (red:tree:{email}).
import { API_URL } from "../config";

export const getRed = async (email) => {
  const response = await fetch(`${API_URL}/api/red?email=${encodeURIComponent(email ?? "")}`);
  if (!response.ok) throw new Error("No se pudo cargar la red");
  return response.json();
};
