// URL base del backend en Go. Se configura con VITE_API_URL en el archivo .env (ver .env.example).
export const API_URL = import.meta.env.VITE_API_URL ?? "http://localhost:3000";
