import { useState, useRef, useEffect } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Menu } from "lucide-react";
import { useCart } from "../context/CartContext";
import { useAuth } from "../context/AuthContext";

const Navbar = ({ isCollapsed, setIsCollapsed, setIsMobileOpen }) => {
  const { totalItems } = useCart();
  const { logout, userEmail } = useAuth();
  const navigate = useNavigate();
  const [isProfileOpen, setIsProfileOpen] = useState(false);
  const profileRef = useRef(null);

  const handleLogout = () => {
    logout();
    navigate("/login");
  };

  const handleMenuClick = () => {
    // En mobile abre/cierra el drawer; en desktop colapsa/expande el sidebar
    setIsMobileOpen((prev) => !prev);
    setIsCollapsed(!isCollapsed);
  };

  // Cierra el menú del perfil al hacer clic/tap fuera de él
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
    <header className="h-16 bg-white border-b border-slate-200 flex items-center justify-between px-4 sm:px-6 lg:px-8">
      <div className="flex items-center gap-2 sm:gap-4 min-w-0">
        <button
          onClick={handleMenuClick}
          className="p-2 hover:bg-slate-100 rounded-full transition shrink-0"
          title="Colapsar/Expandir menú"
        >
          <Menu size={20} className="text-slate-600" />
        </button>
        <h2 className="text-slate-600 font-medium text-base sm:text-lg truncate hidden sm:block">
          Panel de Administración
        </h2>
      </div>
      <div className="flex items-center gap-2 sm:gap-6">
        <Link
          to="/carrito"
          className="relative p-2 hover:bg-slate-100 rounded-full transition"
        >
          <span className="text-xl">🛒</span>
          {totalItems > 0 && (
            <span className="absolute top-0 right-0 bg-indigo-600 text-white text-xs font-bold w-5 h-5 flex items-center justify-center rounded-full transform translate-x-1 -translate-y-1">
              {totalItems}
            </span>
          )}
        </Link>
        <div className="flex items-center gap-2 sm:gap-4">
          <span className="text-sm text-slate-500 hidden sm:inline max-w-[140px] truncate">
            {userEmail}
          </span>
          <div className="relative pb-2" ref={profileRef}>
            <button
              onClick={() => setIsProfileOpen((prev) => !prev)}
              className="w-8 h-8 rounded-full bg-slate-200 overflow-hidden border border-slate-300 flex items-center justify-center"
            >
              <img
                src="https://fastly.picsum.photos/id/64/4326/2884.jpg?hmac=9_SzX666YRpR_fOyYStXpfSiJ_edO3ghlSRnH2w09Kg"
                alt="Avatar del usuario"
                className="w-full h-full object-cover"
              />
            </button>
            {isProfileOpen && (
              <div className="absolute right-0 top-full mt-1 w-36 bg-white border border-slate-200 rounded-lg shadow-lg z-50">
                <button
                  onClick={() => {
                    setIsProfileOpen(false);
                    handleLogout();
                  }}
                  className="w-full text-left px-4 py-2 text-sm text-red-600 font-semibold hover:bg-red-50 rounded-md transition-colors"
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