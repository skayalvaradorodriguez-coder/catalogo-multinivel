import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useCart } from "../context/useCart";
import { enviarPedido } from "../services/checkoutService";
import { useOnline } from "../hooks/useOnline";

const METODOS_PAGO = ["Tarjeta de crédito", "Transferencia bancaria", "Efectivo contra entrega"];

const Checkout = () => {
  const { cart, totalPrice, clearCart } = useCart();
  const navigate = useNavigate();
  const online = useOnline();

  const [datos, setDatos] = useState({
    nombre: "",
    email: "",
    direccion: "",
    ciudad: "",
    telefono: "",
    metodoPago: METODOS_PAGO[0],
  });
  const [procesando, setProcesando] = useState(false);
  const [error, setError] = useState("");

  if (cart.length === 0) {
    return (
      <div className="bg-white p-10 rounded-lg border border-slate-200 text-center">
        <p className="text-slate-600 text-lg mb-4">No tienes productos en el carrito.</p>
        <Link to="/catalogo" className="text-indigo-600 font-semibold hover:underline">
          ← Ir al catálogo
        </Link>
      </div>
    );
  }

  const handleChange = (campo, valor) => {
    setDatos((prev) => ({ ...prev, [campo]: valor }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    setProcesando(true);

    try {
      // El backend guarda el pedido en PostgreSQL (BEGIN ... COMMIT). El comprador se identifica
      // con el token de la sesión (apiFetch lo envía solo), no con datos del formulario.
      const { metodoPago, ...cliente } = datos;
      const guardado = await enviarPedido({
        cliente,
        metodoPago,
        items: cart.map((item) => ({ id: item.id, cantidad: item.cantidad })),
      });

      // Usamos los precios y el total que calculó el servidor.
      const pedido = {
        numero: guardado.numero,
        items: guardado.items,
        total: guardado.total,
        cliente: { nombre: guardado.cliente.nombre, email: guardado.cliente.email, ciudad: guardado.cliente.ciudad },
        fecha: new Date(guardado.creadoEn).toLocaleString("es-EC"),
      };

      clearCart();
      navigate("/confirmacion", { state: { pedido } });
    } catch (err) {
      // Si la transacción falló (ROLLBACK), nos quedamos en el formulario y el carrito se conserva.
      setError(err.mensajeCompleto ?? err.message);
    } finally {
      setProcesando(false);
    }
  };

  return (
    <div>
      <h1 className="text-2xl font-bold text-slate-800 mb-6">Finalizar Compra</h1>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <form onSubmit={handleSubmit} className="lg:col-span-2 space-y-5">
          <div className="bg-white p-6 rounded-lg border border-slate-200 shadow-sm">
            <h2 className="text-lg font-bold text-slate-800 mb-4">Datos de Envío</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">Nombre completo *</label>
                <input
                  type="text"
                  required
                  value={datos.nombre}
                  onChange={(e) => handleChange("nombre", e.target.value)}
                  className="w-full px-4 py-2.5 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 outline-none transition"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">Correo electrónico *</label>
                <input
                  type="email"
                  required
                  value={datos.email}
                  onChange={(e) => handleChange("email", e.target.value)}
                  className="w-full px-4 py-2.5 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 outline-none transition"
                />
              </div>
              <div className="md:col-span-2">
                <label className="block text-sm font-medium text-slate-700 mb-1">Dirección *</label>
                <input
                  type="text"
                  required
                  value={datos.direccion}
                  onChange={(e) => handleChange("direccion", e.target.value)}
                  className="w-full px-4 py-2.5 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 outline-none transition"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">Ciudad *</label>
                <input
                  type="text"
                  required
                  value={datos.ciudad}
                  onChange={(e) => handleChange("ciudad", e.target.value)}
                  className="w-full px-4 py-2.5 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 outline-none transition"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">Teléfono *</label>
                <input
                  type="tel"
                  required
                  value={datos.telefono}
                  onChange={(e) => handleChange("telefono", e.target.value)}
                  className="w-full px-4 py-2.5 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 outline-none transition"
                />
              </div>
            </div>
          </div>

          <div className="bg-white p-6 rounded-lg border border-slate-200 shadow-sm">
            <h2 className="text-lg font-bold text-slate-800 mb-4">Método de Pago</h2>
            <div className="space-y-3">
              {METODOS_PAGO.map((metodo) => (
                <label
                  key={metodo}
                  className="flex items-center gap-3 p-3 rounded-lg border border-slate-200 cursor-pointer hover:border-indigo-300 transition"
                >
                  <input
                    type="radio"
                    name="metodoPago"
                    value={metodo}
                    checked={datos.metodoPago === metodo}
                    onChange={(e) => handleChange("metodoPago", e.target.value)}
                    className="accent-indigo-600"
                  />
                  <span className="text-sm text-slate-700">{metodo}</span>
                </label>
              ))}
            </div>
          </div>

          {error && (
            <div className="bg-red-50 text-red-600 p-4 rounded-lg text-sm border border-red-200">
              <p className="font-semibold">No se pudo completar la compra</p>
              <p>{error}</p>
            </div>
          )}

          {!online && (
            <p className="bg-amber-50 text-amber-800 p-4 rounded-lg text-sm border border-amber-200">
              Estás sin conexión. Tu carrito se conserva; podrás confirmar el pedido cuando vuelvas a tener internet.
            </p>
          )}

          <button
            type="submit"
            disabled={procesando || !online}
            className="w-full min-h-11 bg-indigo-600 text-white py-3 rounded-lg font-semibold hover:bg-indigo-700 transition disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {procesando ? "Procesando pago..." : "Confirmar Pedido"}
          </button>
        </form>

        <div className="bg-white p-6 rounded-lg border border-slate-200 shadow-sm h-fit">
          <h2 className="text-lg font-bold text-slate-800 mb-4">Resumen del Pedido</h2>
          <div className="space-y-3 border-b border-slate-100 pb-4 mb-4">
            {cart.map((item) => (
              <div key={item.id} className="flex justify-between text-sm">
                <span className="text-slate-600">
                  {item.nombre} × {item.cantidad}
                </span>
                <span className="font-semibold text-slate-800">
                  ${(item.precio * item.cantidad).toFixed(2)}
                </span>
              </div>
            ))}
          </div>
          <div className="flex justify-between items-center">
            <span className="text-slate-800 font-bold">Total</span>
            <span className="text-2xl font-bold text-indigo-600">${totalPrice.toFixed(2)}</span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Checkout;

