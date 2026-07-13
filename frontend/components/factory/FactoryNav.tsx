import { NavLink } from "react-router-dom";
import { cn } from "@/lib/utils";

const tabs = [
  { to: "/factory", label: "Podcast clips", end: true },
  { to: "/factory/articles", label: "Articles", end: false },
  { to: "/factory/ideas", label: "Ideas", end: false },
  { to: "/factory/sources", label: "Sources", end: false },
] as const;

export function FactoryNav() {
  return (
    <nav className="mb-6 flex gap-1 rounded-md border border-line p-0.5">
      {tabs.map((tab) => (
        <NavLink
          key={tab.to}
          to={tab.to}
          end={tab.end}
          className={({ isActive }) =>
            cn(
              "rounded px-4 py-2 text-sm transition",
              isActive
                ? "bg-signal/10 text-signal"
                : "text-fog hover:text-white"
            )
          }
        >
          {tab.label}
        </NavLink>
      ))}
    </nav>
  );
}