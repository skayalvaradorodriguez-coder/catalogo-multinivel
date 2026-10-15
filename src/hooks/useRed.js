import { useEffect, useState } from "react";
import { useAuth } from "../context/useAuth";
import { getRed } from "../services/redService";

// Carga la red multinivel del usuario conectado y expone los estados de carga y error.
export const useRed = () => {
  const { user } = useAuth();
  const email = user?.email;
  const [red, setRed] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    let activo = true;
    getRed(email)
      .then((data) => {
        if (activo) setRed(data);
      })
      .catch(() => {
        if (activo) setError("No se pudo cargar la red. Verifica que el backend esté corriendo.");
      })
      .finally(() => {
        if (activo) setIsLoading(false);
      });
    return () => {
      activo = false;
    };
  }, [email]);

  return { red, isLoading, error };
};
