import { useCallback, useEffect, useState } from "react";
import { Newspaper } from "lucide-react";
import {
  deleteArticleSource,
  fetchArticleCandidates,
  fetchArticleScheduled,
  fetchArticleSources,
} from "@/api/articles";
import { FactoryNav } from "@/components/factory/FactoryNav";
import { ArticleCandidateCard } from "@/components/factory/article/ArticleCandidateCard";
import { ArticleIngestForm } from "@/components/factory/article/ArticleIngestForm";
import { ArticleScheduledQueue } from "@/components/factory/article/ArticleScheduledQueue";
import { ArticleSettingsPanel } from "@/components/factory/article/ArticleSettingsPanel";
import { ArticleSourceList } from "@/components/factory/article/ArticleSourceList";
import type { ArticleCandidate, ArticleSource } from "@/types/factory";

export function ArticleFactoryPage() {
  const [sources, setSources] = useState<ArticleSource[]>([]);
  const [candidates, setCandidates] = useState<ArticleCandidate[]>([]);
  const [scheduled, setScheduled] = useState<ArticleCandidate[]>([]);
  const [selectedSourceId, setSelectedSourceId] = useState<string | null>(null);
  const [analyzingId, setAnalyzingId] = useState<string | null>(null);
  const [deletingSourceId, setDeletingSourceId] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const refreshCandidates = useCallback(async (sourceId: string | null) => {
    const data = await fetchArticleCandidates(sourceId ?? undefined);
    setCandidates(data);
  }, []);

  const refreshScheduled = useCallback(async () => {
    setScheduled(await fetchArticleScheduled());
  }, []);

  const refreshSources = useCallback(async () => {
    const data = await fetchArticleSources();
    setSources(data);
    return data;
  }, []);

  const refreshAll = useCallback(async () => {
    try {
      await Promise.all([refreshSources(), refreshScheduled()]);
      if (selectedSourceId) {
        await refreshCandidates(selectedSourceId);
      }
      setError(null);
    } catch {
      setError("Could not reach the Article Factory API. Start the backend with `just dev`.");
    }
  }, [refreshSources, refreshScheduled, refreshCandidates, selectedSourceId]);

  useEffect(() => {
    async function init() {
      setLoading(true);
      try {
        const srcs = await fetchArticleSources();
        setSources(srcs);
        const firstId = srcs[0]?.id ?? null;
        setSelectedSourceId(firstId);
        const [cands, sched] = await Promise.all([
          fetchArticleCandidates(firstId ?? undefined),
          fetchArticleScheduled(),
        ]);
        setCandidates(cands);
        setScheduled(sched);
      } catch {
        setError("Could not reach the Article Factory API. Start the backend with `just dev`.");
      } finally {
        setLoading(false);
      }
    }
    init();
  }, []);

  useEffect(() => {
    if (!selectedSourceId) {
      setCandidates([]);
      return;
    }
    refreshCandidates(selectedSourceId).catch(() => setError("Failed to load candidates."));
  }, [selectedSourceId, refreshCandidates]);

  function handleSourceCreated(source: ArticleSource) {
    setSources((prev) => [source, ...prev]);
    setSelectedSourceId(source.id);
    setCandidates([]);
  }

  function handleCandidateUpdated(updated: ArticleCandidate) {
    setCandidates((prev) => prev.map((c) => (c.id === updated.id ? updated : c)));
  }

  function handleCandidateDeleted(id: string) {
    setCandidates((prev) => prev.filter((c) => c.id !== id));
    refreshScheduled();
  }

  async function handleSourceDeleted(id: string) {
    setDeletingSourceId(id);
    try {
      await deleteArticleSource(id);
      const remaining = sources.filter((s) => s.id !== id);
      setSources(remaining);
      if (selectedSourceId === id) {
        const nextId = remaining[0]?.id ?? null;
        setSelectedSourceId(nextId);
        if (!nextId) setCandidates([]);
      }
      await refreshScheduled();
    } catch {
      setError("Failed to remove article.");
    } finally {
      setDeletingSourceId(null);
    }
  }

  return (
    <div className="w-full px-6 py-10 pb-24 xl:px-10">
      <div className="mb-8 flex items-start justify-between gap-4 border-b border-line pb-6">
        <div>
          <div className="flex items-center gap-2">
            <Newspaper className="h-5 w-5 text-signal" />
            <h1 className="font-display text-2xl font-bold text-white">Article Factory</h1>
          </div>
          <p className="mt-2 text-sm text-fog">
            Article URL in → pick publication → standalone X post candidates → refine → schedule or post now.
          </p>
        </div>
      </div>

      <FactoryNav />

      {error && (
        <p className="mb-6 rounded border border-red-400/30 bg-red-400/10 px-4 py-2 font-mono text-xs text-red-400">
          {error}
        </p>
      )}

      <div className="space-y-6">
        <ArticleSettingsPanel />

        <div className="grid gap-6 xl:grid-cols-2">
          <ArticleIngestForm onCreated={handleSourceCreated} />
          <ArticleSourceList
            sources={sources}
            selectedId={selectedSourceId}
            analyzingId={analyzingId}
            deletingId={deletingSourceId}
            onSelect={setSelectedSourceId}
            onDelete={handleSourceDeleted}
            onAnalyzed={refreshAll}
            onAnalyzing={setAnalyzingId}
          />
        </div>

        <div>
          <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
            <h2 className="font-display text-lg font-bold text-white">Post candidates</h2>
            <span className="font-mono text-xs text-fog">{candidates.length} posts</span>
          </div>

          {loading ? (
            <div className="h-32 animate-pulse rounded-lg bg-ink-muted" />
          ) : candidates.length === 0 ? (
            <div className="rounded-lg border border-dashed border-line px-5 py-12 text-center">
              <p className="text-sm text-fog">
                {selectedSourceId
                  ? "No candidates yet. Run analyze on the article."
                  : "Ingest an article to get started."}
              </p>
            </div>
          ) : (
            <div className="space-y-4">
              {candidates.map((candidate) => (
                <ArticleCandidateCard
                  key={candidate.id}
                  candidate={candidate}
                  onUpdated={handleCandidateUpdated}
                  onScheduled={refreshScheduled}
                  onDeleted={handleCandidateDeleted}
                />
              ))}
            </div>
          )}
        </div>

        <ArticleScheduledQueue posts={scheduled} onTick={refreshAll} />
      </div>
    </div>
  );
}
