// Etiqueta de color para el estado de un pedido.
const COLORES = {
  pendiente: "bg-amber-100 text-amber-700",
  preparando: "bg-sky-100 text-sky-700",
  enviado: "bg-indigo-100 text-indigo-700",
  entregado: "bg-green-100 text-green-700",
  cancelado: "bg-red-100 text-red-700",
  pagada: "bg-green-100 text-green-700",
  revertida: "bg-red-100 text-red-700",
};

const EstadoPedido = ({ estado }) => (
  <span
    className={`inline-block text-xs font-semibold px-2 py-1 rounded-full capitalize ${
      COLORES[estado] ?? "bg-slate-100 text-slate-700"
    }`}
  >
    {estado}
  </span>
);

export default EstadoPedido;
