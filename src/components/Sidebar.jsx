import { Link, useLocation } from "react-router-dom";
import { LayoutDashboard, Grid3x3, Users, Store } from "lucide-react";
import { useAuth } from "../context/useAuth";

const Sidebar = ({ isCollapsed, isMobileOpen, setIsMobileOpen }) => {
  const { user } = useAuth();
  const { pathname } = useLocation();

  const links = [
    { to: "/", label: "Dashboard", icon: LayoutDashboard, soloAdmin: true },
    { to: "/tienda", label: "Tienda", icon: Store, soloAdmin: false },
    { to: "/catalogo", label: "Catálogo", icon: Grid3x3, soloAdmin: false },
    { to: "/mi-red", label: "Mi Red", icon: Users, soloAdmin: true },
  ];

  const visibles = links.filter((link) => !link.soloAdmin || user?.rol === "admin");

  return (
    <aside
      className={`bg-teal-700 text-white flex flex-col transition-all duration-300 z-40
        fixed inset-y-0 left-0 h-full
        md:static md:h-auto md:translate-x-0
        ${isMobileOpen ? "translate-x-0" : "-translate-x-full"}
        ${isCollapsed ? "md:w-20" : "md:w-64"}
        w-64`}
    >
      <div className="p-6 text-2xl font-bold border-b border-teal-600 truncate">
        <span className={isCollapsed ? "md:hidden" : ""}>MultiCatálogo</span>
        {isCollapsed && <span className="hidden md:inline">MC</span>}
      </div>
      <nav className="flex-1 p-4 space-y-2">
        {visibles.map(({ to, label, icon: Icon }) => {
          const esActivo = to === "/" ? pathname === "/" : pathname.startsWith(to);
          return (
            <Link
              key={to}
              to={to}
              onClick={() => setIsMobileOpen && setIsMobileOpen(false)}
              className={`flex items-center gap-3 p-3 rounded transition ${
                esActivo ? "bg-teal-900" : "hover:bg-teal-600"
              } ${isCollapsed ? "md:justify-center" : ""}`}
              title={isCollapsed ? label : undefined}
            >
              <Icon size={20} />
              <span className={isCollapsed ? "md:hidden" : ""}>{label}</span>
            </Link>
          );
        })}
      </nav>
      <div className="p-4 border-t border-teal-600 text-xs text-teal-100">
        {isCollapsed ? (
          <p className="text-center uppercase">{user?.rol}</p>
        ) : (
          <p>
            Conectado como <span className="font-semibold uppercase">{user?.rol}</span>
          </p>
        )}
      </div>
    </aside>
  );
};

export default Sidebar;

