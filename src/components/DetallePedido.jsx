import { useCallback } from "react";
import { useApi } from "../hooks/useApi";
import { getPedido } from "../services/pedidosService";
import { dinero } from "../data/formato";
import MensajeError from "./MensajeError";

// Detalle expandible de un pedido: datos de envío + ítems con el precio guardado.
const DetallePedido = ({ id }) => {
  const cargar = useCallback(() => getPedido(id), [id]);
  const { data: pedido, isLoading, error } = useApi(cargar);

  if (isLoading) return <p className="text-sm text-slate-500 p-4">Cargando detalle...</p>;
  if (error) return <MensajeError mensaje={error} />;

  return (
    <div className="bg-slate-50 rounded-lg p-4 text-sm space-y-3">
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 text-slate-600">
        <p>
          <span className="font-semibold text-slate-700">Envío:</span> {pedido.cliente.direccion}, {pedido.cliente.ciudad}
        </p>
        <p>
          <span className="font-semibold text-slate-700">Teléfono:</span> {pedido.cliente.telefono}
        </p>
        <p>
          <span className="font-semibold text-slate-700">Pago:</span> {pedido.metodoPago}
        </p>
        <p className="break-all">
          <span className="font-semibold text-slate-700">Correo:</span> {pedido.cliente.email}
        </p>
      </div>
      <ul className="divide-y divide-slate-200 border-t border-slate-200">
        {pedido.items.map((item) => (
          <li key={item.productoId} className="flex justify-between gap-3 py-2">
            <span className="text-slate-700">
              {item.nombre} × {item.cantidad}
            </span>
            <span className="font-semibold text-slate-800">{dinero(item.precio * item.cantidad)}</span>
          </li>
        ))}
      </ul>
    </div>
  );
};

export default DetallePedido;
