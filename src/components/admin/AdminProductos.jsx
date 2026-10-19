import { useCallback, useState } from "react";
import { useApi } from "../../hooks/useApi";
import {
  actualizarProducto,
  crearProducto,
  desactivarProducto,
  getProductosAdmin,
} from "../../services/productosService";
import { dinero } from "../../data/formato";
import MensajeError from "../MensajeError";
import Paginacion from "../Paginacion";
import FormularioProducto from "./FormularioProducto";

const LIMITE = 10;

// /admin/productos: catálogo administrable (incluye desactivados) con búsqueda server-side.
const AdminProductos = () => {
  const [page, setPage] = useState(1);
  const [busqueda, setBusqueda] = useState("");
  const [q, setQ] = useState("");
  const [editando, setEditando] = useState(null); // null = cerrado, {} = nuevo, producto = editar
  const [aviso, setAviso] = useState("");

  const cargar = useCallback(() => getProductosAdmin({ page, limit: LIMITE, q }), [page, q]);
  const { data, isLoading, error, recargar } = useApi(cargar);

  const guardar = async (datos) => {
    if (editando?.id) await actualizarProducto(editando.id, datos);
    else await crearProducto(datos);
    setEditando(null);
    setAviso(editando?.id ? "Producto actualizado" : "Producto creado");
    recargar();
  };

  const alternarActivo = async (p) => {
    const accion = p.activo ? "desactivar" : "activar";
    if (!window.confirm(`¿Seguro que quieres ${accion} "${p.nombre}"?`)) return;
    try {
      if (p.activo) await desactivarProducto(p.id);
      else await actualizarProducto(p.id, { activo: true });
      setAviso(`Producto ${p.activo ? "desactivado" : "activado"}`);
      recargar();
    } catch (err) {
      setAviso(err.mensajeCompleto);
    }
  };

  return (
    <div>
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-6">
        <h1 className="text-2xl font-bold text-slate-800">Productos</h1>
        <button
          onClick={() => setEditando({})}
          className="min-h-11 px-5 rounded-lg bg-indigo-600 text-white font-semibold hover:bg-indigo-700 transition"
        >
          + Nuevo producto
        </button>
      </div>

      <form
        onSubmit={(e) => {
          e.preventDefault();
          setPage(1);
          setQ(busqueda.trim());
        }}
        className="flex gap-2 mb-4"
      >
        <input
          value={busqueda}
          onChange={(e) => setBusqueda(e.target.value)}
          placeholder="Buscar por nombre, descripción o categoría"
          className="flex-1 min-w-0 min-h-11 px-3 rounded-lg border border-slate-300"
        />
        <button className="min-h-11 px-4 rounded-lg border border-slate-300 font-semibold hover:bg-slate-100">Buscar</button>
      </form>

      {aviso && (
        <p className="mb-4 text-sm bg-green-50 text-green-700 border border-green-200 rounded-lg p-3">{aviso}</p>
      )}
      {error && <MensajeError mensaje={error} />}

      {isLoading && !data ? (
        <p className="text-slate-500">Cargando productos...</p>
      ) : (
        data && (
          <>
            <div className="bg-white rounded-lg border border-slate-200 shadow-sm overflow-x-auto">
              <table className="w-full text-sm min-w-[640px]">
                <thead>
                  <tr className="text-left text-slate-500 border-b border-slate-200">
                    <th className="p-3">Producto</th>
                    <th className="p-3">Categoría</th>
                    <th className="p-3">Precio</th>
                    <th className="p-3">Stock</th>
                    <th className="p-3">Estado</th>
                    <th className="p-3 text-right">Acciones</th>
                  </tr>
                </thead>
                <tbody>
                  {data.data.map((p) => (
                    <tr key={p.id} className={`border-b border-slate-100 ${p.activo ? "" : "opacity-60"}`}>
                      <td className="p-3">
                        <div className="flex items-center gap-3">
                          {p.img ? (
                            <img src={p.img} alt="" className="w-10 h-10 rounded object-cover bg-slate-100" />
                          ) : (
                            <div className="w-10 h-10 rounded bg-slate-100" />
                          )}
                          <span className="font-semibold text-slate-800">{p.nombre}</span>
                        </div>
                      </td>
                      <td className="p-3 text-slate-600">{p.categoria}</td>
                      <td className="p-3 text-slate-800">{dinero(p.precio)}</td>
                      <td className={`p-3 ${p.stock < 5 ? "text-red-600 font-semibold" : "text-slate-600"}`}>{p.stock}</td>
                      <td className="p-3">
                        <span
                          className={`text-xs font-semibold px-2 py-1 rounded-full ${
                            p.activo ? "bg-green-100 text-green-700" : "bg-slate-200 text-slate-600"
                          }`}
                        >
                          {p.activo ? "Activo" : "Inactivo"}
                        </span>
                      </td>
                      <td className="p-3">
                        <div className="flex justify-end gap-2">
                          <button
                            onClick={() => setEditando(p)}
                            className="min-h-11 px-3 rounded-lg border border-slate-300 font-semibold hover:bg-slate-100"
                          >
                            Editar
                          </button>
                          <button
                            onClick={() => alternarActivo(p)}
                            className={`min-h-11 px-3 rounded-lg font-semibold ${
                              p.activo ? "text-red-600 hover:bg-red-50" : "text-green-700 hover:bg-green-50"
                            }`}
                          >
                            {p.activo ? "Desactivar" : "Activar"}
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              {data.total === 0 && <p className="p-6 text-center text-slate-500">No se encontraron productos.</p>}
            </div>
            <Paginacion page={data.page} limit={data.limit} total={data.total} onCambiar={setPage} />
          </>
        )
      )}

      {editando && (
        <FormularioProducto producto={editando.id ? editando : null} onGuardar={guardar} onCancelar={() => setEditando(null)} />
      )}
    </div>
  );
};

export default AdminProductos;
