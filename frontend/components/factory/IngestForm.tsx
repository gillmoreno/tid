import { useState, useEffect } from "react";
import { Plus } from "lucide-react";
import { createSource } from "@/api/factory";
import type { Source } from "@/types/factory";

interface IngestFormProps {
  onCreated: (source: Source) => void;
}

export function IngestForm({ onCreated }: IngestFormProps) {
  const [url, setUrl] = useState("");
  const [title, setTitle] = useState("");
  const [podcast, setPodcast] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Auto-fill title and podcast from YouTube oEmbed when URL is pasted
  useEffect(() => {
    const trimmed = url.trim();
    if (!trimmed || (title && podcast)) return;
    if (!trimmed.includes("youtube.com") && !trimmed.includes("youtu.be")) return;

    const timer = setTimeout(async () => {
      try {
        const oembedURL = `https://www.youtube.com/oembed?url=${encodeURIComponent(trimmed)}&format=json`;
        const res = await fetch(oembedURL);
        if (res.ok) {
          const data = await res.json();
          if (!title && data.title) setTitle(data.title);
          if (!podcast && data.author_name) setPodcast(data.author_name);
        }
      } catch {
        // ignore fetch errors; backend will still try
      }
    }, 500);

    return () => clearTimeout(timer);
  }, [url, title, podcast]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!url.trim()) return;

    setLoading(true);
    setError(null);
    try {
      const source = await createSource({
        youtube_url: url.trim(),
        title: title.trim(),
        podcast: podcast.trim(),
      });
      onCreated(source);
      setUrl("");
      setTitle("");
      setPodcast("");
    } catch {
      setError("Failed to ingest source. Is the backend running?");
    } finally {
      setLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="rounded-lg border border-line bg-ink-soft p-5">
      <h3 className="font-display text-sm font-bold text-white">Ingest podcast</h3>
      <p className="mt-1 text-xs text-fog">Paste a YouTube URL. Title and podcast name are auto-filled from the video.</p>

      <div className="mt-4 space-y-3">
        <input
          type="url"
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          placeholder="https://youtube.com/watch?v=..."
          className="w-full rounded-md border border-line bg-ink px-3 py-2 font-mono text-sm text-white placeholder:text-fog/60 focus:border-signal/50 focus:outline-none"
          required
        />
        <div className="grid gap-3 sm:grid-cols-2">
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Episode title (auto-filled from URL)"
            className="rounded-md border border-line bg-ink px-3 py-2 text-sm text-white placeholder:text-fog/60 focus:border-signal/50 focus:outline-none"
          />
          <input
            type="text"
            value={podcast}
            onChange={(e) => setPodcast(e.target.value)}
            placeholder="Podcast name (auto-filled from URL)"
            className="rounded-md border border-line bg-ink px-3 py-2 text-sm text-white placeholder:text-fog/60 focus:border-signal/50 focus:outline-none"
          />
        </div>
      </div>

      {error && <p className="mt-3 font-mono text-xs text-red-400">{error}</p>}

      <button
        type="submit"
        disabled={loading || !url.trim()}
        className="mt-4 inline-flex items-center gap-2 rounded-md bg-signal px-4 py-2 text-sm font-medium text-ink transition hover:bg-signal-glow disabled:cursor-not-allowed disabled:opacity-50"
      >
        <Plus className="h-4 w-4" />
        {loading ? "Ingesting…" : "Add source"}
      </button>
    </form>
  );
}