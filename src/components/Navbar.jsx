import { useState, useRef, useEffect } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Menu } from "lucide-react";
import { useCart } from "../context/useCart";
import { useAuth } from "../context/useAuth";
import InstalarApp from "./InstalarApp";

const Navbar = ({ isCollapsed, setIsCollapsed, setIsMobileOpen }) => {
  const { totalItems } = useCart();
  const { logout, user } = useAuth();
  const navigate = useNavigate();
  const [isProfileOpen, setIsProfileOpen] = useState(false);
  const profileRef = useRef(null);

  const handleLogout = () => {
    logout();
    navigate("/login");
  };

  const handleMenuClick = () => {
    setIsMobileOpen((prev) => !prev);
    setIsCollapsed(!isCollapsed);
  };

  useEffect(() => {
    const handleClickOutside = (event) => {
      if (profileRef.current && !profileRef.current.contains(event.target)) {
        setIsProfileOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    document.addEventListener("touchstart", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
      document.removeEventListener("touchstart", handleClickOutside);
    };
  }, []);

  return (
    <header
      className="bg-white border-b border-slate-200 flex items-center justify-between px-4 sm:px-6 lg:px-8"
      style={{ paddingTop: "env(safe-area-inset-top)", minHeight: "calc(4rem + env(safe-area-inset-top))" }}
    >
      <div className="flex items-center gap-2 sm:gap-4 min-w-0">
        <button
          onClick={handleMenuClick}
          className="p-2 min-h-11 min-w-11 flex items-center justify-center hover:bg-slate-100 rounded-full transition shrink-0"
          title="Colapsar/Expandir menú"
        >
          <Menu size={20} className="text-slate-600" />
        </button>
        <h2 className="text-slate-600 font-medium text-base sm:text-lg truncate hidden sm:block">
          {user?.rol === "admin" ? "Panel de Administración" : "Tienda MultiCatálogo"}
        </h2>
      </div>
      <div className="flex items-center gap-2 sm:gap-6">
        <Link to="/carrito" className="relative p-2 min-h-11 min-w-11 flex items-center justify-center hover:bg-slate-100 rounded-full transition">
          <span className="text-xl">🛒</span>
          {totalItems > 0 && (
            <span className="absolute top-0 right-0 bg-indigo-600 text-white text-xs font-bold w-5 h-5 flex items-center justify-center rounded-full transform translate-x-1 -translate-y-1">
              {totalItems}
            </span>
          )}
        </Link>
        <div className="flex items-center gap-2 sm:gap-4">
          <span className="text-sm text-slate-500 hidden sm:inline max-w-[160px] truncate">
            {user?.nombre || user?.email}
          </span>
          <span
            className={`hidden md:inline-block text-xs font-semibold px-2 py-1 rounded-full uppercase ${
              user?.rol === "admin" ? "bg-amber-100 text-amber-700" : "bg-indigo-100 text-indigo-700"
            }`}
          >
            {user?.rol}
          </span>
          <div className="relative" ref={profileRef}>
            <button
              onClick={() => setIsProfileOpen((prev) => !prev)}
              className="w-11 h-11 rounded-full bg-slate-200 overflow-hidden border border-slate-300 flex items-center justify-center"
              aria-label="Menú de usuario"
            >
              <img
                src="https://fastly.picsum.photos/id/64/4326/2884.jpg?hmac=9_SzX666YRpR_fOyYStXpfSiJ_edO3ghlSRnH2w09Kg"
                alt="Avatar del usuario"
                className="w-full h-full object-cover"
              />
            </button>
            {isProfileOpen && (
              <div className="absolute right-0 top-full mt-1 w-48 bg-white border border-slate-200 rounded-lg shadow-lg z-50 py-1">
                <p className="px-4 py-2 text-xs text-slate-500 truncate border-b border-slate-100">{user?.email}</p>
                <Link
                  to="/mis-pedidos"
                  onClick={() => setIsProfileOpen(false)}
                  className="flex items-center w-full min-h-11 px-4 text-sm text-slate-700 hover:bg-slate-50"
                >
                  Mis pedidos
                </Link>
                <Link
                  to="/mi-red"
                  onClick={() => setIsProfileOpen(false)}
                  className="flex items-center w-full min-h-11 px-4 text-sm text-slate-700 hover:bg-slate-50"
                >
                  Mi red y comisiones
                </Link>
                <InstalarApp alTerminar={() => setIsProfileOpen(false)} />
                <button
                  onClick={() => {
                    setIsProfileOpen(false);
                    handleLogout();
                  }}
                  className="w-full min-h-11 text-left px-4 text-sm text-red-600 font-semibold hover:bg-red-50 transition-colors"
                >
                  Cerrar Sesión
                </button>
              </div>
            )}
          </div>
        </div>
      </div>
    </header>
  );
};

export default Navbar;


