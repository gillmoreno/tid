import { useEffect, useState } from "react";
import { Plus } from "lucide-react";
import { createSource, fetchPodcasts } from "@/api/factory";
import type { PodcastOption, Source } from "@/types/factory";

interface IngestFormProps {
  onCreated: (source: Source) => void;
}

export function IngestForm({ onCreated }: IngestFormProps) {
  const [url, setUrl] = useState("");
  const [podcast, setPodcast] = useState("");
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
      .catch(() => setError("Could not load podcast list. Check mentions dictionary."));
  }, []);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!url.trim() || !podcast) return;

    setLoading(true);
    setError(null);
    try {
      const source = await createSource({
        youtube_url: url.trim(),
        podcast,
      });
      onCreated(source);
      setUrl("");
    } catch {
      setError("Failed to ingest source. Is the backend running?");
    } finally {
      setLoading(false);
    }
  }

  const selected = podcasts.find((p) => p.name === podcast);

  return (
    <form onSubmit={handleSubmit} className="rounded-lg border border-line bg-ink-soft p-5">
      <h3 className="font-display text-sm font-bold text-white">Ingest podcast</h3>
      <p className="mt-1 text-xs text-fog">
        Paste a YouTube URL and pick the podcast — posts attribute @{selected?.handle ?? "handle"}.
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
        <select
          value={podcast}
          onChange={(e) => setPodcast(e.target.value)}
          required
          className="w-full rounded-md border border-line bg-ink px-3 py-2 text-sm text-white focus:border-signal/50 focus:outline-none"
        >
          <option value="" disabled>
            Select podcast…
          </option>
          {podcasts.map((p) => (
            <option key={p.name} value={p.name}>
              {p.name} (@{p.handle})
            </option>
          ))}
        </select>
      </div>

      {error && <p className="mt-3 font-mono text-xs text-red-400">{error}</p>}

      <button
        type="submit"
        disabled={loading || !url.trim() || !podcast}
        className="mt-4 inline-flex items-center gap-2 rounded-md bg-signal px-4 py-2 text-sm font-medium text-ink transition hover:bg-signal-glow disabled:cursor-not-allowed disabled:opacity-50"
      >
        <Plus className="h-4 w-4" />
        {loading ? "Ingesting…" : "Add source"}
      </button>
    </form>
  );
}