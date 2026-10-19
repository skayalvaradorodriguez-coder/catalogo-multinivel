import { useEffect, useState } from "react";
import { AuthContext } from "./AuthContextValue";
import { borrarSesion, guardarSesion, leerSesion, EVENTO_SESION_EXPIRADA } from "../services/http";

// La sesión (usuario + JWT) se guarda en localStorage: sobrevive a recargar la página.
export const AuthProvider = ({ children }) => {
  const [sesion, setSesion] = useState(() => leerSesion());

  // Si el backend responde 401 (token vencido o adulterado), apiFetch avisa con este evento.
  useEffect(() => {
    const alExpirar = () => setSesion(null);
    window.addEventListener(EVENTO_SESION_EXPIRADA, alExpirar);
    return () => window.removeEventListener(EVENTO_SESION_EXPIRADA, alExpirar);
  }, []);

  // respuesta = { token, email, rol, nombre } de /api/login o /api/register
  const login = ({ token, email, rol, nombre }) => {
    const nueva = { token, usuario: { email, rol: rol === "admin" ? "admin" : "cliente", nombre } };
    guardarSesion(nueva);
    setSesion(nueva);
  };

  const logout = () => {
    borrarSesion();
    setSesion(null);
  };

  return (
    <AuthContext.Provider
      value={{
        isAuthenticated: Boolean(sesion?.token),
        user: sesion?.usuario ?? null,
        token: sesion?.token ?? null,
        login,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};
