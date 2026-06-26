import { NavLink, Outlet } from "react-router-dom";
import { cn } from "@/lib/utils";
import { Settings, Grid3x3, Building2, CreditCard, Users, Dice5 } from "lucide-react";
import { Separator } from "@/components/ui/separator";

const nav = [
  { to: "/admin/config",     label: "Game Config",  icon: Settings },
  { to: "/admin/tiles",      label: "Board Tiles",  icon: Grid3x3 },
  { to: "/admin/properties", label: "Properties",   icon: Building2 },
  { to: "/admin/cards",      label: "Cards",        icon: CreditCard },
  { to: "/admin/players",    label: "Players",      icon: Users },
];

export default function AdminLayout() {
  return (
    <div className="flex h-screen bg-background">
      {/* Sidebar */}
      <aside className="w-56 shrink-0 border-r flex flex-col">
        <div className="flex items-center gap-2 px-4 h-14 border-b">
          <Dice5 className="h-5 w-5 text-primary" />
          <span className="font-semibold text-sm">Monopoly Admin</span>
        </div>

        <nav className="flex-1 px-2 py-3 space-y-0.5">
          {nav.map(({ to, label, icon: Icon }) => (
            <NavLink
              key={to}
              to={to}
              className={({ isActive }) =>
                cn(
                  "flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors",
                  isActive
                    ? "bg-primary text-primary-foreground"
                    : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                )
              }
            >
              <Icon className="h-4 w-4" />
              {label}
            </NavLink>
          ))}
        </nav>

        <Separator />
        <p className="px-4 py-3 text-xs text-muted-foreground">shadovx.me platform</p>
      </aside>

      {/* Main */}
      <main className="flex-1 overflow-auto">
        <Outlet />
      </main>
    </div>
  );
}
