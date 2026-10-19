import { useState } from "react";
import { categorias } from "../../data/productos";

const VACIO = { nombre: "", descripcion: "", precio: "", categoria: categorias[0], img: "", galeria: "", stock: "0" };

// Formulario de alta/edición de producto (modal). Mobile-first: un campo por fila.
const FormularioProducto = ({ producto, onGuardar, onCancelar }) => {
  const [datos, setDatos] = useState(() =>
    producto
      ? {
          ...producto,
          precio: String(producto.precio),
          stock: String(producto.stock),
          galeria: (producto.galeria ?? []).join("\n"),
        }
      : VACIO
  );
  const [error, setError] = useState("");
  const [guardando, setGuardando] = useState(false);

  const cambiar = (campo) => (e) => setDatos((prev) => ({ ...prev, [campo]: e.target.value }));

  const handleSubmit = async (e) => {
    e.preventDefault();
    const precio = Number(datos.precio);
    const stock = Number(datos.stock);
    if (!datos.nombre.trim() || !datos.categoria.trim()) return setError("Nombre y categoría son obligatorios");
    if (!(precio > 0)) return setError("El precio debe ser mayor que cero");
    if (!Number.isInteger(stock) || stock < 0) return setError("El stock debe ser un número entero ≥ 0");

    setError("");
    setGuardando(true);
    try {
      await onGuardar({
        nombre: datos.nombre.trim(),
        descripcion: datos.descripcion.trim(),
        precio,
        categoria: datos.categoria.trim(),
        img: datos.img.trim(),
        galeria: datos.galeria
          .split("\n")
          .map((url) => url.trim())
          .filter(Boolean),
        stock,
      });
    } catch (err) {
      setError(err.mensajeCompleto ?? err.message);
      setGuardando(false);
    }
  };

  const input = "w-full px-3 py-2.5 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 outline-none";

  return (
    <div className="fixed inset-0 z-50 bg-black/50 flex items-end sm:items-center justify-center p-0 sm:p-4">
      <form
        onSubmit={handleSubmit}
        className="bg-white w-full sm:max-w-lg max-h-[92vh] overflow-y-auto rounded-t-2xl sm:rounded-2xl p-5 sm:p-6 space-y-4"
      >
        <h2 className="text-xl font-bold text-slate-800">{producto ? "Editar producto" : "Nuevo producto"}</h2>

        {error && <p className="bg-red-50 text-red-600 text-sm p-3 rounded-lg border border-red-200">{error}</p>}

        <label className="block text-sm font-medium text-slate-700">
          Nombre *
          <input className={`${input} mt-1`} value={datos.nombre} onChange={cambiar("nombre")} required />
        </label>
        <label className="block text-sm font-medium text-slate-700">
          Descripción
          <textarea className={`${input} mt-1`} rows={3} value={datos.descripcion} onChange={cambiar("descripcion")} />
        </label>
        <label className="block text-sm font-medium text-slate-700">
          Precio (USD) *
          <input
            type="number"
            step="0.01"
            min="0.01"
            inputMode="decimal"
            className={`${input} mt-1`}
            value={datos.precio}
            onChange={cambiar("precio")}
            required
          />
        </label>
        <label className="block text-sm font-medium text-slate-700">
          Categoría *
          <input className={`${input} mt-1`} list="categorias" value={datos.categoria} onChange={cambiar("categoria")} required />
          <datalist id="categorias">
            {categorias.map((c) => (
              <option key={c} value={c} />
            ))}
          </datalist>
        </label>
        <label className="block text-sm font-medium text-slate-700">
          Stock *
          <input
            type="number"
            min="0"
            step="1"
            inputMode="numeric"
            className={`${input} mt-1`}
            value={datos.stock}
            onChange={cambiar("stock")}
            required
          />
        </label>
        <label className="block text-sm font-medium text-slate-700">
          URL de la imagen principal
          <input type="url" className={`${input} mt-1`} value={datos.img} onChange={cambiar("img")} placeholder="https://..." />
        </label>
        <label className="block text-sm font-medium text-slate-700">
          Galería (una URL por línea)
          <textarea className={`${input} mt-1 font-mono text-xs`} rows={3} value={datos.galeria} onChange={cambiar("galeria")} />
        </label>

        <div className="flex flex-col-reverse sm:flex-row gap-3 pt-2">
          <button
            type="button"
            onClick={onCancelar}
            className="min-h-11 flex-1 rounded-lg border border-slate-300 font-semibold text-slate-700 hover:bg-slate-100"
          >
            Cancelar
          </button>
          <button
            disabled={guardando}
            className="min-h-11 flex-1 rounded-lg bg-indigo-600 text-white font-semibold hover:bg-indigo-700 disabled:opacity-60"
          >
            {guardando ? "Guardando..." : "Guardar"}
          </button>
        </div>
      </form>
    </div>
  );
};

export default FormularioProducto;
