import { useCallback, useState } from "react";
import { Link } from "react-router-dom";
import { useApi } from "../hooks/useApi";
import { getComisionesPendientes, getResumen, getTopProductos, getTopReferidos } from "../services/reportesService";
import { dinero } from "../data/formato";
import MensajeError from "./MensajeError";

// Dashboard del administrador: KPIs reales calculados con SQL en /api/reportes/*.
const Dashboard = () => {
  const [filtro, setFiltro] = useState({ desde: "", hasta: "" });
  const [aplicado, setAplicado] = useState({ desde: "", hasta: "" });

  const cargarResumen = useCallback(() => getResumen(aplicado), [aplicado]);
  const resumen = useApi(cargarResumen);
  const topProductos = useApi(getTopProductos);
  const topReferidos = useApi(getTopReferidos);
  const deudas = useApi(getComisionesPendientes);

  const kpis = resumen.data
    ? [
        { etiqueta: "Total Vendido", valor: dinero(resumen.data.totalVendido), color: "text-indigo-600" },
        { etiqueta: "Pedidos", valor: resumen.data.numPedidos, color: "text-indigo-600" },
        { etiqueta: "Ticket Promedio", valor: dinero(resumen.data.ticketPromedio), color: "text-indigo-600" },
        { etiqueta: "Comisiones Devengadas", valor: dinero(resumen.data.comisionesDevengadas), color: "text-green-600" },
      ]
    : [];

  const inputFecha = "min-h-11 px-3 rounded-lg border border-slate-300 text-sm w-full sm:w-auto";

  return (
    <div>
      <h1 className="text-2xl font-bold text-slate-800 mb-6">Resumen General</h1>

      <form
        onSubmit={(e) => {
          e.preventDefault();
          setAplicado(filtro);
        }}
        className="flex flex-col sm:flex-row sm:items-end gap-3 mb-6"
      >
        <label className="text-sm text-slate-600">
          Desde
          <input
            type="date"
            className={`${inputFecha} block mt-1`}
            value={filtro.desde}
            onChange={(e) => setFiltro((f) => ({ ...f, desde: e.target.value }))}
          />
        </label>
        <label className="text-sm text-slate-600">
          Hasta
          <input
            type="date"
            className={`${inputFecha} block mt-1`}
            value={filtro.hasta}
            onChange={(e) => setFiltro((f) => ({ ...f, hasta: e.target.value }))}
          />
        </label>
        <button className="min-h-11 px-5 rounded-lg bg-indigo-600 text-white font-semibold hover:bg-indigo-700 transition">
          Filtrar
        </button>
        {(aplicado.desde || aplicado.hasta) && (
          <button
            type="button"
            onClick={() => {
              setFiltro({ desde: "", hasta: "" });
              setAplicado({ desde: "", hasta: "" });
            }}
            className="min-h-11 px-5 rounded-lg border border-slate-300 text-slate-700 font-semibold hover:bg-slate-100 transition"
          >
            Ver todo
          </button>
        )}
      </form>

      {resumen.error && <MensajeError mensaje={resumen.error} />}
      {resumen.isLoading && !resumen.data ? (
        <p className="text-slate-500 mb-8">Cargando indicadores...</p>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-4 sm:gap-6 mb-8">
          {kpis.map((kpi) => (
            <div key={kpi.etiqueta} className="bg-white p-5 sm:p-6 rounded-lg shadow-sm border border-slate-100">
              <p className="text-sm text-slate-500 uppercase font-semibold">{kpi.etiqueta}</p>
              <p className={`text-3xl font-bold mt-2 ${kpi.color}`}>{kpi.valor}</p>
            </div>
          ))}
        </div>
      )}
      {resumen.data && (
        <p className="text-sm text-slate-500 -mt-4 mb-8">
          Comisiones por pagar: <b>{dinero(resumen.data.comisionesPendientes)}</b> · pagadas:{" "}
          <b>{dinero(resumen.data.comisionesPagadas)}</b> · pedidos cancelados: <b>{resumen.data.pedidosCancelados}</b>
        </p>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Tabla
          titulo="Top Productos"
          estado={topProductos}
          vacio="Aún no hay ventas registradas."
          columnas={["Producto", "Unidades", "Ingresos"]}
          filas={(topProductos.data ?? []).map((p) => [p.nombre, p.unidades, dinero(p.ingresos)])}
        />
        <Tabla
          titulo="Top Distribuidores"
          estado={topReferidos}
          vacio="Aún no hay ventas de la red."
          columnas={["Distribuidor", "Pedidos", "Ventas", "Comisiones"]}
          filas={(topReferidos.data ?? []).map((r) => [r.nombre, r.pedidos, dinero(r.ventas), dinero(r.comisiones)])}
        />
        <div className="lg:col-span-2">
          <Tabla
            titulo="Comisiones Pendientes por Distribuidor"
            estado={deudas}
            vacio="No hay comisiones pendientes de pago."
            columnas={["Distribuidor", "Código", "Comisiones", "Por pagar"]}
            filas={(deudas.data ?? []).map((d) => [d.nombre, d.codigo, d.comisiones, dinero(d.pendiente)])}
            pie={
              <Link to="/admin/liquidaciones" className="text-indigo-600 font-semibold hover:underline text-sm">
                Ir a liquidaciones →
              </Link>
            }
          />
        </div>
      </div>
    </div>
  );
};

// Tabla de reporte con estados de carga, error y vacío.
const Tabla = ({ titulo, estado, vacio, columnas, filas, pie }) => (
  <div className="bg-white p-4 sm:p-6 rounded-lg shadow-sm border border-slate-100">
    <h2 className="text-lg font-bold text-slate-800 mb-4">{titulo}</h2>
    {estado.isLoading ? (
      <p className="text-sm text-slate-500">Cargando...</p>
    ) : estado.error ? (
      <MensajeError mensaje={estado.error} />
    ) : filas.length === 0 ? (
      <p className="text-sm text-slate-500">{vacio}</p>
    ) : (
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="text-left text-slate-500 border-b border-slate-200">
              {columnas.map((c) => (
                <th key={c} className="py-2 pr-3 font-semibold">
                  {c}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {filas.map((fila, i) => (
              <tr key={i} className="border-b border-slate-100">
                {fila.map((celda, j) => (
                  <td key={j} className={`py-2 pr-3 ${j === 0 ? "font-semibold text-slate-800" : "text-slate-600"}`}>
                    {celda}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    )}
    {pie && <div className="mt-4">{pie}</div>}
  </div>
);

export default Dashboard;
