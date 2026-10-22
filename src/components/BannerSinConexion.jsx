import { useOnline } from "../hooks/useOnline";

// Franja que aparece solo cuando el navegador pierde la conexión.
const BannerSinConexion = () => {
  const online = useOnline();
  if (online) return null;

  return (
    <div
      role="status"
      aria-live="polite"
      className="bg-amber-500 text-amber-950 text-sm font-semibold text-center px-4 py-2"
    >
      Sin conexión — mostrando los datos guardados. Las compras y tu cuenta necesitan internet.
    </div>
  );
};

export default BannerSinConexion;
