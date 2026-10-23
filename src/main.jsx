import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
// Antes de montar React: el navegador puede avisar "se puede instalar" apenas carga la página.
import "./pwa/instalacion";
import App from "./App.jsx";

createRoot(document.getElementById("root")).render(
  <StrictMode>
    <App />
  </StrictMode>
);
