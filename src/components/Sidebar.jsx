import { Link } from "react-router-dom";
import { LayoutDashboard, Grid3x3, Users } from "lucide-react";

const Sidebar = ({ isCollapsed, isMobileOpen, setIsMobileOpen }) => {
  const links = [
    { to: "/", label: "Dashboard", icon: LayoutDashboard },
    { to: "/catalogo", label: "Catálogo", icon: Grid3x3 },
    { to: "/mi-red", label: "Mi Red", icon: Users },
  ];

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
        {links.map(({ to, label, icon: Icon }) => (
          <Link
            key={to}
            to={to}
            onClick={() => setIsMobileOpen && setIsMobileOpen(false)}
            className={`flex items-center gap-3 p-3 rounded hover:bg-teal-600 transition ${
              isCollapsed ? "md:justify-center" : ""
            }`}
            title={isCollapsed ? label : undefined}
          >
            <Icon size={20} />
            <span className={isCollapsed ? "md:hidden" : ""}>{label}</span>
          </Link>
        ))}
      </nav>
    </aside>
  );
};

export default Sidebar;