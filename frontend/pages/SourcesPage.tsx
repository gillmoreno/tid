import { useCallback, useEffect, useMemo, useState } from "react";
import { ChevronDown, Database, Plus, Save, Trash2 } from "lucide-react";
import { fetchDictionary, parseDictionary, saveDictionary } from "@/api/articles";
import { FactoryNav } from "@/components/factory/FactoryNav";
import { cn } from "@/lib/utils";
import type { Dictionary, DictionaryCategory, MentionEntry } from "@/types/factory";

interface EditEntry {
  name: string;
  handle: string;
  aliases: string;
  url: string;
}

type EditDict = Record<DictionaryCategory, EditEntry[]>;

const CATEGORIES: {
  key: DictionaryCategory;
  label: string;
  description: string;
  showUrl: boolean;
}[] = [
  {
    key: "podcasts",
    label: "Podcasts",
    description: "Podcast sources for the clip factory. Selected on ingest and used for @ attribution.",
    showUrl: false,
  },
  {
    key: "news_feeds",
    label: "News feeds",
    description: "Publication sources for the Articles factory. Selected on ingest and used for @ attribution.",
    showUrl: true,
  },
  {
    key: "companies",
    label: "Companies",
    description: "Company @ handles — tagged in posts when the company is mentioned.",
    showUrl: true,
  },
  {
    key: "people",
    label: "People",
    description: "People @ handles — tagged in posts when the person is mentioned.",
    showUrl: false,
  },
];

function emptyEntry(): EditEntry {
  return { name: "", handle: "", aliases: "", url: "" };
}

function toEdit(dict: Dictionary): EditDict {
  const map = (entries: MentionEntry[]): EditEntry[] =>
    entries.map((e) => ({
      name: e.name ?? "",
      handle: e.handle ?? "",
      aliases: (e.aliases ?? []).join(", "),
      url: e.url ?? "",
    }));
  return {
    people: map(dict.people),
    companies: map(dict.companies),
    podcasts: map(dict.podcasts),
    news_feeds: map(dict.news_feeds),
  };
}

function toDictionary(edit: EditDict): Dictionary {
  const map = (entries: EditEntry[]): MentionEntry[] =>
    entries
      .filter((e) => e.name.trim() || e.handle.trim())
      .map((e) => {
        const aliases = e.aliases
          .split(",")
          .map((a) => a.trim())
          .filter(Boolean);
        const entry: MentionEntry = {
          name: e.name.trim(),
          handle: e.handle.trim().replace(/^@/, ""),
        };
        if (aliases.length > 0) entry.aliases = aliases;
        if (e.url.trim()) entry.url = e.url.trim();
        return entry;
      });
  return {
    people: map(edit.people),
    companies: map(edit.companies),
    podcasts: map(edit.podcasts),
    news_feeds: map(edit.news_feeds),
  };
}

export function SourcesPage() {
  const [edit, setEdit] = useState<EditDict | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [dirty, setDirty] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [rawOpen, setRawOpen] = useState(false);
  const [rawJson, setRawJson] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const { dictionary } = await fetchDictionary();
      setEdit(toEdit(dictionary));
      setRawJson(JSON.stringify(dictionary, null, 2));
      setDirty(false);
    } catch {
      setMessage("Could not load the sources dictionary. Is the backend running?");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  function mutate(next: EditDict) {
    setEdit(next);
    setDirty(true);
    setMessage(null);
  }

  function updateCell(cat: DictionaryCategory, index: number, field: keyof EditEntry, value: string) {
    if (!edit) return;
    const rows = edit[cat].slice();
    rows[index] = { ...rows[index], [field]: value };
    mutate({ ...edit, [cat]: rows });
  }

  function addRow(cat: DictionaryCategory) {
    if (!edit) return;
    mutate({ ...edit, [cat]: [...edit[cat], emptyEntry()] });
  }

  function removeRow(cat: DictionaryCategory, index: number) {
    if (!edit) return;
    mutate({ ...edit, [cat]: edit[cat].filter((_, i) => i !== index) });
  }

  const currentDictionary = useMemo(() => (edit ? toDictionary(edit) : null), [edit]);

  async function handleSave() {
    if (!currentDictionary) return;
    setSaving(true);
    setMessage(null);
    try {
      await saveDictionary(currentDictionary);
      setRawJson(JSON.stringify(currentDictionary, null, 2));
      setDirty(false);
      setMessage("Sources dictionary saved.");
    } catch {
      setMessage("Save failed.");
    } finally {
      setSaving(false);
    }
  }

  async function handleSaveRaw() {
    setSaving(true);
    setMessage(null);
    try {
      const dictionary = parseDictionary(rawJson);
      JSON.parse(rawJson); // surface invalid JSON
      await saveDictionary(dictionary);
      setEdit(toEdit(dictionary));
      setDirty(false);
      setMessage("Saved from raw JSON.");
    } catch {
      setMessage("Failed to save — raw content must be valid JSON.");
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="w-full px-6 py-10 pb-32 xl:px-10">
      <div className="mb-8 flex items-start justify-between gap-4 border-b border-line pb-6">
        <div>
          <div className="flex items-center gap-2">
            <Database className="h-5 w-5 text-signal" />
            <h1 className="font-display text-2xl font-bold text-white">Sources</h1>
          </div>
          <p className="mt-2 text-sm text-fog">
            Your dictionary of podcasts, news feeds, companies, and people — the sources you pull from and the @
            handles posts tag. Shared across every factory.
          </p>
        </div>
      </div>

      <FactoryNav />

      {loading ? (
        <div className="h-40 animate-pulse rounded-lg bg-ink-muted" />
      ) : edit ? (
        <div className="space-y-6">
          {CATEGORIES.map((cat) => (
            <DictionaryTable
              key={cat.key}
              label={cat.label}
              description={cat.description}
              showUrl={cat.showUrl}
              rows={edit[cat.key]}
              onChange={(index, field, value) => updateCell(cat.key, index, field, value)}
              onAdd={() => addRow(cat.key)}
              onRemove={(index) => removeRow(cat.key, index)}
            />
          ))}

          <div className="rounded-lg border border-line bg-ink-soft">
            <button
              type="button"
              onClick={() => {
                if (!rawOpen && currentDictionary) {
                  setRawJson(JSON.stringify(currentDictionary, null, 2));
                }
                setRawOpen((v) => !v);
              }}
              className="flex w-full items-center justify-between px-5 py-4 text-left"
            >
              <div>
                <h3 className="font-display text-sm font-bold text-white">Raw JSON</h3>
                <p className="mt-0.5 text-xs text-fog">Advanced — edit the whole dictionary as JSON.</p>
              </div>
              <ChevronDown className={cn("h-4 w-4 text-fog transition", rawOpen && "rotate-180")} />
            </button>
            {rawOpen && (
              <div className="border-t border-line px-5 pb-5 pt-4">
                <textarea
                  value={rawJson}
                  onChange={(e) => setRawJson(e.target.value)}
                  rows={18}
                  spellCheck={false}
                  className="w-full resize-y rounded-md border border-line bg-ink px-3 py-2 font-mono text-xs leading-relaxed text-fog-light focus:border-signal/50 focus:outline-none"
                />
                <button
                  type="button"
                  onClick={handleSaveRaw}
                  disabled={saving}
                  className="mt-2 inline-flex items-center gap-1.5 rounded border border-line px-3 py-1.5 text-xs text-white transition hover:border-signal/40 disabled:opacity-50"
                >
                  <Save className="h-3 w-3" />
                  {saving ? "Saving…" : "Save from JSON"}
                </button>
              </div>
            )}
          </div>
        </div>
      ) : (
        <p className="font-mono text-sm text-signal">{message}</p>
      )}

      {edit && (
        <div className="fixed inset-x-0 bottom-0 z-10 border-t border-line bg-ink/95 backdrop-blur">
          <div className="mx-auto flex max-w-full items-center justify-between gap-4 px-6 py-3 xl:px-10">
            <p className="font-mono text-xs text-fog">
              {message ?? (dirty ? "Unsaved changes" : "All changes saved")}
            </p>
            <button
              type="button"
              onClick={handleSave}
              disabled={saving || !dirty}
              className="inline-flex items-center gap-2 rounded-md bg-signal px-4 py-2 text-sm font-medium text-ink transition hover:bg-signal-glow disabled:cursor-not-allowed disabled:opacity-50"
            >
              <Save className="h-4 w-4" />
              {saving ? "Saving…" : "Save changes"}
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

interface DictionaryTableProps {
  label: string;
  description: string;
  showUrl: boolean;
  rows: EditEntry[];
  onChange: (index: number, field: keyof EditEntry, value: string) => void;
  onAdd: () => void;
  onRemove: (index: number) => void;
}

const inputClass =
  "w-full rounded border border-line bg-ink px-2.5 py-1.5 text-sm text-white placeholder:text-fog/50 focus:border-signal/50 focus:outline-none";

function DictionaryTable({ label, description, showUrl, rows, onChange, onAdd, onRemove }: DictionaryTableProps) {
  return (
    <section className="rounded-lg border border-line bg-ink-soft">
      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-line px-5 py-3">
        <div>
          <h3 className="font-display text-sm font-bold text-white">
            {label} <span className="font-mono text-xs text-fog">· {rows.length}</span>
          </h3>
          <p className="mt-0.5 text-xs text-fog">{description}</p>
        </div>
        <button
          type="button"
          onClick={onAdd}
          className="inline-flex items-center gap-1.5 rounded border border-line px-3 py-1.5 text-xs text-white transition hover:border-signal/40"
        >
          <Plus className="h-3 w-3" />
          Add
        </button>
      </div>

      {rows.length === 0 ? (
        <p className="px-5 py-6 text-sm text-fog">Nothing here yet. Add a row to get started.</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full border-collapse text-left">
            <thead>
              <tr className="border-b border-line text-[10px] uppercase tracking-wider text-fog">
                <th className="px-5 py-2 font-mono font-normal">Name</th>
                <th className="px-3 py-2 font-mono font-normal">@ Handle</th>
                <th className="px-3 py-2 font-mono font-normal">Aliases (comma-separated)</th>
                {showUrl && <th className="px-3 py-2 font-mono font-normal">URL</th>}
                <th className="px-3 py-2" />
              </tr>
            </thead>
            <tbody>
              {rows.map((row, index) => (
                <tr key={index} className="border-b border-line/60 last:border-b-0">
                  <td className="px-5 py-2 align-top">
                    <input
                      value={row.name}
                      onChange={(e) => onChange(index, "name", e.target.value)}
                      placeholder="Display name"
                      className={inputClass}
                    />
                  </td>
                  <td className="px-3 py-2 align-top">
                    <input
                      value={row.handle}
                      onChange={(e) => onChange(index, "handle", e.target.value)}
                      placeholder="handle"
                      className={cn(inputClass, "font-mono")}
                    />
                  </td>
                  <td className="px-3 py-2 align-top">
                    <input
                      value={row.aliases}
                      onChange={(e) => onChange(index, "aliases", e.target.value)}
                      placeholder="alias one, alias two"
                      className={inputClass}
                    />
                  </td>
                  {showUrl && (
                    <td className="px-3 py-2 align-top">
                      <input
                        value={row.url}
                        onChange={(e) => onChange(index, "url", e.target.value)}
                        placeholder="https://…"
                        className={cn(inputClass, "font-mono")}
                      />
                    </td>
                  )}
                  <td className="px-3 py-2 align-top text-right">
                    <button
                      type="button"
                      onClick={() => onRemove(index)}
                      className="rounded border border-red-400/30 p-1.5 text-red-400 transition hover:bg-red-400/10"
                      title="Remove row"
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}
