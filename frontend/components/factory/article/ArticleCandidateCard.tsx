import { useEffect, useState } from "react";
import { Calendar, Save, Send, Sparkles, Trash2 } from "lucide-react";
import {
  deleteArticleCandidate,
  postNowArticleCandidate,
  rewriteArticleCandidate,
  scheduleArticleCandidate,
  updateArticleCandidate,
} from "@/api/articles";
import { defaultScheduleTime, factoryStatusBadge, toRFC3339 } from "@/components/factory/factory-utils";
import type { ArticleCandidate } from "@/types/factory";

const REFINE_PRESETS = [
  {
    label: "More controversial",
    instruction:
      "Make the post more provocative and skeptical while staying factual. Draw a sharper implication — who wins, who loses, who should worry.",
  },
  {
    label: "Less controversial",
    instruction: "Tone down heat. More measured, analytical, and precise. Keep curiosity without sounding alarmist.",
  },
  {
    label: "More verbose",
    instruction:
      "Add 1-2 sentences of context. Explain the mechanism and why it matters to builders, operators, or small companies.",
  },
  {
    label: "Ask a question",
    instruction:
      "End with a sharp curiosity-driving question that makes the reader want to reply or think.",
  },
  {
    label: "Add my angle",
    instruction:
      "Rewrite with Gil's skeptical-curious voice: what would have to be true for this to matter, and what should a smart reader conclude?",
  },
  {
    label: "Tighter hook",
    instruction: "Rewrite so the first line is a punchy standalone hook — the reframe, the number, or the tension.",
  },
] as const;

interface ArticleCandidateCardProps {
  candidate: ArticleCandidate;
  onUpdated: (candidate: ArticleCandidate) => void;
  onScheduled: () => void;
  onDeleted: (id: string) => void;
}

export function ArticleCandidateCard({ candidate, onUpdated, onScheduled, onDeleted }: ArticleCandidateCardProps) {
  const [postText, setPostText] = useState(candidate.post_text);
  const [scheduleAt, setScheduleAt] = useState(defaultScheduleTime());
  const [saving, setSaving] = useState(false);
  const [scheduling, setScheduling] = useState(false);
  const [postingNow, setPostingNow] = useState(false);
  const [rewriting, setRewriting] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [refineInstruction, setRefineInstruction] = useState("");
  const [message, setMessage] = useState<string | null>(null);

  useEffect(() => {
    setPostText(candidate.post_text);
  }, [candidate]);

  async function handleSave() {
    setSaving(true);
    setMessage(null);
    try {
      const updated = await updateArticleCandidate(candidate.id, { post_text: postText });
      onUpdated(updated);
      setMessage("Saved.");
    } catch {
      setMessage("Save failed.");
    } finally {
      setSaving(false);
    }
  }

  async function handleRefine(instruction: string) {
    const trimmed = instruction.trim();
    if (!trimmed) return;
    setRewriting(true);
    setMessage(null);
    try {
      const updated = await rewriteArticleCandidate(candidate.id, trimmed, { post_text: postText });
      setPostText(updated.post_text);
      onUpdated(updated);
      setRefineInstruction("");
      setMessage("Refined.");
    } catch (e: unknown) {
      const err = e as { response?: { data?: { error?: string } }; message?: string };
      const msg = err?.response?.data?.error || err?.message || "unknown error";
      setMessage(`Refine failed: ${String(msg).slice(0, 120)}`);
    } finally {
      setRewriting(false);
    }
  }

  async function handleSchedule() {
    setScheduling(true);
    setMessage(null);
    try {
      const updated = await scheduleArticleCandidate(candidate.id, toRFC3339(scheduleAt));
      onUpdated(updated);
      onScheduled();
      setMessage("Scheduled.");
    } catch {
      setMessage("Schedule failed.");
    } finally {
      setScheduling(false);
    }
  }

  async function handlePostNow() {
    setPostingNow(true);
    setMessage(null);
    try {
      const updated = await postNowArticleCandidate(candidate.id, { post_text: postText });
      onUpdated(updated);
      setMessage("Post text copied — X compose opened. Cmd+V and post.");
    } catch {
      setMessage("Post now failed.");
    } finally {
      setPostingNow(false);
    }
  }

  async function handleDelete() {
    if (!window.confirm("Remove this post candidate?")) return;
    setDeleting(true);
    try {
      await deleteArticleCandidate(candidate.id);
      onDeleted(candidate.id);
    } catch {
      setMessage("Delete failed.");
      setDeleting(false);
    }
  }

  return (
    <article className="rounded-lg border border-line bg-ink-soft p-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <div className="flex flex-wrap items-center gap-2">
            <span className="font-mono text-xs text-signal">#{candidate.rank}</span>
            <span className={factoryStatusBadge(candidate.status)}>{candidate.status}</span>
            <span className="font-mono text-xs text-fog">{Math.round(candidate.confidence * 100)}% conf</span>
          </div>
          <p className="mt-2 text-sm leading-relaxed text-fog">{candidate.why_interesting}</p>
        </div>
        <button
          type="button"
          onClick={handleDelete}
          disabled={deleting}
          className="rounded border border-red-400/30 p-2 text-red-400 transition hover:bg-red-400/10 disabled:opacity-50"
          title="Remove candidate"
        >
          <Trash2 className="h-4 w-4" />
        </button>
      </div>

      <div className="mt-4 space-y-3">
        <div>
          <label className="font-mono text-xs uppercase tracking-wider text-fog">Post text</label>
          <textarea
            value={postText}
            onChange={(e) => setPostText(e.target.value)}
            rows={10}
            className="mt-2 w-full resize-y rounded-md border border-line bg-ink px-4 py-3 text-base leading-7 text-fog-light focus:border-signal/50 focus:outline-none"
          />
        </div>

        <div className="rounded-md border border-line/80 bg-ink/50 p-3">
          <label className="font-mono text-xs uppercase tracking-wider text-signal">Refine</label>
          <p className="mt-1 text-sm text-fog">Add your angle — uses your biases + mentions dictionary.</p>
          <div className="mt-3 flex flex-wrap gap-2">
            {REFINE_PRESETS.map((preset) => (
              <button
                key={preset.label}
                type="button"
                onClick={() => handleRefine(preset.instruction)}
                disabled={rewriting}
                className="rounded border border-line px-3 py-1.5 text-sm text-fog-light transition hover:border-signal/40 hover:text-white disabled:opacity-50"
              >
                {preset.label}
              </button>
            ))}
          </div>
          <div className="mt-3 flex flex-wrap gap-2">
            <input
              type="text"
              value={refineInstruction}
              onChange={(e) => setRefineInstruction(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  e.preventDefault();
                  handleRefine(refineInstruction);
                }
              }}
              placeholder="Custom instruction — e.g. frame it around what this means for small teams"
              className="min-w-[240px] flex-1 rounded-md border border-line bg-ink px-3 py-2.5 text-sm text-white placeholder:text-fog/60 focus:border-signal/50 focus:outline-none"
            />
            <button
              type="button"
              onClick={() => handleRefine(refineInstruction)}
              disabled={rewriting || !refineInstruction.trim()}
              className="inline-flex items-center gap-1.5 rounded bg-signal/10 px-3 py-2 text-xs text-signal transition hover:bg-signal/20 disabled:opacity-50"
            >
              <Sparkles className="h-3 w-3" />
              {rewriting ? "Refining…" : "Apply"}
            </button>
          </div>
        </div>
      </div>

      <div className="mt-4 flex flex-wrap items-center gap-2 border-t border-line pt-4">
        <button
          type="button"
          onClick={handleSave}
          disabled={saving}
          className="inline-flex items-center gap-1.5 rounded border border-line px-3 py-1.5 text-xs text-white transition hover:border-signal/40 disabled:opacity-50"
        >
          <Save className="h-3 w-3" />
          {saving ? "Saving…" : "Save edits"}
        </button>

        <div className="flex flex-1 flex-wrap items-center justify-end gap-2">
          <input
            type="datetime-local"
            value={scheduleAt}
            onChange={(e) => setScheduleAt(e.target.value)}
            className="h-[30px] rounded-md border border-line bg-ink px-2 font-mono text-xs text-white focus:border-signal/50 focus:outline-none"
          />
          <button
            type="button"
            onClick={handlePostNow}
            disabled={postingNow}
            className="inline-flex h-[30px] items-center gap-1.5 rounded bg-emerald-500/15 px-3 text-xs text-emerald-400 transition hover:bg-emerald-500/25 disabled:opacity-50"
          >
            <Send className="h-3 w-3" />
            {postingNow ? "Preparing…" : "Post now"}
          </button>
          <button
            type="button"
            onClick={handleSchedule}
            disabled={scheduling}
            className="inline-flex h-[30px] items-center gap-1.5 rounded bg-signal/10 px-3 text-xs text-signal transition hover:bg-signal/20 disabled:opacity-50"
          >
            <Calendar className="h-3 w-3" />
            {scheduling ? "Scheduling…" : "Schedule post"}
          </button>
        </div>
      </div>

      {message && <p className="mt-2 font-mono text-[10px] text-fog">{message}</p>}
    </article>
  );
}
