// Capa de servicios del catálogo. Hoy devuelve el mock simulando una llamada
// asíncrona; en la Unidad 2 se cambia el cuerpo por un fetch real sin tocar
// los componentes que la consumen.
import { productosMock } from "../data/productos";

export const getProductos = () => {
  return new Promise((resolve) => {
    setTimeout(() => resolve(productosMock), 400);
  });
};

export const getProductoById = (id) => {
  return new Promise((resolve) => {
    setTimeout(() => resolve(productosMock.find((p) => p.id === id)), 300);
  });
};