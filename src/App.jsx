import { BrowserRouter, Routes, Route, Navigate, Outlet } from "react-router-dom";
import Layout from "./components/Layout";
import Dashboard from "./components/Dashboard";
import Catalogo from "./components/Catalogo";
import Storefront from "./components/Storefront";
import DetalleProducto from "./components/DetalleProducto";
import MiRed from "./components/MiRed";
import Carrito from "./components/Carrito";
import Checkout from "./components/Checkout";
import Confirmacion from "./components/Confirmacion";
import Login from "./components/Login";
import { CartProvider } from "./context/CartContext";
import { AuthProvider } from "./context/AuthContext";
import { useAuth } from "./context/useAuth";

// Exige estar autenticado (cualquier rol)
const ProtectedRoute = () => {
  const { isAuthenticated } = useAuth();
  if (!isAuthenticated) return <Navigate to="/login" replace />;
  return <Outlet />;
};

// Exige rol admin; si un cliente entra, lo mandamos a la tienda
const AdminRoute = () => {
  const { isAuthenticated, user } = useAuth();
  if (!isAuthenticated) return <Navigate to="/login" replace />;
  if (user?.rol !== "admin") return <Navigate to="/tienda" replace />;
  return <Outlet />;
};

// Remonta el CartProvider al cambiar de usuario: cada cuenta lee y persiste
// SU propio carrito (localStorage) sin mezclarse con el de otra cuenta.
const CartBoundary = ({ children }) => {
  const { user } = useAuth();
  return <CartProvider key={user?.email ?? "anonimo"}>{children}</CartProvider>;
};

function App() {
  return (
    <AuthProvider>
      <CartBoundary>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<Login />} />

            <Route element={<ProtectedRoute />}>
              <Route path="/" element={<Layout />}>
                {/* Solo administrador */}
                <Route element={<AdminRoute />}>
                  <Route index element={<Dashboard />} />
                  <Route path="mi-red" element={<MiRed />} />
                </Route>

                {/* Ambos roles */}
                <Route path="tienda" element={<Storefront />} />
                <Route path="catalogo" element={<Catalogo />} />
                <Route path="producto/:id" element={<DetalleProducto />} />
                <Route path="carrito" element={<Carrito />} />
                <Route path="checkout" element={<Checkout />} />
                <Route path="confirmacion" element={<Confirmacion />} />
              </Route>
            </Route>

            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </BrowserRouter>
      </CartBoundary>
    </AuthProvider>
  );
}

export default App;


