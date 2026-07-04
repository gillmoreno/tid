import { Link } from "react-router-dom";
import { ArrowUpRight, Code2, RefreshCw, Users } from "lucide-react";
import type { Section } from "@/types/site";
import { cn } from "@/lib/utils";

const iconMap = {
  guild: Users,
  loop: RefreshCw,
  code: Code2,
} as const;

type SectionCardProps = {
  section: Section;
  index: number;
};

export function SectionCard({ section, index }: SectionCardProps) {
  const Icon = iconMap[section.icon as keyof typeof iconMap] ?? Code2;
  const href = section.external_url ?? `/sections/${section.slug}`;
  const isExternal = Boolean(section.external_url);

  const content = (
    <article
      className={cn(
        "group relative flex h-full flex-col rounded-xl border border-line bg-ink-soft p-6 transition",
        "hover:border-signal/40 hover:bg-ink-muted",
        "animate-fade-up"
      )}
      style={{ animationDelay: `${index * 100}ms`, opacity: 0 }}
    >
      <div className="mb-4 flex items-start justify-between">
        <div className="flex h-10 w-10 items-center justify-center rounded-lg border border-line bg-ink text-signal transition group-hover:border-signal/30 group-hover:bg-signal/5">
          <Icon className="h-5 w-5" />
        </div>
        <span className="font-mono text-xs text-fog/60">0{section.order}</span>
      </div>

      <h3 className="font-display text-xl font-bold text-white">{section.title}</h3>
      <p className="mt-1 font-mono text-xs text-signal">{section.subtitle}</p>
      <p className="mt-3 flex-1 text-sm leading-relaxed text-fog">{section.description}</p>

      <div className="mt-5 flex items-center gap-1 text-sm font-medium text-signal transition group-hover:gap-2">
        Explore
        <ArrowUpRight className="h-4 w-4" />
      </div>
    </article>
  );

  if (isExternal) {
    return (
      <a href={href} target="_blank" rel="noopener noreferrer" className="block h-full">
        {content}
      </a>
    );
  }

  return (
    <Link to={href} className="block h-full">
      {content}
    </Link>
  );
}