import { useCallback, useEffect, useState } from "react";
import { Factory, LayoutGrid, Table2 } from "lucide-react";
import {
  deleteCandidate,
  deleteSource,
  fetchCandidates,
  fetchScheduled,
  fetchSources,
} from "@/api/factory";
import { FactoryNav } from "@/components/factory/FactoryNav";
import { CandidateCard } from "@/components/factory/CandidateCard";
import { CandidatesTable } from "@/components/factory/CandidatesTable";
import { IngestForm } from "@/components/factory/IngestForm";
import { SpotMomentForm } from "@/components/factory/SpotMomentForm";
import { ScheduledQueue } from "@/components/factory/ScheduledQueue";
import { SettingsPanel } from "@/components/factory/SettingsPanel";
import { SourceList } from "@/components/factory/SourceList";
import { cn } from "@/lib/utils";
import type { Candidate, ScheduledPost, Source } from "@/types/factory";

type CandidateView = "cards" | "table";

export function FactoryPage() {
  const [sources, setSources] = useState<Source[]>([]);
  const [candidates, setCandidates] = useState<Candidate[]>([]);
  const [scheduled, setScheduled] = useState<ScheduledPost[]>([]);
  const [selectedSourceId, setSelectedSourceId] = useState<string | null>(null);
  const [selectedCandidateId, setSelectedCandidateId] = useState<string | null>(null);
  const [candidateView, setCandidateView] = useState<CandidateView>("cards");
  const [deletingCandidateId, setDeletingCandidateId] = useState<string | null>(null);
  const [deletingSourceId, setDeletingSourceId] = useState<string | null>(null);
  const [analyzingId, setAnalyzingId] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const refreshSources = useCallback(async () => {
    const data = await fetchSources();
    setSources(data);
    if (!selectedSourceId && data.length > 0) {
      setSelectedSourceId(data[0].id);
    }
  }, [selectedSourceId]);

  const refreshCandidates = useCallback(async (sourceId: string | null) => {
    const data = await fetchCandidates(sourceId ?? undefined);
    setCandidates(data);
    setSelectedCandidateId((prev) => (prev && data.some((c) => c.id === prev) ? prev : null));
  }, []);

  const refreshScheduled = useCallback(async () => {
    setScheduled(await fetchScheduled());
  }, []);

  const refreshAll = useCallback(async () => {
    try {
      await Promise.all([refreshSources(), refreshScheduled()]);
      if (selectedSourceId) {
        await refreshCandidates(selectedSourceId);
      }
      setError(null);
    } catch {
      setError("Could not reach the Post Factory API. Start the backend with `just dev` or `just dev-local`.");
    }
  }, [refreshSources, refreshScheduled, refreshCandidates, selectedSourceId]);

  useEffect(() => {
    async function init() {
      setLoading(true);
      try {
        const srcs = await fetchSources();
        setSources(srcs);
        const firstId = srcs[0]?.id ?? null;
        setSelectedSourceId(firstId);
        const [cands, sched] = await Promise.all([
          fetchCandidates(firstId ?? undefined),
          fetchScheduled(),
        ]);
        setCandidates(cands);
        setScheduled(sched);
      } catch {
        setError("Could not reach the Post Factory API. Start the backend with `just dev` or `just dev-local`.");
      } finally {
        setLoading(false);
      }
    }
    init();
  }, []);

  useEffect(() => {
    if (!selectedSourceId) {
      setCandidates([]);
      setSelectedCandidateId(null);
      return;
    }
    setSelectedCandidateId(null);
    refreshCandidates(selectedSourceId).catch(() => {
      setError("Failed to load candidates.");
    });
  }, [selectedSourceId, refreshCandidates]);

  function handleSourceCreated(source: Source) {
    setSources((prev) => [source, ...prev]);
    setSelectedSourceId(source.id);
    setCandidates([]);
    setSelectedCandidateId(null);
  }

  function handleCandidateUpdated(updated: Candidate) {
    setCandidates((prev) => prev.map((c) => (c.id === updated.id ? updated : c)));
  }

  async function handleSourceDeleted(id: string) {
    setDeletingSourceId(id);
    try {
      await deleteSource(id);
      const remaining = sources.filter((s) => s.id !== id);
      setSources(remaining);
      if (selectedSourceId === id) {
        const nextId = remaining[0]?.id ?? null;
        setSelectedSourceId(nextId);
        if (!nextId) {
          setCandidates([]);
          setSelectedCandidateId(null);
        }
      }
      await refreshScheduled();
    } catch {
      setError("Failed to remove source.");
    } finally {
      setDeletingSourceId(null);
    }
  }

  async function handleCandidateDeleted(id: string) {
    setDeletingCandidateId(id);
    try {
      await deleteCandidate(id);
      setCandidates((prev) => prev.filter((c) => c.id !== id));
      setSelectedCandidateId((prev) => (prev === id ? null : prev));
      await refreshScheduled();
    } catch {
      setError("Failed to remove candidate.");
    } finally {
      setDeletingCandidateId(null);
    }
  }

  const selectedCandidate =
    candidates.find((c) => c.id === selectedCandidateId) ?? null;

  if (error && loading) {
    return (
      <div className="w-full px-6 py-24 text-center xl:px-10">
        <p className="font-mono text-sm text-signal">{error}</p>
      </div>
    );
  }

  return (
    <div className="w-full px-6 py-10 pb-24 xl:px-10">
      <div className="mb-8 flex items-start justify-between gap-4 border-b border-line pb-6">
        <div>
          <div className="flex items-center gap-2">
            <Factory className="h-5 w-5 text-signal" />
            <h1 className="font-display text-2xl font-bold text-white">Post Factory</h1>
          </div>
          <p className="mt-2 text-sm text-fog">
            YouTube in → pick podcast → clip candidates + post text → refine → schedule or post now.
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
        <SettingsPanel defaultOpen />

        <SpotMomentForm
          onAnalyzed={(result) => {
            setSelectedSourceId(result.source_id);
            setCandidates((prev) => {
              const ids = new Set(prev.map((c) => c.id));
              const merged = [...prev];
              for (const c of result.candidates) {
                if (!ids.has(c.id)) merged.push(c);
              }
              return merged.sort((a, b) => a.rank - b.rank);
            });
            if (result.candidates[0]) {
              setSelectedCandidateId(result.candidates[0].id);
            }
            refreshAll();
          }}
        />

        <div className="grid gap-6 xl:grid-cols-2">
          <IngestForm onCreated={handleSourceCreated} />
          <SourceList
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
            <h2 className="font-display text-lg font-bold text-white">Candidates</h2>
            <div className="flex items-center gap-3">
              <span className="font-mono text-xs text-fog">{candidates.length} clips</span>
              <div className="flex rounded-md border border-line p-0.5">
                <button
                  type="button"
                  onClick={() => setCandidateView("cards")}
                  className={cn(
                    "inline-flex items-center gap-1.5 rounded px-2.5 py-1 text-xs transition",
                    candidateView === "cards"
                      ? "bg-signal/15 text-signal"
                      : "text-fog hover:text-white"
                  )}
                >
                  <LayoutGrid className="h-3.5 w-3.5" />
                  Cards
                </button>
                <button
                  type="button"
                  onClick={() => setCandidateView("table")}
                  className={cn(
                    "inline-flex items-center gap-1.5 rounded px-2.5 py-1 text-xs transition",
                    candidateView === "table"
                      ? "bg-signal/15 text-signal"
                      : "text-fog hover:text-white"
                  )}
                >
                  <Table2 className="h-3.5 w-3.5" />
                  Table
                </button>
              </div>
            </div>
          </div>

          {loading ? (
            <div className="h-32 animate-pulse rounded-lg bg-ink-muted" />
          ) : candidates.length === 0 ? (
            <div className="rounded-lg border border-dashed border-line px-5 py-12 text-center">
              <p className="text-sm text-fog">
                {selectedSourceId
                  ? "No candidates yet. Run analyze on the source."
                  : "Select or ingest a source to see candidates."}
              </p>
            </div>
          ) : candidateView === "table" ? (
            <div className="space-y-4">
              <CandidatesTable
                candidates={candidates}
                selectedId={selectedCandidateId}
                onSelect={setSelectedCandidateId}
                onDelete={handleCandidateDeleted}
                deletingId={deletingCandidateId}
              />
              {selectedCandidate ? (
                <CandidateCard
                  key={selectedCandidate.id}
                  candidate={selectedCandidate}
                  onUpdated={handleCandidateUpdated}
                  onScheduled={refreshAll}
                />
              ) : (
                <p className="rounded-lg border border-dashed border-line px-5 py-8 text-center text-sm text-fog">
                  Select a row to open and edit a candidate.
                </p>
              )}
            </div>
          ) : (
            <div className="space-y-4">
              {candidates.map((candidate) => (
                <CandidateCard
                  key={candidate.id}
                  candidate={candidate}
                  onUpdated={handleCandidateUpdated}
                  onScheduled={refreshAll}
                />
              ))}
            </div>
          )}
        </div>

        <ScheduledQueue posts={scheduled} onTick={refreshAll} />
      </div>
    </div>
  );
}