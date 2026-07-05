import { useEffect, useState } from "react";
import { Calendar, Film, Save, Scissors, Send, Sparkles } from "lucide-react";
import {
  clipCandidate,
  postNowCandidate,
  rewriteCandidate,
  scheduleCandidate,
  updateCandidate,
} from "@/api/factory";
import {
  defaultScheduleTime,
  factoryStatusBadge,
  toRFC3339,
} from "@/components/factory/factory-utils";
import type { Candidate } from "@/types/factory";

const REFINE_PRESETS = [
  {
    label: "More controversial",
    instruction:
      "Make the take more provocative and skeptical while staying factual. Draw a sharper implication — who wins, who loses, who should worry.",
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
      "End with a sharp curiosity-driving question — e.g. what does this mean, who can we trust, what's the alternative for teams that cannot spend hundreds of thousands on infra.",
  },
  {
    label: "Add my angle",
    instruction:
      "Rewrite with Gil's skeptical-curious voice: what would have to be true for this to matter, and what should a smart reader conclude?",
  },
  {
    label: "Build the argument",
    instruction:
      "Format A: reframe → mechanism → named proof → pattern (list examples) → historical parallel → memorable closer or sharp question. Staccato beats. Tag @ handles from mentions dictionary.",
  },
  {
    label: "Quote mode",
    instruction:
      "Format B: optional topic header, then tight quote in the speaker's voice — story, reversal, thesis, contrast, punchline. Keep vivid phrases. Tag people and companies with @ handles.",
  },
] as const;

interface CandidateCardProps {
  candidate: Candidate;
  onUpdated: (candidate: Candidate) => void;
  onScheduled: () => void;
}

export function CandidateCard({ candidate, onUpdated, onScheduled }: CandidateCardProps) {
  const [hook, setHook] = useState(candidate.hook);
  const [take, setTake] = useState(candidate.take);
  const [postText, setPostText] = useState(candidate.post_text);
  const [scheduleAt, setScheduleAt] = useState(defaultScheduleTime());
  const [saving, setSaving] = useState(false);
  const [clipping, setClipping] = useState(false);
  const [scheduling, setScheduling] = useState(false);
  const [postingNow, setPostingNow] = useState(false);
  const [rewriting, setRewriting] = useState(false);
  const [refineInstruction, setRefineInstruction] = useState("");
  const [message, setMessage] = useState<string | null>(null);

  useEffect(() => {
    setHook(candidate.hook);
    setTake(candidate.take);
    setPostText(candidate.post_text);
  }, [candidate]);

  async function handleSave() {
    setSaving(true);
    setMessage(null);
    try {
      const updated = await updateCandidate(candidate.id, { hook, take, post_text: postText });
      onUpdated(updated);
      setMessage("Saved.");
    } catch {
      setMessage("Save failed.");
    } finally {
      setSaving(false);
    }
  }

  async function handleClip() {
    setClipping(true);
    setMessage(null);
    try {
      const updated = await clipCandidate(candidate.id);
      onUpdated(updated);
      setMessage("Clip ready.");
    } catch {
      setMessage("Clip failed.");
    } finally {
      setClipping(false);
    }
  }

  async function handleSchedule() {
    setScheduling(true);
    setMessage(null);
    try {
      await scheduleCandidate(candidate.id, toRFC3339(scheduleAt));
      onScheduled();
      setMessage("Scheduled.");
    } catch {
      setMessage("Schedule failed.");
    } finally {
      setScheduling(false);
    }
  }

  async function handleRefine(instruction: string) {
    const trimmed = instruction.trim();
    if (!trimmed) return;

    setRewriting(true);
    setMessage(null);
    try {
      const updated = await rewriteCandidate(candidate.id, trimmed, {
        hook,
        take,
        post_text: postText,
      });
      setHook(updated.hook);
      setTake(updated.take);
      setPostText(updated.post_text);
      onUpdated(updated);
      setRefineInstruction("");
      setMessage("Refined.");
    } catch {
      setMessage("Refine failed — is grok available?");
    } finally {
      setRewriting(false);
    }
  }

  async function handlePostNow() {
    setPostingNow(true);
    setMessage(null);
    try {
      const updated = await postNowCandidate(candidate.id, { hook, take, post_text: postText });
      onUpdated(updated);
      setMessage("Post text copied — X compose + Finder opened. Cmd+V, drag clip, post.");
    } catch {
      setMessage("Post now failed.");
    } finally {
      setPostingNow(false);
    }
  }

  return (
    <article className="rounded-lg border border-line bg-ink-soft p-5">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <div className="flex flex-wrap items-center gap-2">
            <span className="font-mono text-xs text-signal">#{candidate.rank}</span>
            <span className={factoryStatusBadge(candidate.status)}>{candidate.status}</span>
            <span className="font-mono text-[10px] text-fog">
              {candidate.start_time} → {candidate.end_time}
            </span>
            <span className="font-mono text-[10px] text-fog">
              {Math.round(candidate.confidence * 100)}% conf
            </span>
          </div>
          <p className="mt-2 text-xs text-fog">{candidate.why_interesting}</p>
        </div>
      </div>

      <div className="mt-4 space-y-3">
        <div>
          <label className="font-mono text-[10px] uppercase tracking-wider text-fog">Hook</label>
          <input
            value={hook}
            onChange={(e) => setHook(e.target.value)}
            className="mt-1 w-full rounded-md border border-line bg-ink px-3 py-2 text-sm text-white focus:border-signal/50 focus:outline-none"
          />
        </div>
        <div>
          <label className="font-mono text-[10px] uppercase tracking-wider text-fog">Take</label>
          <textarea
            value={take}
            onChange={(e) => setTake(e.target.value)}
            rows={3}
            className="mt-1 w-full resize-y rounded-md border border-line bg-ink px-3 py-2 text-sm text-fog-light focus:border-signal/50 focus:outline-none"
          />
        </div>
        <div>
          <label className="font-mono text-[10px] uppercase tracking-wider text-fog">Post text</label>
          <textarea
            value={postText}
            onChange={(e) => setPostText(e.target.value)}
            rows={8}
            className="mt-1 w-full resize-y rounded-md border border-line bg-ink px-3 py-2 font-mono text-xs leading-relaxed text-fog-light focus:border-signal/50 focus:outline-none"
          />
        </div>

        <div className="rounded-md border border-line/80 bg-ink/50 p-3">
          <label className="font-mono text-[10px] uppercase tracking-wider text-signal">Refine</label>
          <p className="mt-1 text-[11px] text-fog">
            Add your angle — curiosity, skepticism, a question. Uses your biases above.
          </p>
          <div className="mt-3 flex flex-wrap gap-2">
            {REFINE_PRESETS.map((preset) => (
              <button
                key={preset.label}
                type="button"
                onClick={() => handleRefine(preset.instruction)}
                disabled={rewriting}
                className="rounded border border-line px-2.5 py-1 text-[11px] text-fog-light transition hover:border-signal/40 hover:text-white disabled:opacity-50"
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
              placeholder="Custom instruction — e.g. make it about trust in Anthropic for small teams"
              className="min-w-[240px] flex-1 rounded-md border border-line bg-ink px-3 py-2 text-xs text-white placeholder:text-fog/60 focus:border-signal/50 focus:outline-none"
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

      {candidate.clip_path && (
        <p className="mt-3 truncate font-mono text-[10px] text-emerald-400">{candidate.clip_path}</p>
      )}

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
        <button
          type="button"
          onClick={handleClip}
          disabled={clipping}
          className="inline-flex items-center gap-1.5 rounded border border-line px-3 py-1.5 text-xs text-white transition hover:border-signal/40 disabled:opacity-50"
        >
          <Scissors className="h-3 w-3" />
          {clipping ? "Clipping…" : candidate.clip_path ? "Re-clip" : "Clip video"}
        </button>

        <div className="flex flex-1 flex-wrap items-center justify-end gap-2">
          <input
            type="datetime-local"
            value={scheduleAt}
            onChange={(e) => setScheduleAt(e.target.value)}
            className="h-[30px] rounded-md border border-line bg-ink px-2 font-mono text-xs text-white focus:border-signal/50 focus:outline-none"
          />
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={handlePostNow}
              disabled={postingNow || clipping}
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
      </div>

      {message && (
        <p className="mt-2 flex items-center gap-1 font-mono text-[10px] text-fog">
          {candidate.clip_path && <Film className="h-3 w-3 text-emerald-400" />}
          {message}
        </p>
      )}
    </article>
  );
}