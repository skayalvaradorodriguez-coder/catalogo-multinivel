import { useApi } from "../hooks/useApi";
import { getMisComisiones } from "../services/comisionesService";
import { dinero, fecha } from "../data/formato";
import MensajeError from "./MensajeError";
import EstadoPedido from "./EstadoPedido";

// Panel "Mis comisiones": el ledger real del usuario (lo que cobró por cada venta de su red).
const MisComisiones = () => {
  const { data, isLoading, error } = useApi(getMisComisiones);

  if (isLoading) return <p className="text-slate-500">Cargando tus comisiones...</p>;
  if (error) return <MensajeError mensaje={error} />;

  const tarjetas = [
    { etiqueta: "Por cobrar", valor: data.totales.pendiente, color: "text-amber-600" },
    { etiqueta: "Cobrado", valor: data.totales.pagada, color: "text-green-600" },
    { etiqueta: "Revertido", valor: data.totales.revertida, color: "text-red-600" },
  ];

  return (
    <div className="bg-white rounded-lg border border-slate-200 shadow-sm p-4 sm:p-6">
      <h2 className="text-lg font-bold text-slate-800 mb-4">Mis Comisiones</h2>

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
        {tarjetas.map((t) => (
          <div key={t.etiqueta} className="rounded-lg border border-slate-100 bg-slate-50 p-4">
            <p className="text-xs uppercase font-semibold text-slate-500">{t.etiqueta}</p>
            <p className={`text-2xl font-bold mt-1 ${t.color}`}>{dinero(t.valor)}</p>
          </div>
        ))}
      </div>

      {data.comisiones.length === 0 ? (
        <p className="text-sm text-slate-500">
          Aún no tienes comisiones. Comparte tu link: cuando alguien de tu red compre, aparecerán aquí.
        </p>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-sm min-w-[480px]">
            <thead>
              <tr className="text-left text-slate-500 border-b border-slate-200">
                <th className="py-2 pr-3">Pedido</th>
                <th className="py-2 pr-3">Nivel</th>
                <th className="py-2 pr-3">Monto</th>
                <th className="py-2 pr-3">Estado</th>
                <th className="py-2">Fecha</th>
              </tr>
            </thead>
            <tbody>
              {data.comisiones.map((c) => (
                <tr key={c.id} className="border-b border-slate-100">
                  <td className="py-2 pr-3 font-mono">{c.pedidoNumero}</td>
                  <td className="py-2 pr-3">
                    {c.nivel} ({c.porcentaje}%)
                  </td>
                  <td className={`py-2 pr-3 font-semibold ${c.monto < 0 ? "text-red-600" : "text-slate-800"}`}>
                    {dinero(c.monto)}
                  </td>
                  <td className="py-2 pr-3">
                    <EstadoPedido estado={c.estado} />
                  </td>
                  <td className="py-2 text-slate-500">{fecha(c.creadoEn)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};

export default MisComisiones;
