import { Link, NavLink } from "react-router-dom";
import { Logo } from "@/components/brand/Logo";
import { cn } from "@/lib/utils";

const navItems = [
  { to: "/sections/ai-guild", label: "AI Guild" },
  { to: "/sections/loops-with-taste", label: "Loops" },
  { to: "/sections/software-ideas", label: "Software" },
];

export function Header() {
  return (
    <header className="sticky top-0 z-50 border-b border-line/60 bg-ink/80 backdrop-blur-md">
      <div className="mx-auto flex h-16 max-w-6xl items-center justify-between px-6">
        <Link to="/" className="group flex items-center gap-3">
          <Logo size="sm" className="transition group-hover:drop-shadow-[0_0_8px_rgba(0,255,65,0.45)]" />
          <span className="font-display text-sm font-bold tracking-tight text-white">
            The Idea Guy
          </span>
        </Link>

        <nav className="hidden items-center gap-1 sm:flex">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) =>
                cn(
                  "rounded-md px-3 py-1.5 text-sm text-fog transition hover:text-white",
                  isActive && "bg-ink-muted text-signal"
                )
              }
            >
              {item.label}
            </NavLink>
          ))}
        </nav>
      </div>
    </header>
  );
}