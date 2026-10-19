import { useApi } from "./useApi";
import { getMiRed } from "../services/redService";

// Carga la red multinivel del usuario conectado (su subárbol) con estados de carga y error.
export const useRed = () => {
  const { data: red, isLoading, error, recargar } = useApi(getMiRed);
  return { red, isLoading, error, recargar };
};
