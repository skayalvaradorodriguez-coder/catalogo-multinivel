import { useState, useEffect } from "react";
import { useCart } from "../context/CartContext";
import { API_URL } from "../config";

const Catalogo = () => {
  const { addToCart } = useCart();
  const [productos, setProductos] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    const fetchProductos = async () => {
      try {
        const response = await fetch(`${API_URL}/api/productos`);
        if (!response.ok) {
          throw new Error("Error al obtener los productos");
        }
        const data = await response.json();
        setProductos(data);
      } catch (err) {
        setError("No se pudo cargar el catálogo. Verifica que el backend esté corriendo.");
      } finally {
        setIsLoading(false);
      }
    };

    fetchProductos();
  }, []);

  return (
    <div>
      <h1 className="text-2xl font-bold text-slate-800 mb-6">
        Catálogo de Productos
      </h1>

      {isLoading && <p className="text-slate-500">Cargando productos...</p>}

      {error && (
        <div className="bg-red-50 text-red-600 p-4 rounded-lg border border-red-200">
          {error}
        </div>
      )}

      {!isLoading && !error && (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
          {productos.map((prod) => (
            <div
              key={prod.id}
              className="bg-white rounded-lg overflow-hidden border border-slate-200 shadow-sm flex flex-col"
            >
              <img
                src={prod.img}
                alt={prod.nombre}
                className="w-full h-40 object-cover"
              />
              <div className="p-4 flex flex-col flex-1">
                <h3 className="font-semibold text-slate-700">{prod.nombre}</h3>
                <p className="text-indigo-600 font-bold mt-2 mb-4">
                  ${prod.precio.toFixed(2)}
                </p>
                <button
                  onClick={() => addToCart(prod)}
                  className="mt-auto w-full bg-slate-900 text-white py-2 rounded text-sm hover:bg-indigo-600 transition"
                >
                  Añadir al Carrito
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default Catalogo;