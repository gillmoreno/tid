import { Link } from "react-router-dom";
import { ArrowUpRight } from "lucide-react";
import type { Item } from "@/types/site";
import { cn, statusColor, statusLabel } from "@/lib/utils";

type IdeaCardProps = {
  item: Item;
};

export function IdeaCard({ item }: IdeaCardProps) {
  const inner = (
    <>
      <div className="flex items-start justify-between gap-4">
        <div>
          <h3 className="font-display text-lg font-bold text-white">{item.title}</h3>
          <p className="mt-1 text-sm text-fog">{item.tagline}</p>
        </div>
        <span
          className={cn(
            "shrink-0 rounded-full border px-2.5 py-0.5 font-mono text-[10px] uppercase tracking-wider",
            statusColor(item.status)
          )}
        >
          {statusLabel(item.status)}
        </span>
      </div>

      <p className="mt-4 text-sm leading-relaxed text-fog">{item.description}</p>

      <div className="mt-4 flex flex-wrap gap-2">
        {item.tags.map((tag) => (
          <span
            key={tag}
            className="rounded-md border border-line bg-ink px-2 py-0.5 font-mono text-[10px] text-fog-light"
          >
            {tag}
          </span>
        ))}
      </div>
    </>
  );

  if (item.external_url) {
    return (
      <a
        href={item.external_url}
        target="_blank"
        rel="noopener noreferrer"
        className="group block rounded-xl border border-line bg-ink-soft p-5 transition hover:border-signal/30 hover:bg-ink-muted"
      >
        {inner}
        <div className="mt-4 flex items-center gap-1 text-sm text-signal">
          Visit <ArrowUpRight className="h-3.5 w-3.5" />
        </div>
      </a>
    );
  }

  return (
    <Link
      to={`/ideas/${item.slug}`}
      className="group block rounded-xl border border-line bg-ink-soft p-5 transition hover:border-signal/30 hover:bg-ink-muted"
    >
      {inner}
    </Link>
  );
}