import { useState } from "react";
import { Link, Navigate, useNavigate, useSearchParams } from "react-router-dom";
import { useAuth } from "../context/useAuth";
import { useOnline } from "../hooks/useOnline";
import BannerSinConexion from "./BannerSinConexion";
import { registrarse } from "../services/usuariosService";

// Vista pública /registro. Si llega con ?ref=CODIGO (link de invitación), el código se rellena solo
// y la nueva cuenta queda colgando de quien la invitó en la red multinivel.
const Registro = () => {
  const [params] = useSearchParams();
  const { login, isAuthenticated } = useAuth();
  const online = useOnline();
  const navigate = useNavigate();

  const [datos, setDatos] = useState({
    nombre: "",
    email: "",
    password: "",
    ref: (params.get("ref") ?? "").toUpperCase(),
  });
  const [error, setError] = useState("");
  const [enviando, setEnviando] = useState(false);

  if (isAuthenticated) return <Navigate to="/tienda" replace />;

  const cambiar = (campo) => (e) => setDatos((prev) => ({ ...prev, [campo]: e.target.value }));

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    setEnviando(true);
    try {
      // El backend responde igual que el login (token incluido): entramos directamente.
      const sesion = await registrarse(datos);
      login(sesion);
      navigate("/tienda");
    } catch (err) {
      setError(err.mensajeCompleto);
    } finally {
      setEnviando(false);
    }
  };

  const input =
    "w-full px-4 py-3 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600 outline-none transition";

  return (
    <div className="min-h-screen flex flex-col bg-slate-100">
      <BannerSinConexion />
      <div className="flex-1 flex items-center justify-center px-4 py-8">
      <div className="max-w-md w-full bg-white rounded-xl shadow-md p-6 sm:p-8 border border-slate-200">
        <div className="text-center mb-8">
          <h2 className="text-3xl font-bold text-slate-900">Crear cuenta</h2>
          <p className="text-slate-500 mt-2">Únete a MultiCatálogo y empieza tu propia red</p>
        </div>

        {error && (
          <div className="bg-red-50 text-red-600 p-3 rounded-lg text-sm mb-6 text-center border border-red-200">
            {error}
          </div>
        )}

        {/* Mobile-first: un campo por fila */}
        <form onSubmit={handleSubmit} className="space-y-5">
          <div>
            <label htmlFor="reg-nombre" className="block text-sm font-medium text-slate-700 mb-2">
              Nombre completo *
            </label>
            <input
              id="reg-nombre"
              className={input}
              value={datos.nombre}
              onChange={cambiar("nombre")}
              autoComplete="name"
              required
            />
          </div>
          <div>
            <label htmlFor="reg-email" className="block text-sm font-medium text-slate-700 mb-2">
              Correo electrónico *
            </label>
            <input
              type="email"
              className={input}
              id="reg-email"
              value={datos.email}
              onChange={cambiar("email")}
              autoComplete="email"
              required
            />
          </div>
          <div>
            <label htmlFor="reg-password" className="block text-sm font-medium text-slate-700 mb-2">
              Contraseña * (mínimo 6 caracteres)
            </label>
            <input
              type="password"
              className={input}
              id="reg-password"
              value={datos.password}
              onChange={cambiar("password")}
              minLength={6}
              autoComplete="new-password"
              required
            />
          </div>
          <div>
            <label htmlFor="reg-ref" className="block text-sm font-medium text-slate-700 mb-2">
              Código de referido (opcional)
            </label>
            <input
              className={`${input} uppercase font-mono`}
              id="reg-ref"
              value={datos.ref}
              onChange={cambiar("ref")}
              placeholder="Ej. LUIS"
            />
            <p className="text-xs text-slate-500 mt-1">Si alguien te invitó, escribe su código para unirte a su red.</p>
          </div>

          <button
            type="submit"
            disabled={!online || enviando}
            className="w-full min-h-11 bg-indigo-600 text-white font-bold py-3 rounded-lg hover:bg-indigo-700 transition disabled:opacity-60"
          >
            {enviando ? "Creando cuenta..." : "Crear cuenta"}
          </button>
        </form>

        <p className="text-center text-sm text-slate-600 mt-6">
          ¿Ya tienes cuenta?{" "}
          <Link to="/login" className="text-indigo-600 font-semibold hover:underline">
            Inicia sesión
          </Link>
        </p>
      </div>
      </div>
    </div>
  );
};

export default Registro;
