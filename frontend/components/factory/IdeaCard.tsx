import { useEffect, useState } from "react";
import { Save, Trash2 } from "lucide-react";
import { deleteIdea, updateIdea } from "@/api/factory";
import { factoryStatusBadge } from "@/components/factory/factory-utils";
import type { Idea, IdeaKind, IdeaStatus } from "@/types/factory";

const KINDS: IdeaKind[] = ["glossary", "essay", "x_article", "thread"];
const STATUSES: IdeaStatus[] = ["idea", "drafting", "ready", "published"];

interface IdeaCardProps {
  idea: Idea;
  onUpdated: (idea: Idea) => void;
  onDeleted: (id: string) => void;
}

export function IdeaCard({ idea, onUpdated, onDeleted }: IdeaCardProps) {
  const [title, setTitle] = useState(idea.title);
  const [kind, setKind] = useState(idea.kind);
  const [status, setStatus] = useState(idea.status);
  const [summary, setSummary] = useState(idea.summary);
  const [body, setBody] = useState(idea.body);
  const [xPost, setXPost] = useState(idea.x_post);
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [message, setMessage] = useState<string | null>(null);

  useEffect(() => {
    setTitle(idea.title);
    setKind(idea.kind);
    setStatus(idea.status);
    setSummary(idea.summary);
    setBody(idea.body);
    setXPost(idea.x_post);
  }, [idea]);

  async function handleSave() {
    setSaving(true);
    setMessage(null);
    try {
      const updated = await updateIdea(idea.id, {
        title,
        kind,
        status,
        summary,
        body,
        x_post: xPost,
      });
      onUpdated(updated);
      setMessage("Saved.");
    } catch {
      setMessage("Save failed.");
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete() {
    if (!window.confirm(`Remove idea "${idea.title}"?`)) return;
    setDeleting(true);
    try {
      await deleteIdea(idea.id);
      onDeleted(idea.id);
    } catch {
      setMessage("Delete failed.");
      setDeleting(false);
    }
  }

  return (
    <article className="rounded-lg border border-line bg-ink-soft p-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0 flex-1">
          <input
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            className="w-full rounded-md border border-line bg-ink px-3 py-2 text-lg font-semibold text-white focus:border-signal/50 focus:outline-none"
          />
          <p className="mt-1 font-mono text-xs text-fog">/{idea.slug}</p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <span className={factoryStatusBadge(status)}>{status}</span>
          <select
            value={kind}
            onChange={(e) => setKind(e.target.value as IdeaKind)}
            className="rounded-md border border-line bg-ink px-2 py-1.5 text-xs text-white focus:border-signal/50 focus:outline-none"
          >
            {KINDS.map((k) => (
              <option key={k} value={k}>
                {k}
              </option>
            ))}
          </select>
          <select
            value={status}
            onChange={(e) => setStatus(e.target.value as IdeaStatus)}
            className="rounded-md border border-line bg-ink px-2 py-1.5 text-xs text-white focus:border-signal/50 focus:outline-none"
          >
            {STATUSES.map((s) => (
              <option key={s} value={s}>
                {s}
              </option>
            ))}
          </select>
        </div>
      </div>

      <div className="mt-4 space-y-4">
        <div>
          <label className="font-mono text-xs uppercase tracking-wider text-fog">Idea summary</label>
          <textarea
            value={summary}
            onChange={(e) => setSummary(e.target.value)}
            rows={3}
            placeholder="What is this piece about? Who is it for?"
            className="mt-2 w-full resize-y rounded-md border border-line bg-ink px-4 py-3 text-base leading-7 text-fog-light focus:border-signal/50 focus:outline-none"
          />
        </div>

        <div>
          <label className="font-mono text-xs uppercase tracking-wider text-fog">
            Article / blog body
          </label>
          <p className="mt-1 text-sm text-fog">
            Long-form draft — can become a blog post and X Article from the same source.
          </p>
          <textarea
            value={body}
            onChange={(e) => setBody(e.target.value)}
            rows={14}
            placeholder="Markdown-friendly draft goes here..."
            className="mt-2 w-full resize-y rounded-md border border-line bg-ink px-4 py-3 font-mono text-sm leading-7 text-fog-light focus:border-signal/50 focus:outline-none"
          />
        </div>

        <div>
          <label className="font-mono text-xs uppercase tracking-wider text-fog">X post teaser</label>
          <textarea
            value={xPost}
            onChange={(e) => setXPost(e.target.value)}
            rows={5}
            placeholder="Short post to promote or excerpt the piece on X..."
            className="mt-2 w-full resize-y rounded-md border border-line bg-ink px-4 py-3 text-base leading-7 text-fog-light focus:border-signal/50 focus:outline-none"
          />
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
          {saving ? "Saving…" : "Save"}
        </button>
        <button
          type="button"
          onClick={handleDelete}
          disabled={deleting}
          className="inline-flex items-center gap-1.5 rounded border border-red-400/30 px-3 py-1.5 text-xs text-red-400 transition hover:bg-red-400/10 disabled:opacity-50"
        >
          <Trash2 className="h-3 w-3" />
          {deleting ? "Removing…" : "Remove"}
        </button>
        {message && <p className="font-mono text-xs text-fog">{message}</p>}
      </div>
    </article>
  );
}