// Capa de servicios del catálogo. Consulta el backend en Go, que a su vez lee
// los productos desde PostgreSQL. Los componentes no cambian: siguen recibiendo
// una promesa con los mismos campos que tenía el mock.
import { API_URL } from "../config";

export const getProductos = async () => {
  const response = await fetch(`${API_URL}/api/productos`);
  if (!response.ok) throw new Error("No se pudo cargar el catálogo");
  return response.json();
};

export const getProductoById = async (id) => {
  const response = await fetch(`${API_URL}/api/productos/${id}`);
  // 404 = el producto no existe: devolvemos undefined, igual que hacía el mock.
  if (response.status === 404) return undefined;
  if (!response.ok) throw new Error("No se pudo cargar el producto");
  return response.json();
};
