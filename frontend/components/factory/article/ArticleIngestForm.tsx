import { useEffect, useState } from "react";
import { Plus } from "lucide-react";
import { Link } from "react-router-dom";
import { createArticleSource, fetchPublications } from "@/api/articles";
import type { ArticleSource, PublicationOption } from "@/types/factory";

interface ArticleIngestFormProps {
  onCreated: (source: ArticleSource) => void;
}

export function ArticleIngestForm({ onCreated }: ArticleIngestFormProps) {
  const [url, setUrl] = useState("");
  const [publication, setPublication] = useState("");
  const [publications, setPublications] = useState<PublicationOption[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchPublications()
      .then((items) => {
        setPublications(items);
        if (items.length > 0) {
          setPublication((prev) => prev || items[0].name);
        }
      })
      .catch(() => setError("Could not load publications. Add news feeds in Sources."));
  }, []);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!url.trim() || !publication) return;
    setLoading(true);
    setError(null);
    try {
      const source = await createArticleSource({ url: url.trim(), publication });
      onCreated(source);
      setUrl("");
    } catch {
      setError("Failed to ingest article. Is the backend running?");
    } finally {
      setLoading(false);
    }
  }

  const selected = publications.find((p) => p.name === publication);

  return (
    <form onSubmit={handleSubmit} className="rounded-lg border border-line bg-ink-soft p-5">
      <h3 className="font-display text-sm font-bold text-white">Ingest article</h3>
      <p className="mt-1 text-xs text-fog">
        Paste an article URL and pick the publication — posts attribute @{selected?.handle ?? "handle"}.
      </p>

      <div className="mt-4 space-y-3">
        <input
          type="url"
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          placeholder="https://www.theinformation.com/articles/..."
          className="w-full rounded-md border border-line bg-ink px-3 py-2 font-mono text-sm text-white placeholder:text-fog/60 focus:border-signal/50 focus:outline-none"
          required
        />
        {publications.length === 0 ? (
          <p className="text-xs text-fog">
            No publications yet.{" "}
            <Link to="/factory/sources" className="text-signal hover:underline">
              Add a news feed in Sources
            </Link>
            .
          </p>
        ) : (
          <select
            value={publication}
            onChange={(e) => setPublication(e.target.value)}
            required
            className="w-full rounded-md border border-line bg-ink px-3 py-2 text-sm text-white focus:border-signal/50 focus:outline-none"
          >
            <option value="" disabled>
              Select publication…
            </option>
            {publications.map((p) => (
              <option key={p.name} value={p.name}>
                {p.name}
              </option>
            ))}
          </select>
        )}
      </div>

      {error && <p className="mt-3 font-mono text-xs text-red-400">{error}</p>}

      <button
        type="submit"
        disabled={loading || !url.trim() || !publication}
        className="mt-4 inline-flex items-center gap-2 rounded-md bg-signal px-4 py-2 text-sm font-medium text-ink transition hover:bg-signal-glow disabled:cursor-not-allowed disabled:opacity-50"
      >
        <Plus className="h-4 w-4" />
        {loading ? "Ingesting…" : "Add article"}
      </button>
    </form>
  );
}
