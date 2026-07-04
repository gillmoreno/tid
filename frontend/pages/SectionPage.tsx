import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { ArrowLeft } from "lucide-react";
import { fetchSection } from "@/api/site";
import { IdeaCard } from "@/components/sections/IdeaCard";
import type { SectionDetail } from "@/types/site";

export function SectionPage() {
  const { slug } = useParams<{ slug: string }>();
  const [section, setSection] = useState<SectionDetail | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!slug) return;
    fetchSection(slug)
      .then(setSection)
      .catch(() => setError("Section not found."));
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

  if (!section) {
    return (
      <div className="mx-auto max-w-6xl px-6 py-24">
        <div className="h-10 w-64 animate-pulse rounded bg-ink-muted" />
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-6xl px-6 py-12 md:py-16">
      <Link
        to="/"
        className="mb-8 inline-flex items-center gap-2 text-sm text-fog transition hover:text-white"
      >
        <ArrowLeft className="h-4 w-4" />
        Home
      </Link>

      <header className="max-w-2xl">
        <p className="font-mono text-xs uppercase tracking-[0.15em] text-signal">
          Section 0{section.order}
        </p>
        <h1 className="mt-2 font-display text-4xl font-extrabold text-white">{section.title}</h1>
        <p className="mt-2 font-mono text-sm text-signal">{section.subtitle}</p>
        <p className="mt-4 text-base leading-relaxed text-fog">{section.description}</p>

        {section.external_url && (
          <a
            href={section.external_url}
            target="_blank"
            rel="noopener noreferrer"
            className="mt-6 inline-flex rounded-lg border border-signal/40 bg-signal/10 px-4 py-2 text-sm font-medium text-signal transition hover:bg-signal/20"
          >
            Visit {section.subtitle}
          </a>
        )}
      </header>

      <div className="mt-12 grid gap-4 md:grid-cols-2">
        {section.items.length === 0 ? (
          <p className="text-fog">Nothing here yet — experiments incoming.</p>
        ) : (
          section.items.map((item) => <IdeaCard key={item.slug} item={item} />)
        )}
      </div>
    </div>
  );
}