import { useCallback, useEffect, useState } from "react";
import { Lightbulb, Plus } from "lucide-react";
import { createIdea, fetchIdeas } from "@/api/factory";
import { FactoryNav } from "@/components/factory/FactoryNav";
import { IdeaCard } from "@/components/factory/IdeaCard";
import { factoryStatusBadge } from "@/components/factory/factory-utils";
import { cn } from "@/lib/utils";
import type { Idea, IdeaKind } from "@/types/factory";

export function IdeasPage() {
  const [ideas, setIdeas] = useState<Idea[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    const data = await fetchIdeas();
    setIdeas(data);
    setSelectedId((prev) => (prev && data.some((i) => i.id === prev) ? prev : data[0]?.id ?? null));
  }, []);

  useEffect(() => {
    refresh()
      .catch(() => setError("Could not load ideas."))
      .finally(() => setLoading(false));
  }, [refresh]);

  async function handleCreate(kind: IdeaKind = "essay") {
    setCreating(true);
    setError(null);
    try {
      const idea = await createIdea({
        title: "New idea",
        kind,
        summary: "",
      });
      setIdeas((prev) => [idea, ...prev]);
      setSelectedId(idea.id);
    } catch {
      setError("Could not create idea.");
    } finally {
      setCreating(false);
    }
  }

  const selected = ideas.find((i) => i.id === selectedId) ?? null;

  return (
    <div className="w-full px-6 py-10 pb-24 xl:px-10">
      <div className="mb-8 flex items-start justify-between gap-4 border-b border-line pb-6">
        <div>
          <div className="flex items-center gap-2">
            <Lightbulb className="h-5 w-5 text-signal" />
            <h1 className="font-display text-2xl font-bold text-white">Post Factory</h1>
          </div>
          <p className="mt-2 text-sm text-fog">
            Ideas backlog — blog posts, X Articles, glossaries. One draft, multiple surfaces.
          </p>
        </div>
      </div>

      <FactoryNav />

      {error && (
        <p className="mb-6 rounded border border-red-400/30 bg-red-400/10 px-4 py-2 font-mono text-xs text-red-400">
          {error}
        </p>
      )}

      <div className="grid gap-6 xl:grid-cols-[280px_minmax(0,1fr)]">
        <aside className="rounded-lg border border-line bg-ink-soft p-4">
          <div className="mb-3 flex items-center justify-between gap-2">
            <h2 className="font-display text-sm font-bold text-white">Ideas</h2>
            <span className="font-mono text-xs text-fog">{ideas.length}</span>
          </div>

          <div className="mb-3 flex flex-wrap gap-2">
            <button
              type="button"
              onClick={() => handleCreate("essay")}
              disabled={creating}
              className="inline-flex items-center gap-1 rounded border border-line px-2.5 py-1 text-xs text-fog-light transition hover:border-signal/40 hover:text-white disabled:opacity-50"
            >
              <Plus className="h-3 w-3" />
              Essay
            </button>
            <button
              type="button"
              onClick={() => handleCreate("glossary")}
              disabled={creating}
              className="inline-flex items-center gap-1 rounded border border-line px-2.5 py-1 text-xs text-fog-light transition hover:border-signal/40 hover:text-white disabled:opacity-50"
            >
              <Plus className="h-3 w-3" />
              Glossary
            </button>
          </div>

          {loading ? (
            <p className="text-sm text-fog">Loading…</p>
          ) : ideas.length === 0 ? (
            <p className="text-sm text-fog">No ideas yet.</p>
          ) : (
            <ul className="space-y-1">
              {ideas.map((idea) => (
                <li key={idea.id}>
                  <button
                    type="button"
                    onClick={() => setSelectedId(idea.id)}
                    className={cn(
                      "w-full rounded-md border px-3 py-2 text-left transition",
                      selectedId === idea.id
                        ? "border-signal/40 bg-signal/10"
                        : "border-transparent hover:border-line hover:bg-ink/60"
                    )}
                  >
                    <p className="truncate text-sm text-white">{idea.title}</p>
                    <div className="mt-1 flex items-center gap-2">
                      <span className="font-mono text-[10px] uppercase text-fog">{idea.kind}</span>
                      <span className={factoryStatusBadge(idea.status)}>{idea.status}</span>
                    </div>
                  </button>
                </li>
              ))}
            </ul>
          )}
        </aside>

        <section>
          {selected ? (
            <IdeaCard
              idea={selected}
              onUpdated={(updated) =>
                setIdeas((prev) => prev.map((i) => (i.id === updated.id ? updated : i)))
              }
              onDeleted={(id) => {
                setIdeas((prev) => prev.filter((i) => i.id !== id));
                setSelectedId((prev) => (prev === id ? null : prev));
              }}
            />
          ) : (
            <div className="rounded-lg border border-dashed border-line px-6 py-16 text-center">
              <p className="text-sm text-fog">Select an idea or create a new one.</p>
            </div>
          )}
        </section>
      </div>
    </div>
  );
}