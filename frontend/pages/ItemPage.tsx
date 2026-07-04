import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { ArrowLeft } from "lucide-react";
import { fetchItem } from "@/api/site";
import type { Item } from "@/types/site";
import { cn, statusColor, statusLabel } from "@/lib/utils";

export function ItemPage() {
  const { slug } = useParams<{ slug: string }>();
  const [item, setItem] = useState<Item | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!slug) return;
    fetchItem(slug)
      .then(setItem)
      .catch(() => setError("Idea not found."));
  }, [slug]);

  if (error) {
    return (
      <div className="mx-auto max-w-6xl px-6 py-24 text-center">
        <p className="text-fog">{error}</p>
        <Link to="/" className="mt-4 inline-block text-signal">
          Back home
        </Link>
      </div>
    );
  }

  if (!item) {
    return (
      <div className="mx-auto max-w-6xl px-6 py-24">
        <div className="h-10 w-64 animate-pulse rounded bg-ink-muted" />
      </div>
    );
  }

  return (
    <article className="mx-auto max-w-3xl px-6 py-12 md:py-16">
      <Link
        to={`/sections/${item.section_slug}`}
        className="mb-8 inline-flex items-center gap-2 text-sm text-fog transition hover:text-white"
      >
        <ArrowLeft className="h-4 w-4" />
        Back to section
      </Link>

      <header>
        <span
          className={cn(
            "inline-block rounded-full border px-2.5 py-0.5 font-mono text-[10px] uppercase tracking-wider",
            statusColor(item.status)
          )}
        >
          {statusLabel(item.status)}
        </span>
        <h1 className="mt-4 font-display text-4xl font-extrabold text-white">{item.title}</h1>
        <p className="mt-2 text-lg text-fog-light">{item.tagline}</p>
      </header>

      <p className="mt-8 text-base leading-relaxed text-fog">{item.description}</p>

      <div className="mt-8 flex flex-wrap gap-2">
        {item.tags.map((tag) => (
          <span
            key={tag}
            className="rounded-md border border-line bg-ink-soft px-2.5 py-1 font-mono text-xs text-fog-light"
          >
            {tag}
          </span>
        ))}
      </div>
    </article>
  );
}