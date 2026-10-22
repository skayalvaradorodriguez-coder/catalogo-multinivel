import { useRegisterSW } from "virtual:pwa-register/react";

const UNA_HORA = 60 * 60 * 1000;

// Registra el service worker y avisa cuando:
// - la app quedó lista para usarse sin conexión (primera visita), o
// - hay una versión nueva esperando (el usuario decide cuándo recargar).
const AvisoActualizacion = () => {
  const {
    offlineReady: [listaOffline, setListaOffline],
    needRefresh: [hayNueva, setHayNueva],
    updateServiceWorker,
  } = useRegisterSW({
    onRegisteredSW(_url, registro) {
      // Mientras la app esté abierta, revisamos si hay versión nueva cada hora.
      if (registro) setInterval(() => registro.update(), UNA_HORA);
    },
  });

  if (!listaOffline && !hayNueva) return null;

  const cerrar = () => {
    setListaOffline(false);
    setHayNueva(false);
  };

  return (
    <div
      role="alert"
      className="fixed z-50 left-4 right-4 sm:left-auto sm:right-6 sm:w-96 bg-slate-900 text-white rounded-xl shadow-xl p-4"
      style={{ bottom: "calc(1rem + env(safe-area-inset-bottom))" }}
    >
      <p className="font-semibold">{hayNueva ? "Hay una nueva versión de MultiCatálogo" : "Lista para usar sin conexión"}</p>
      <p className="text-sm text-slate-300 mt-1">
        {hayNueva
          ? "Actualiza para usar la última versión. Tu carrito y tu sesión se conservan."
          : "El catálogo que ya visitaste se podrá consultar aunque no tengas internet."}
      </p>
      <div className="flex gap-2 mt-3">
        {hayNueva && (
          <button
            onClick={() => updateServiceWorker(true)}
            className="min-h-11 flex-1 rounded-lg bg-teal-500 text-slate-950 font-semibold hover:bg-teal-400"
          >
            Actualizar
          </button>
        )}
        <button
          onClick={cerrar}
          className="min-h-11 flex-1 rounded-lg border border-slate-600 font-semibold hover:bg-slate-800"
        >
          {hayNueva ? "Más tarde" : "Entendido"}
        </button>
      </div>
    </div>
  );
};

export default AvisoActualizacion;
