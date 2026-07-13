import { useEffect, useState } from "react";
import { Calendar, Film, Save, Scissors, Send, Sparkles } from "lucide-react";
import {
  clipCandidate,
  fetchCandidateTranscript,
  postNowCandidate,
  rewriteCandidate,
  scheduleCandidate,
  trimCandidate,
  updateCandidate,
} from "@/api/factory";
import {
  defaultScheduleTime,
  factoryStatusBadge,
  formatClipDuration,
  postTranscriptMismatch,
  toRFC3339,
} from "@/components/factory/factory-utils";
import { TranscriptLines } from "@/components/factory/TranscriptLines";
import type { Candidate } from "@/types/factory";

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
      "End with a sharp curiosity-driving question — e.g. what does this mean, who can we trust, what's the alternative for teams that cannot spend hundreds of thousands on infra.",
  },
  {
    label: "Gil's take",
    instruction:
      "Format C commentary: open with Gil's skeptical-curious opinion — what he thinks this means, agrees with, or pushes back on. Anchor in one concrete thing the speaker said. Not a neutral summary or full quote. End with implication or sharp question.",
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
  {
    label: "Humanize",
    instruction:
      "Humanize this post. Remove the AI stylistic fingerprints so it reads like clean, natural human writing by a careful person. Strip boilerplate openers, inflated diction (leverage, robust, crucial, landscape, etc.), mechanical parallel structures and tricolons, uniform sentence length, em-dash overload, empty hedging, and restate-everything conclusions. Vary rhythm hard — mix long clauses with short sentences and fragments. Keep every real claim, number, name, and meaning exactly. Do not invent facts, quotes, or anecdotes. Flag gaps instead of filling them. End result should feel like the same content written by a slightly uneven, opinionated human, not a model.",
  },
] as const;

interface CandidateCardProps {
  candidate: Candidate;
  onUpdated: (candidate: Candidate) => void;
  onScheduled: () => void;
}

export function CandidateCard({ candidate, onUpdated, onScheduled }: CandidateCardProps) {
  const [postText, setPostText] = useState(candidate.post_text);
  const [scheduleAt, setScheduleAt] = useState(defaultScheduleTime());
  const [saving, setSaving] = useState(false);
  const [startTime, setStartTime] = useState(candidate.start_time);
  const [endTime, setEndTime] = useState(candidate.end_time);
  const [clipping, setClipping] = useState(false);
  const [trimming, setTrimming] = useState(false);
  const [scheduling, setScheduling] = useState(false);
  const [postingNow, setPostingNow] = useState(false);
  const [rewriting, setRewriting] = useState(false);
  const [refineInstruction, setRefineInstruction] = useState("");
  const [transcript, setTranscript] = useState("");
  const [transcriptLoading, setTranscriptLoading] = useState(false);
  const [transcriptAvailable, setTranscriptAvailable] = useState(true);
  const [message, setMessage] = useState<string | null>(null);

  useEffect(() => {
    setPostText(candidate.post_text);
    setStartTime(candidate.start_time);
    setEndTime(candidate.end_time);
  }, [candidate]);

  const timesChanged =
    startTime !== candidate.start_time || endTime !== candidate.end_time;
  const transcriptMismatch =
    !transcriptLoading && transcript.length > 0 && postTranscriptMismatch(postText, transcript);

  useEffect(() => {
    let cancelled = false;
    const timer = window.setTimeout(async () => {
      setTranscriptLoading(true);
      try {
        const result = await fetchCandidateTranscript(candidate.id, {
          start_time: startTime,
          end_time: endTime,
        });
        if (cancelled) return;
        setTranscript(result.text);
        setTranscriptAvailable(result.available);
      } catch {
        if (!cancelled) {
          setTranscript("");
          setTranscriptAvailable(false);
        }
      } finally {
        if (!cancelled) setTranscriptLoading(false);
      }
    }, 350);

    return () => {
      cancelled = true;
      window.clearTimeout(timer);
    };
  }, [candidate.id, startTime, endTime]);

  async function handleSave() {
    setSaving(true);
    setMessage(null);
    try {
      const updated = await updateCandidate(candidate.id, { post_text: postText });
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

  async function handleTrim() {
    setTrimming(true);
    setMessage(null);
    try {
      const updated = await trimCandidate(candidate.id, {
        start_time: startTime,
        end_time: endTime,
      });
      onUpdated(updated);
      setMessage(`Trimmed to ${formatClipDuration(updated.start_time, updated.end_time)}.`);
    } catch {
      setMessage("Trim failed.");
    } finally {
      setTrimming(false);
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
      const updated = await rewriteCandidate(candidate.id, trimmed, { post_text: postText });
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

  async function handlePostNow() {
    setPostingNow(true);
    setMessage(null);
    try {
      const updated = await postNowCandidate(candidate.id, { post_text: postText });
      onUpdated(updated);
      setMessage("Post text copied — X compose + Finder opened. Cmd+V, drag clip, post.");
    } catch {
      setMessage("Post now failed.");
    } finally {
      setPostingNow(false);
    }
  }

  return (
    <article className="rounded-lg border border-line bg-ink-soft p-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <div className="flex flex-wrap items-center gap-2">
            <span className="font-mono text-xs text-signal">#{candidate.rank}</span>
            <span className={factoryStatusBadge(candidate.status)}>{candidate.status}</span>
            <span className="font-mono text-xs text-fog">
              {formatClipDuration(startTime, endTime)} clip
            </span>
            <span className="font-mono text-xs text-fog">
              {Math.round(candidate.confidence * 100)}% conf
            </span>
          </div>
          <p className="mt-2 text-sm leading-relaxed text-fog">{candidate.why_interesting}</p>
        </div>
      </div>

      <div className="mt-3 flex flex-wrap items-end gap-2 rounded-md border border-line/80 bg-ink/50 p-3">
        <div>
          <label className="font-mono text-xs uppercase tracking-wider text-fog">Start</label>
          <input
            type="text"
            value={startTime}
            onChange={(e) => setStartTime(e.target.value)}
            placeholder="00:06:00"
            className="mt-1 block w-[8.5rem] rounded-md border border-line bg-ink px-2.5 py-2 font-mono text-sm text-white focus:border-signal/50 focus:outline-none"
          />
        </div>
        <div>
          <label className="font-mono text-xs uppercase tracking-wider text-fog">End</label>
          <input
            type="text"
            value={endTime}
            onChange={(e) => setEndTime(e.target.value)}
            placeholder="00:09:20"
            className="mt-1 block w-[8.5rem] rounded-md border border-line bg-ink px-2.5 py-2 font-mono text-sm text-white focus:border-signal/50 focus:outline-none"
          />
        </div>
        <p className="pb-2 font-mono text-xs text-fog">
          {formatClipDuration(startTime, endTime)} total
          {timesChanged && candidate.clip_path ? " · unsaved trim" : ""}
        </p>
      </div>

      <div className="mt-4 rounded-md border border-line/80 bg-ink/50 p-4">
        <label className="font-mono text-xs uppercase tracking-wider text-fog">Source transcript</label>
        <p className="mt-1 text-sm text-fog">
          What was actually said in this range — check relevance before editing or clipping.
        </p>
        {transcriptMismatch ? (
          <p className="mt-2 rounded border border-amber-400/30 bg-amber-400/10 px-3 py-2 text-sm text-amber-200">
            Post text may not match this time range — the analyzer likely mixed up timestamps. Adjust
            start/end to where the topic is actually discussed, or re-analyze the source.
          </p>
        ) : null}
        <div className="mt-3 max-h-64 overflow-y-auto rounded-md border border-line bg-ink px-4 py-3">
          {transcriptLoading ? (
            <p className="text-sm text-fog">Loading transcript…</p>
          ) : transcript ? (
            <TranscriptLines text={transcript} />
          ) : (
            <p className="text-sm text-fog">
              {transcriptAvailable
                ? "No transcript text in this range."
                : "Timed captions unavailable for this source."}
            </p>
          )}
        </div>
      </div>

      <div className="mt-4 space-y-3">
        <div>
          <label className="font-mono text-xs uppercase tracking-wider text-fog">Post text</label>
          <textarea
            value={postText}
            onChange={(e) => setPostText(e.target.value)}
            rows={12}
            className="mt-2 w-full resize-y rounded-md border border-line bg-ink px-4 py-3 text-base leading-7 text-fog-light focus:border-signal/50 focus:outline-none"
          />
        </div>

        <div className="rounded-md border border-line/80 bg-ink/50 p-3">
          <label className="font-mono text-xs uppercase tracking-wider text-signal">Refine</label>
          <p className="mt-1 text-sm text-fog">
            Add Gil's take — use "Gil's take" or type your opinion in the box below (e.g. "I think harness beats model size here because…"). Uses biases above.
          </p>
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
              placeholder="Your take — e.g. I think they're wrong about ROI; the real story is pricing power, not AI spend"
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

      {candidate.clip_path && (
        <p className="mt-2 truncate font-mono text-[10px] text-emerald-400">{candidate.clip_path}</p>
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
        {candidate.clip_path && timesChanged ? (
          <button
            type="button"
            onClick={handleTrim}
            disabled={trimming || clipping}
            className="inline-flex items-center gap-1.5 rounded border border-emerald-400/40 bg-emerald-400/10 px-3 py-1.5 text-xs text-emerald-400 transition hover:bg-emerald-400/20 disabled:opacity-50"
          >
            <Scissors className="h-3 w-3" />
            {trimming ? "Trimming…" : "Trim clip"}
          </button>
        ) : null}
        <button
          type="button"
          onClick={handleClip}
          disabled={clipping || trimming}
          className="inline-flex items-center gap-1.5 rounded border border-line px-3 py-1.5 text-xs text-white transition hover:border-signal/40 disabled:opacity-50"
        >
          <Scissors className="h-3 w-3" />
          {clipping ? "Clipping…" : candidate.clip_path ? "Re-clip from source" : "Clip video"}
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