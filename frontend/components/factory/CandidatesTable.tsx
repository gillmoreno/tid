import { ExternalLink, Pencil, Trash2 } from "lucide-react";
import { factoryStatusBadge } from "@/components/factory/factory-utils";
import { cn } from "@/lib/utils";
import type { Candidate } from "@/types/factory";

interface CandidatesTableProps {
  candidates: Candidate[];
  selectedId: string | null;
  onSelect: (id: string) => void;
  onDelete: (id: string) => void;
  deletingId: string | null;
}

function truncate(text: string, max: number): string {
  const t = text.trim();
  if (t.length <= max) return t;
  return `${t.slice(0, max)}…`;
}

export function CandidatesTable({
  candidates,
  selectedId,
  onSelect,
  onDelete,
  deletingId,
}: CandidatesTableProps) {
  return (
    <div className="overflow-x-auto rounded-lg border border-line bg-ink-soft">
      <table className="w-full min-w-[720px] text-left text-sm">
        <thead>
          <tr className="border-b border-line font-mono text-[10px] uppercase tracking-wider text-fog">
            <th className="px-4 py-3 font-normal">#</th>
            <th className="px-4 py-3 font-normal">Hook</th>
            <th className="px-4 py-3 font-normal">Clip</th>
            <th className="px-4 py-3 font-normal">Status</th>
            <th className="px-4 py-3 font-normal">Conf</th>
            <th className="px-4 py-3 text-right font-normal">Actions</th>
          </tr>
        </thead>
        <tbody>
          {candidates.map((candidate) => {
            const selected = candidate.id === selectedId;
            return (
              <tr
                key={candidate.id}
                onClick={() => onSelect(candidate.id)}
                className={cn(
                  "cursor-pointer border-b border-line/60 transition last:border-0 hover:bg-ink/60",
                  selected && "bg-signal/5"
                )}
              >
                <td className="px-4 py-3 font-mono text-xs text-signal">#{candidate.rank}</td>
                <td className="px-4 py-3">
                  <p className="font-medium text-white">{truncate(candidate.hook, 72)}</p>
                  <p className="mt-0.5 font-mono text-[10px] text-fog">
                    {candidate.start_time} → {candidate.end_time}
                  </p>
                </td>
                <td className="px-4 py-3 font-mono text-[10px] text-fog">
                  {candidate.clip_path ? (
                    <span className="text-emerald-400">Ready</span>
                  ) : (
                    <span>—</span>
                  )}
                </td>
                <td className="px-4 py-3">
                  <span className={factoryStatusBadge(candidate.status)}>{candidate.status}</span>
                </td>
                <td className="px-4 py-3 font-mono text-xs text-fog">
                  {Math.round(candidate.confidence * 100)}%
                </td>
                <td className="px-4 py-3">
                  <div className="flex items-center justify-end gap-1">
                    <button
                      type="button"
                      title="Open & edit"
                      onClick={(e) => {
                        e.stopPropagation();
                        onSelect(candidate.id);
                      }}
                      className="inline-flex items-center gap-1 rounded border border-line px-2 py-1 text-[11px] text-fog-light transition hover:border-signal/40 hover:text-white"
                    >
                      <Pencil className="h-3 w-3" />
                      Edit
                    </button>
                    <button
                      type="button"
                      title="Remove candidate"
                      disabled={deletingId === candidate.id}
                      onClick={(e) => {
                        e.stopPropagation();
                        if (window.confirm(`Remove candidate #${candidate.rank}?`)) {
                          onDelete(candidate.id);
                        }
                      }}
                      className="inline-flex items-center gap-1 rounded border border-red-400/30 px-2 py-1 text-[11px] text-red-400 transition hover:bg-red-400/10 disabled:opacity-50"
                    >
                      <Trash2 className="h-3 w-3" />
                      {deletingId === candidate.id ? "…" : "Remove"}
                    </button>
                  </div>
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
      {selectedId && (
        <p className="border-t border-line px-4 py-2 font-mono text-[10px] text-fog">
          <ExternalLink className="mr-1 inline h-3 w-3" />
          Editing selected candidate below
        </p>
      )}
    </div>
  );
}