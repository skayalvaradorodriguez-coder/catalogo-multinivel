// Aviso reutilizable para errores de carga de datos desde la API.
const MensajeError = ({ mensaje }) => (
  <div className="bg-red-50 text-red-600 p-4 rounded-lg text-sm border border-red-200 text-center">
    {mensaje}
  </div>
);

export default MensajeError;
