import { useState } from "react";
import { Clock, Zap } from "lucide-react";
import { tickScheduler } from "@/api/factory";
import { factoryStatusBadge, formatScheduledAt } from "@/components/factory/factory-utils";
import type { ScheduledPost } from "@/types/factory";

interface ScheduledQueueProps {
  posts: ScheduledPost[];
  onTick: () => void;
}

export function ScheduledQueue({ posts, onTick }: ScheduledQueueProps) {
  const [ticking, setTicking] = useState(false);
  const [result, setResult] = useState<string | null>(null);

  async function handleTick() {
    setTicking(true);
    setResult(null);
    try {
      const res = await tickScheduler();
      onTick();
      if (res.prepared.length === 0) {
        setResult("No posts due right now.");
      } else {
        setResult(`Prepared ${res.prepared.length} post(s) — clipboard + Chrome + Finder.`);
      }
    } catch {
      setResult("Tick failed.");
    } finally {
      setTicking(false);
    }
  }

  return (
    <div className="rounded-lg border border-line bg-ink-soft">
      <div className="flex items-center justify-between border-b border-line px-5 py-3">
        <div className="flex items-center gap-2">
          <Clock className="h-4 w-4 text-signal" />
          <h3 className="font-display text-sm font-bold text-white">Scheduled queue</h3>
        </div>
        <button
          type="button"
          onClick={handleTick}
          disabled={ticking}
          className="inline-flex items-center gap-1.5 rounded border border-signal/30 bg-signal/10 px-3 py-1.5 text-xs text-signal transition hover:bg-signal/20 disabled:opacity-50"
        >
          <Zap className="h-3 w-3" />
          {ticking ? "Running…" : "Tick now"}
        </button>
      </div>

      {result && (
        <p className="border-b border-line px-5 py-2 font-mono text-[10px] text-fog">{result}</p>
      )}

      {posts.length === 0 ? (
        <p className="px-5 py-6 text-sm text-fog">Nothing scheduled yet.</p>
      ) : (
        <ul className="divide-y divide-line">
          {posts.map((post) => {
            const c = post.candidate;
            return (
              <li key={post.id} className="px-5 py-3">
                <div className="flex flex-wrap items-center gap-2">
                  <span className="text-sm text-white">
                    {formatScheduledAt(post.scheduled_at)}
                  </span>
                  <span className={factoryStatusBadge(post.status)}>{post.status}</span>
                </div>
                {c && (
                  <p className="mt-1 truncate text-xs text-fog">
                    {c.hook || c.id}
                  </p>
                )}
                {post.prepared_at && (
                  <p className="mt-1 font-mono text-[10px] text-emerald-400">
                    Prepared {formatScheduledAt(post.prepared_at)}
                  </p>
                )}
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}