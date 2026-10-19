// Controles de paginación para las respuestas { data, page, limit, total } del backend.
const Paginacion = ({ page, limit, total, onCambiar }) => {
  const paginas = Math.max(1, Math.ceil(total / limit));
  if (total <= limit) return null;

  const boton =
    "min-h-11 min-w-11 px-4 rounded-lg border border-slate-300 text-sm font-semibold text-slate-700 " +
    "hover:bg-slate-100 disabled:opacity-40 disabled:cursor-not-allowed transition";

  return (
    <div className="flex items-center justify-between gap-3 mt-4">
      <button className={boton} disabled={page <= 1} onClick={() => onCambiar(page - 1)}>
        ← Anterior
      </button>
      <span className="text-sm text-slate-500">
        Página {page} de {paginas} · {total} registros
      </span>
      <button className={boton} disabled={page >= paginas} onClick={() => onCambiar(page + 1)}>
        Siguiente →
      </button>
    </div>
  );
};

export default Paginacion;
