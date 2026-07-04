import { useEffect, useState } from "react";
import { Hero } from "@/components/sections/Hero";
import { SectionCard } from "@/components/sections/SectionCard";
import { IdeaCard } from "@/components/sections/IdeaCard";
import { fetchSections, fetchSite } from "@/api/site";
import type { Item, Section, Site } from "@/types/site";
import { apiClient } from "@/api/client";

export function HomePage() {
  const [site, setSite] = useState<Site | null>(null);
  const [sections, setSections] = useState<Section[]>([]);
  const [featured, setFeatured] = useState<Item[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function load() {
      try {
        const [siteData, sectionsData] = await Promise.all([fetchSite(), fetchSections()]);
        setSite(siteData);
        setSections(sectionsData);

        const sectionDetails = await Promise.all(
          sectionsData.map((s) => apiClient.get<{ items: Item[] }>(`/sections/${s.slug}`))
        );
        const allItems = sectionDetails.flatMap((r) => r.data.items);
        setFeatured(allItems.filter((i) => i.featured));
      } catch {
        setError("Could not reach the API. Start the backend with `just dev`.");
      }
    }
    load();
  }, []);

  if (error) {
    return (
      <div className="mx-auto max-w-6xl px-6 py-24 text-center">
        <p className="font-mono text-sm text-signal">{error}</p>
      </div>
    );
  }

  if (!site) {
    return (
      <div className="mx-auto max-w-6xl px-6 py-24">
        <div className="h-8 w-48 animate-pulse rounded bg-ink-muted" />
      </div>
    );
  }

  return (
    <>
      <Hero site={site} />

      <section className="mx-auto max-w-6xl px-6 pb-20">
        <div className="mb-8 flex items-end justify-between border-b border-line pb-4">
          <div>
            <h2 className="font-display text-2xl font-bold text-white">What I&apos;m testing</h2>
            <p className="mt-1 text-sm text-fog">Three lanes. One lab.</p>
          </div>
        </div>

        <div className="grid gap-5 md:grid-cols-3">
          {sections.map((section, i) => (
            <SectionCard key={section.slug} section={section} index={i} />
          ))}
        </div>
      </section>

      <section className="mx-auto max-w-6xl px-6 pb-24">
        <div className="mb-8 border-b border-line pb-4">
          <h2 className="font-display text-2xl font-bold text-white">Active experiments</h2>
          <p className="mt-1 text-sm text-fog">Ideas with status — live, testing, or building.</p>
        </div>

        <div className="grid gap-4 md:grid-cols-2">
          {featured.map((item) => (
            <IdeaCard key={item.slug} item={item} />
          ))}
        </div>
      </section>
    </>
  );
}