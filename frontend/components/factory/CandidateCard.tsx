import { useEffect, useState } from "react";
import { Calendar, Film, Save, Scissors, Send } from "lucide-react";
import { clipCandidate, postNowCandidate, scheduleCandidate, updateCandidate } from "@/api/factory";
import {
  defaultScheduleTime,
  factoryStatusBadge,
  toRFC3339,
} from "@/components/factory/factory-utils";
import type { Candidate } from "@/types/factory";

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
            rows={6}
            className="mt-1 w-full resize-y rounded-md border border-line bg-ink px-3 py-2 font-mono text-xs leading-relaxed text-fog-light focus:border-signal/50 focus:outline-none"
          />
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