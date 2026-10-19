import { useCallback, useState } from "react";
import { useApi } from "../../hooks/useApi";
import { useAuth } from "../../context/useAuth";
import { actualizarUsuario, getUsuarios } from "../../services/usuariosService";
import { fecha } from "../../data/formato";
import MensajeError from "../MensajeError";
import Paginacion from "../Paginacion";

const LIMITE = 10;

// /admin/usuarios: cuentas registradas, cambio de rol y activación/desactivación.
const AdminUsuarios = () => {
  const { user } = useAuth();
  const [page, setPage] = useState(1);
  const [busqueda, setBusqueda] = useState("");
  const [q, setQ] = useState("");
  const [aviso, setAviso] = useState("");

  const cargar = useCallback(() => getUsuarios({ page, limit: LIMITE, q }), [page, q]);
  const { data, isLoading, error, recargar } = useApi(cargar);

  const actualizar = async (u, cambios, pregunta) => {
    if (!window.confirm(pregunta)) return;
    try {
      await actualizarUsuario(u.id, cambios);
      setAviso(`Cuenta de ${u.email} actualizada`);
      recargar();
    } catch (err) {
      setAviso(err.mensajeCompleto);
    }
  };

  return (
    <div>
      <h1 className="text-2xl font-bold text-slate-800 mb-6">Usuarios</h1>

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
          placeholder="Buscar por nombre o email"
          className="flex-1 min-w-0 min-h-11 px-3 rounded-lg border border-slate-300"
        />
        <button className="min-h-11 px-4 rounded-lg border border-slate-300 font-semibold hover:bg-slate-100">Buscar</button>
      </form>

      {aviso && <p className="mb-4 text-sm bg-slate-50 border border-slate-200 rounded-lg p-3">{aviso}</p>}
      {error && <MensajeError mensaje={error} />}

      {isLoading && !data ? (
        <p className="text-slate-500">Cargando usuarios...</p>
      ) : (
        data && (
          <>
            <div className="bg-white rounded-lg border border-slate-200 shadow-sm overflow-x-auto">
              <table className="w-full text-sm min-w-[680px]">
                <thead>
                  <tr className="text-left text-slate-500 border-b border-slate-200">
                    <th className="p-3">Usuario</th>
                    <th className="p-3">Código</th>
                    <th className="p-3">Rol</th>
                    <th className="p-3">Alta</th>
                    <th className="p-3 text-right">Acciones</th>
                  </tr>
                </thead>
                <tbody>
                  {data.data.map((u) => {
                    const esYo = u.email === user?.email;
                    return (
                      <tr key={u.id} className={`border-b border-slate-100 ${u.activo ? "" : "opacity-60"}`}>
                        <td className="p-3">
                          <p className="font-semibold text-slate-800">{u.nombre || "—"}</p>
                          <p className="text-xs text-slate-500">{u.email}</p>
                        </td>
                        <td className="p-3 font-mono">{u.codigo || "—"}</td>
                        <td className="p-3">
                          <span
                            className={`text-xs font-semibold px-2 py-1 rounded-full uppercase ${
                              u.rol === "admin" ? "bg-amber-100 text-amber-700" : "bg-indigo-100 text-indigo-700"
                            }`}
                          >
                            {u.rol}
                          </span>
                          {!u.activo && <span className="ml-2 text-xs text-red-600 font-semibold">desactivada</span>}
                        </td>
                        <td className="p-3 text-slate-500">{fecha(u.creadoEn)}</td>
                        <td className="p-3">
                          {esYo ? (
                            <p className="text-right text-xs text-slate-400">Tu cuenta</p>
                          ) : (
                            <div className="flex justify-end gap-2">
                              <button
                                onClick={() =>
                                  actualizar(
                                    u,
                                    { rol: u.rol === "admin" ? "cliente" : "admin" },
                                    `¿Cambiar el rol de ${u.email} a ${u.rol === "admin" ? "cliente" : "admin"}?`
                                  )
                                }
                                className="min-h-11 px-3 rounded-lg border border-slate-300 font-semibold hover:bg-slate-100"
                              >
                                {u.rol === "admin" ? "Hacer cliente" : "Hacer admin"}
                              </button>
                              <button
                                onClick={() =>
                                  actualizar(
                                    u,
                                    { activo: !u.activo },
                                    `¿${u.activo ? "Desactivar" : "Activar"} la cuenta de ${u.email}?`
                                  )
                                }
                                className={`min-h-11 px-3 rounded-lg font-semibold ${
                                  u.activo ? "text-red-600 hover:bg-red-50" : "text-green-700 hover:bg-green-50"
                                }`}
                              >
                                {u.activo ? "Desactivar" : "Activar"}
                              </button>
                            </div>
                          )}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
              {data.total === 0 && <p className="p-6 text-center text-slate-500">No se encontraron usuarios.</p>}
            </div>
            <Paginacion page={data.page} limit={data.limit} total={data.total} onCambiar={setPage} />
          </>
        )
      )}
    </div>
  );
};

export default AdminUsuarios;
