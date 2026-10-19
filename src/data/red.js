// Funciones puras de la red multinivel: comisiones, totales y nivel alcanzado.
// Son estimaciones a partir de las ventas; el monto oficial está en el ledger (/api/comisiones/mias).

export const TASA_COMISION = {
  1: 0.1, // 10% de las ventas de nivel 1
  2: 0.05, // 5% de las ventas de nivel 2
  3: 0.02, // 2% de las ventas de nivel 3
};

// Forma de un nodo (viene de GET /api/red/mia):
// { id, nombre, nivel, ventas, codigo, hijos?: Nodo[] } — nivel relativo al usuario (él es 0).

// Total de referidos de la red (sin contar al usuario raíz)
export const contarRed = (raiz) => {
  const directos = raiz.hijos ?? [];
  return directos.reduce((total, hijo) => total + 1 + contarRed(hijo), 0);
};

// Suma de ventas de toda la red (excluye las ventas propias del usuario raíz)
export const sumarVentasRed = (raiz) => {
  const directos = raiz.hijos ?? [];
  return directos.reduce(
    (total, hijo) => total + hijo.ventas + sumarVentasRed(hijo),
    0
  );
};

// Comisión que recibe el usuario por las ventas de un referido puntual
export const comisionDeReferido = (ref) =>
  ref.ventas * (TASA_COMISION[ref.nivel] ?? 0);

// Comisiones totales del mes acumuladas sobre toda la red
export const sumarComisiones = (raiz) => {
  const directos = raiz.hijos ?? [];
  return directos.reduce(
    (total, hijo) => total + comisionDeReferido(hijo) + sumarComisiones(hijo),
    0
  );
};

// Nivel alcanzado según la cantidad de referidos directos
export const nivelAlcanzado = (referidosDirectos) => {
  if (referidosDirectos >= 6) return "Diamante";
  if (referidosDirectos >= 4) return "Oro";
  if (referidosDirectos >= 2) return "Plata";
  return "Bronce";
};