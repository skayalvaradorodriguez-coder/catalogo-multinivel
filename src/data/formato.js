// Funciones puras de formato para mostrar montos y fechas.
export const dinero = (valor) =>
  `$${Number(valor ?? 0).toLocaleString("es-EC", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;

export const fecha = (iso) =>
  iso ? new Date(iso).toLocaleString("es-EC", { dateStyle: "medium", timeStyle: "short" }) : "—";

export const ESTADOS_PEDIDO = ["pendiente", "preparando", "enviado", "entregado", "cancelado"];
