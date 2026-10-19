import { useCallback, useState } from "react";
import { useApi } from "../../hooks/useApi";
import { cambiarEstadoPedido, getPedidos } from "../../services/pedidosService";
import { dinero, fecha, ESTADOS_PEDIDO } from "../../data/formato";
import MensajeError from "../MensajeError";
import EstadoPedido from "../EstadoPedido";
import Paginacion from "../Paginacion";
import DetallePedido from "../DetallePedido";

const LIMITE = 10;

// /admin/pedidos: todos los pedidos con filtro por estado y cambio de estado.
const AdminPedidos = () => {
  const [page, setPage] = useState(1);
  const [estado, setEstado] = useState("");
  const [abierto, setAbierto] = useState(null);
  const [aviso, setAviso] = useState({ tipo: "", texto: "" });

  const cargar = useCallback(() => getPedidos({ page, limit: LIMITE, estado }), [page, estado]);
  const { data, isLoading, error, recargar } = useApi(cargar);

  const cambiar = async (pedido, nuevo) => {
    if (!nuevo) return;
    const extra = nuevo === "cancelado" ? "\n\nSe devolverá el stock y se revertirán las comisiones." : "";
    if (!window.confirm(`¿Pasar el pedido ${pedido.numero} a "${nuevo}"?${extra}`)) return;
    try {
      await cambiarEstadoPedido(pedido.id, nuevo);
      setAviso({ tipo: "ok", texto: `Pedido ${pedido.numero}: ${pedido.estado} → ${nuevo}` });
      recargar();
    } catch (err) {
      setAviso({ tipo: "error", texto: err.mensajeCompleto });
    }
  };

  return (
    <div>
      <h1 className="text-2xl font-bold text-slate-800 mb-6">Pedidos</h1>

      <div className="flex flex-wrap gap-2 mb-4">
        {["", ...ESTADOS_PEDIDO].map((e) => (
          <button
            key={e || "todos"}
            onClick={() => {
              setEstado(e);
              setPage(1);
            }}
            className={`min-h-11 px-4 rounded-full text-sm font-semibold capitalize border transition ${
              estado === e ? "bg-indigo-600 text-white border-indigo-600" : "border-slate-300 text-slate-700 hover:bg-slate-100"
            }`}
          >
            {e || "Todos"}
          </button>
        ))}
      </div>

      {aviso.texto && (
        <p
          className={`mb-4 text-sm rounded-lg p-3 border ${
            aviso.tipo === "ok" ? "bg-green-50 text-green-700 border-green-200" : "bg-red-50 text-red-600 border-red-200"
          }`}
        >
          {aviso.texto}
        </p>
      )}
      {error && <MensajeError mensaje={error} />}

      {isLoading && !data ? (
        <p className="text-slate-500">Cargando pedidos...</p>
      ) : (
        data && (
          <>
            {data.total === 0 && <p className="text-slate-500">No hay pedidos con ese estado.</p>}
            <ul className="space-y-3">
              {data.data.map((p) => (
                <li key={p.id} className="bg-white rounded-lg border border-slate-200 shadow-sm p-4">
                  <div className="flex flex-wrap items-center justify-between gap-3">
                    <button onClick={() => setAbierto(abierto === p.id ? null : p.id)} className="text-left min-h-11">
                      <p className="font-mono font-bold text-slate-800">
                        {p.numero} <span className="text-slate-400">{abierto === p.id ? "▲" : "▼"}</span>
                      </p>
                      <p className="text-xs text-slate-500">
                        {p.clienteNombre} · {fecha(p.creadoEn)} · {p.unidades} u.
                      </p>
                    </button>
                    <div className="flex flex-wrap items-center gap-3">
                      <EstadoPedido estado={p.estado} />
                      <span className="font-bold text-slate-800">{dinero(p.total)}</span>
                      {p.siguientes.length > 0 && (
                        <select
                          value=""
                          onChange={(e) => cambiar(p, e.target.value)}
                          className="min-h-11 px-3 rounded-lg border border-slate-300 text-sm bg-white"
                          aria-label="Cambiar estado"
                        >
                          <option value="">Cambiar a…</option>
                          {p.siguientes.map((s) => (
                            <option key={s} value={s}>
                              {s}
                            </option>
                          ))}
                        </select>
                      )}
                    </div>
                  </div>
                  {abierto === p.id && (
                    <div className="mt-3">
                      <DetallePedido id={p.id} />
                    </div>
                  )}
                </li>
              ))}
            </ul>
            <Paginacion page={data.page} limit={data.limit} total={data.total} onCambiar={setPage} />
          </>
        )
      )}
    </div>
  );
};

export default AdminPedidos;
