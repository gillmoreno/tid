import { useEffect, useState } from "react";
import { ChevronDown, Save } from "lucide-react";
import { Link } from "react-router-dom";
import { fetchArticlePrompt, updateArticlePrompt } from "@/api/articles";
import { cn } from "@/lib/utils";

interface ArticleSettingsPanelProps {
  defaultOpen?: boolean;
}

export function ArticleSettingsPanel({ defaultOpen = false }: ArticleSettingsPanelProps) {
  const [open, setOpen] = useState(defaultOpen);
  const [prompt, setPrompt] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<string | null>(null);

  useEffect(() => {
    async function load() {
      try {
        const p = await fetchArticlePrompt();
        setPrompt(p.content);
      } catch {
        setMessage("Could not load the article prompt.");
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  async function savePrompt() {
    setSaving(true);
    setMessage(null);
    try {
      await updateArticlePrompt(prompt);
      setMessage("Article prompt saved.");
    } catch {
      setMessage("Failed to save prompt.");
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="rounded-lg border border-line bg-ink-soft">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="flex w-full items-center justify-between px-5 py-4 text-left"
      >
        <div>
          <h3 className="font-display text-sm font-bold text-white">Analysis prompt</h3>
          <p className="mt-0.5 text-xs text-fog">
            Instructions the analyzer follows per article. Biases + @ handles live in{" "}
            <Link to="/factory/sources" className="text-signal hover:underline">
              Sources
            </Link>
            .
          </p>
        </div>
        <ChevronDown className={cn("h-4 w-4 text-fog transition", open && "rotate-180")} />
      </button>

      {open && (
        <div className="border-t border-line px-5 pb-5 pt-4">
          {loading ? (
            <p className="py-4 text-sm text-fog">Loading…</p>
          ) : (
            <div>
              <textarea
                value={prompt}
                onChange={(e) => setPrompt(e.target.value)}
                rows={18}
                className="w-full resize-y rounded-md border border-line bg-ink px-3 py-2 font-mono text-xs leading-relaxed text-fog-light focus:border-signal/50 focus:outline-none"
              />
              <button
                type="button"
                onClick={savePrompt}
                disabled={saving}
                className="mt-2 inline-flex items-center gap-1.5 rounded border border-line px-3 py-1.5 text-xs text-white transition hover:border-signal/40 disabled:opacity-50"
              >
                <Save className="h-3 w-3" />
                {saving ? "Saving…" : "Save prompt"}
              </button>
            </div>
          )}
          {message && <p className="mt-3 font-mono text-xs text-fog">{message}</p>}
        </div>
      )}
    </div>
  );
}
