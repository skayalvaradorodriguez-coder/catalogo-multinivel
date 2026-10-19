import { useCallback, useState } from "react";
import { useApi } from "../../hooks/useApi";
import { getComisionesPendientes } from "../../services/reportesService";
import { getLiquidaciones, liquidar } from "../../services/comisionesService";
import { dinero, fecha } from "../../data/formato";
import MensajeError from "../MensajeError";
import Paginacion from "../Paginacion";

const LIMITE = 10;

// /admin/liquidaciones: deuda por distribuidor + pago (liquidación) + historial.
const AdminLiquidaciones = () => {
  const [page, setPage] = useState(1);
  const [aviso, setAviso] = useState({ tipo: "", texto: "" });
  const [pagando, setPagando] = useState(null);

  const deudas = useApi(getComisionesPendientes);
  const cargarHistorial = useCallback(() => getLiquidaciones({ page, limit: LIMITE }), [page]);
  const historial = useApi(cargarHistorial);

  const pagar = async (d) => {
    if (!window.confirm(`¿Liquidar ${dinero(d.pendiente)} a ${d.nombre}? Sus ${d.comisiones} comisión(es) pasarán a "pagada".`))
      return;
    setPagando(d.referidoId);
    try {
      const liq = await liquidar(d.referidoId);
      setAviso({ tipo: "ok", texto: `Liquidación #${liq.id}: ${dinero(liq.total)} pagados a ${liq.referidoNombre}` });
      deudas.recargar();
      historial.recargar();
    } catch (err) {
      setAviso({ tipo: "error", texto: err.mensajeCompleto });
    } finally {
      setPagando(null);
    }
  };

  return (
    <div>
      <h1 className="text-2xl font-bold text-slate-800 mb-6">Liquidaciones</h1>

      {aviso.texto && (
        <p
          className={`mb-4 text-sm rounded-lg p-3 border ${
            aviso.tipo === "ok" ? "bg-green-50 text-green-700 border-green-200" : "bg-red-50 text-red-600 border-red-200"
          }`}
        >
          {aviso.texto}
        </p>
      )}

      <section className="bg-white rounded-lg border border-slate-200 shadow-sm p-4 sm:p-6 mb-8">
        <h2 className="text-lg font-bold text-slate-800 mb-4">Comisiones por pagar</h2>
        {deudas.isLoading ? (
          <p className="text-sm text-slate-500">Cargando...</p>
        ) : deudas.error ? (
          <MensajeError mensaje={deudas.error} />
        ) : deudas.data.length === 0 ? (
          <p className="text-sm text-slate-500">No hay comisiones pendientes. ¡Todo está pagado!</p>
        ) : (
          <ul className="divide-y divide-slate-100">
            {deudas.data.map((d) => (
              <li key={d.referidoId} className="py-3 flex flex-wrap items-center justify-between gap-3">
                <div>
                  <p className="font-semibold text-slate-800">
                    {d.nombre} <span className="font-mono text-xs text-slate-500">{d.codigo}</span>
                  </p>
                  <p className="text-xs text-slate-500">
                    {d.email ?? "sin cuenta"} · {d.comisiones} comisión(es)
                  </p>
                </div>
                <div className="flex items-center gap-3">
                  <span className="font-bold text-amber-600">{dinero(d.pendiente)}</span>
                  <button
                    onClick={() => pagar(d)}
                    disabled={pagando === d.referidoId}
                    className="min-h-11 px-4 rounded-lg bg-green-600 text-white font-semibold hover:bg-green-700 disabled:opacity-60"
                  >
                    {pagando === d.referidoId ? "Pagando..." : "Liquidar"}
                  </button>
                </div>
              </li>
            ))}
          </ul>
        )}
      </section>

      <section className="bg-white rounded-lg border border-slate-200 shadow-sm p-4 sm:p-6">
        <h2 className="text-lg font-bold text-slate-800 mb-4">Historial de pagos</h2>
        {historial.isLoading && !historial.data ? (
          <p className="text-sm text-slate-500">Cargando...</p>
        ) : historial.error ? (
          <MensajeError mensaje={historial.error} />
        ) : historial.data.total === 0 ? (
          <p className="text-sm text-slate-500">Todavía no se ha hecho ninguna liquidación.</p>
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="w-full text-sm min-w-[480px]">
                <thead>
                  <tr className="text-left text-slate-500 border-b border-slate-200">
                    <th className="py-2 pr-3">#</th>
                    <th className="py-2 pr-3">Distribuidor</th>
                    <th className="py-2 pr-3">Comisiones</th>
                    <th className="py-2 pr-3">Total</th>
                    <th className="py-2">Pagada</th>
                  </tr>
                </thead>
                <tbody>
                  {historial.data.data.map((l) => (
                    <tr key={l.id} className="border-b border-slate-100">
                      <td className="py-2 pr-3 font-mono">{l.id}</td>
                      <td className="py-2 pr-3 font-semibold text-slate-800">{l.referidoNombre}</td>
                      <td className="py-2 pr-3">{l.comisiones}</td>
                      <td className="py-2 pr-3 font-semibold text-green-700">{dinero(l.total)}</td>
                      <td className="py-2 text-slate-500">{fecha(l.pagadaEn)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            <Paginacion
              page={historial.data.page}
              limit={historial.data.limit}
              total={historial.data.total}
              onCambiar={setPage}
            />
          </>
        )}
      </section>
    </div>
  );
};

export default AdminLiquidaciones;
