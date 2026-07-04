import { Play, RefreshCw } from "lucide-react";
import { analyzeSource } from "@/api/factory";
import type { Source } from "@/types/factory";
import { factoryStatusBadge } from "@/components/factory/factory-utils";
import { cn } from "@/lib/utils";

interface SourceListProps {
  sources: Source[];
  selectedId: string | null;
  analyzingId: string | null;
  onSelect: (id: string) => void;
  onAnalyzed: () => void;
  onAnalyzing: (id: string | null) => void;
}

export function SourceList({
  sources,
  selectedId,
  analyzingId,
  onSelect,
  onAnalyzed,
  onAnalyzing,
}: SourceListProps) {
  async function handleAnalyze(source: Source) {
    onAnalyzing(source.id);
    try {
      await analyzeSource(source.id);
      onAnalyzed();
    } catch {
      onAnalyzed();
    } finally {
      onAnalyzing(null);
    }
  }

  if (sources.length === 0) {
    return (
      <div className="rounded-lg border border-dashed border-line px-5 py-8 text-center">
        <p className="text-sm text-fog">No sources yet. Ingest a YouTube URL above.</p>
      </div>
    );
  }

  return (
    <div className="rounded-lg border border-line bg-ink-soft">
      <div className="border-b border-line px-5 py-3">
        <h3 className="font-display text-sm font-bold text-white">Sources</h3>
      </div>
      <ul className="divide-y divide-line">
        {sources.map((source) => {
          const isSelected = selectedId === source.id;
          const isAnalyzing = analyzingId === source.id;

          return (
            <li
              key={source.id}
              className={cn(
                "flex items-start gap-3 px-5 py-3 transition hover:bg-ink-muted/50",
                isSelected && "bg-ink-muted"
              )}
            >
              <button
                type="button"
                onClick={() => onSelect(source.id)}
                className="min-w-0 flex-1 text-left"
              >
                <div className="flex flex-wrap items-center gap-2">
                  <span className="truncate text-sm font-medium text-white">
                    {source.title || source.podcast || source.id}
                  </span>
                  <span className={factoryStatusBadge(source.status)}>{source.status}</span>
                </div>
                <p className="mt-1 truncate font-mono text-[10px] text-fog">{source.youtube_url}</p>
                {source.error_message && (
                  <p className="mt-1 text-xs text-red-400">{source.error_message}</p>
                )}
              </button>
              <button
                type="button"
                onClick={() => handleAnalyze(source)}
                disabled={isAnalyzing}
                className="shrink-0 rounded border border-line p-2 text-fog transition hover:border-signal/40 hover:text-signal disabled:opacity-50"
                title="Analyze transcript"
              >
                {isAnalyzing ? (
                  <RefreshCw className="h-4 w-4 animate-spin" />
                ) : (
                  <Play className="h-4 w-4" />
                )}
              </button>
            </li>
          );
        })}
      </ul>
    </div>
  );
}