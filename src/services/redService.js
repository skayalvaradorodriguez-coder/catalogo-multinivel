// Red multinivel. La identidad sale del token, no de la URL.
import { apiFetch } from "./http";

// Subárbol del usuario conectado (cada usuario ve SU red).
export const getMiRed = () => apiFetch("/api/red/mia");

// Árbol completo (solo admin).
export const getRedGlobal = () => apiFetch("/api/red");
