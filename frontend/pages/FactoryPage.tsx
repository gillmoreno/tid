import { useCallback, useEffect, useState } from "react";
import { Factory } from "lucide-react";
import { fetchCandidates, fetchScheduled, fetchSources } from "@/api/factory";
import { CandidateCard } from "@/components/factory/CandidateCard";
import { IngestForm } from "@/components/factory/IngestForm";
import { ScheduledQueue } from "@/components/factory/ScheduledQueue";
import { SettingsPanel } from "@/components/factory/SettingsPanel";
import { SourceList } from "@/components/factory/SourceList";
import type { Candidate, ScheduledPost, Source } from "@/types/factory";

export function FactoryPage() {
  const [sources, setSources] = useState<Source[]>([]);
  const [candidates, setCandidates] = useState<Candidate[]>([]);
  const [scheduled, setScheduled] = useState<ScheduledPost[]>([]);
  const [selectedSourceId, setSelectedSourceId] = useState<string | null>(null);
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
      return;
    }
    refreshCandidates(selectedSourceId).catch(() => {
      setError("Failed to load candidates.");
    });
  }, [selectedSourceId, refreshCandidates]);

  function handleSourceCreated(source: Source) {
    setSources((prev) => [source, ...prev]);
    setSelectedSourceId(source.id);
    setCandidates([]);
  }

  function handleCandidateUpdated(updated: Candidate) {
    setCandidates((prev) => prev.map((c) => (c.id === updated.id ? updated : c)));
  }

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
            YouTube in → biases + prompt → clip candidates + takes → refine → schedule or post now.
          </p>
        </div>
      </div>

      {error && (
        <p className="mb-6 rounded border border-red-400/30 bg-red-400/10 px-4 py-2 font-mono text-xs text-red-400">
          {error}
        </p>
      )}

      <div className="space-y-6">
        <SettingsPanel defaultOpen />

        <div className="grid gap-6 xl:grid-cols-2">
          <IngestForm onCreated={handleSourceCreated} />
          <SourceList
            sources={sources}
            selectedId={selectedSourceId}
            analyzingId={analyzingId}
            onSelect={setSelectedSourceId}
            onAnalyzed={refreshAll}
            onAnalyzing={setAnalyzingId}
          />
        </div>

        <div>
          <div className="mb-4 flex items-center justify-between">
            <h2 className="font-display text-lg font-bold text-white">Candidates</h2>
            <span className="font-mono text-xs text-fog">{candidates.length} clips</span>
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