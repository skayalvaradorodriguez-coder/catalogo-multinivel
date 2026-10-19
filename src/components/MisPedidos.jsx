import { useCallback, useState } from "react";
import { Link } from "react-router-dom";
import { useApi } from "../hooks/useApi";
import { getMisPedidos } from "../services/pedidosService";
import { dinero, fecha } from "../data/formato";
import MensajeError from "./MensajeError";
import EstadoPedido from "./EstadoPedido";
import Paginacion from "./Paginacion";
import DetallePedido from "./DetallePedido";

const LIMITE = 10;

// Historial de compras del usuario conectado (filtrado por el token en el backend).
const MisPedidos = () => {
  const [page, setPage] = useState(1);
  const [abierto, setAbierto] = useState(null);
  const cargar = useCallback(() => getMisPedidos({ page, limit: LIMITE }), [page]);
  const { data, isLoading, error } = useApi(cargar);

  if (isLoading && !data) return <p className="text-slate-500">Cargando tus pedidos...</p>;
  if (error) return <MensajeError mensaje={error} />;

  return (
    <div className="max-w-4xl">
      <h1 className="text-2xl font-bold text-slate-800 mb-2">Mis Pedidos</h1>
      <p className="text-slate-500 mb-6">Consulta el estado de tus compras.</p>

      {data.total === 0 ? (
        <div className="bg-white p-10 rounded-lg border border-slate-200 text-center">
          <p className="text-slate-600 mb-4">Todavía no has hecho ningún pedido.</p>
          <Link to="/catalogo" className="text-indigo-600 font-semibold hover:underline">
            Ir al catálogo →
          </Link>
        </div>
      ) : (
        <ul className="space-y-3">
          {data.data.map((p) => (
            <li key={p.id} className="bg-white rounded-lg border border-slate-200 shadow-sm">
              <button
                onClick={() => setAbierto(abierto === p.id ? null : p.id)}
                className="w-full min-h-11 p-4 flex flex-wrap items-center justify-between gap-3 text-left"
              >
                <div>
                  <p className="font-mono font-bold text-slate-800">{p.numero}</p>
                  <p className="text-xs text-slate-500">
                    {fecha(p.creadoEn)} · {p.unidades} producto(s)
                  </p>
                </div>
                <div className="flex items-center gap-3">
                  <EstadoPedido estado={p.estado} />
                  <span className="font-bold text-slate-800">{dinero(p.total)}</span>
                  <span className="text-slate-400">{abierto === p.id ? "▲" : "▼"}</span>
                </div>
              </button>
              {abierto === p.id && (
                <div className="px-4 pb-4">
                  <DetallePedido id={p.id} />
                </div>
              )}
            </li>
          ))}
        </ul>
      )}

      <Paginacion page={data.page} limit={data.limit} total={data.total} onCambiar={setPage} />
    </div>
  );
};

export default MisPedidos;
