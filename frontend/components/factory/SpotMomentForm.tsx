import { useEffect, useState } from "react";
import { Sparkles } from "lucide-react";
import { analyzeMoment, fetchPodcasts } from "@/api/factory";
import type { AnalyzeResult, PodcastOption } from "@/types/factory";

interface SpotMomentFormProps {
  onAnalyzed: (result: AnalyzeResult) => void;
}

export function SpotMomentForm({ onAnalyzed }: SpotMomentFormProps) {
  const [url, setUrl] = useState("");
  const [podcast, setPodcast] = useState("");
  const [startTime, setStartTime] = useState("");
  const [endTime, setEndTime] = useState("");
  const [focusNote, setFocusNote] = useState("");
  const [podcasts, setPodcasts] = useState<PodcastOption[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchPodcasts()
      .then((items) => {
        setPodcasts(items);
        if (items.length > 0) {
          setPodcast((prev) => prev || items[0].name);
        }
      })
      .catch(() => setError("Could not load podcast list."));
  }, []);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!url.trim() || !podcast || !startTime.trim() || !endTime.trim()) return;

    setLoading(true);
    setError(null);
    try {
      const result = await analyzeMoment({
        youtube_url: url.trim(),
        podcast,
        start_time: startTime.trim(),
        end_time: endTime.trim(),
        focus_note: focusNote.trim(),
      });
      onAnalyzed(result);
      setFocusNote("");
    } catch {
      setError("Spot analysis failed. Check URL, from/to timestamps, and that grok is available.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="rounded-lg border border-signal/30 bg-signal/5 p-5">
      <h3 className="font-display text-sm font-bold text-white">Spot a moment</h3>
      <p className="mt-1 text-sm text-fog">
        Set the exact clip range (from → to) — that becomes the video for all 3 takes. Optional
        your take; posts vary in length (short, medium, long), not the clip.
      </p>

      <div className="mt-4 space-y-3">
        <input
          type="url"
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          placeholder="https://youtube.com/watch?v=..."
          className="w-full rounded-md border border-line bg-ink px-3 py-2 font-mono text-sm text-white placeholder:text-fog/60 focus:border-signal/50 focus:outline-none"
          required
        />
        <div className="grid gap-3 sm:grid-cols-3">
          <input
            type="text"
            value={startTime}
            onChange={(e) => setStartTime(e.target.value)}
            placeholder="From — 42:10"
            className="rounded-md border border-line bg-ink px-3 py-2 font-mono text-sm text-white placeholder:text-fog/60 focus:border-signal/50 focus:outline-none"
            required
          />
          <input
            type="text"
            value={endTime}
            onChange={(e) => setEndTime(e.target.value)}
            placeholder="To — 47:30"
            className="rounded-md border border-line bg-ink px-3 py-2 font-mono text-sm text-white placeholder:text-fog/60 focus:border-signal/50 focus:outline-none"
            required
          />
          <select
            value={podcast}
            onChange={(e) => setPodcast(e.target.value)}
            required
            className="rounded-md border border-line bg-ink px-3 py-2 text-sm text-white focus:border-signal/50 focus:outline-none"
          >
            <option value="" disabled>
              Select podcast…
            </option>
            {podcasts.map((p) => (
              <option key={p.name} value={p.name}>
                {p.name}
              </option>
            ))}
          </select>
        </div>
        <textarea
          value={focusNote}
          onChange={(e) => setFocusNote(e.target.value)}
          rows={3}
          placeholder="Your take (optional) — e.g. I think harness matters more than model size…"
          className="w-full resize-y rounded-md border border-line bg-ink px-3 py-2 text-sm leading-relaxed text-white placeholder:text-fog/60 focus:border-signal/50 focus:outline-none"
        />
      </div>

      {error && <p className="mt-3 font-mono text-xs text-red-400">{error}</p>}

      <button
        type="submit"
        disabled={loading || !url.trim() || !podcast || !startTime.trim() || !endTime.trim()}
        className="mt-4 inline-flex items-center gap-2 rounded-md bg-signal px-4 py-2 text-sm font-medium text-ink transition hover:bg-signal-glow disabled:cursor-not-allowed disabled:opacity-50"
      >
        <Sparkles className="h-4 w-4" />
        {loading ? "Analyzing moment…" : "Get 3 takes"}
      </button>
    </form>
  );
}