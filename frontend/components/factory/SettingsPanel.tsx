import { useEffect, useState } from "react";
import { ChevronDown, Save } from "lucide-react";
import { fetchBiases, fetchPrompt, updateBiases, updatePrompt } from "@/api/factory";
import { cn } from "@/lib/utils";

export function SettingsPanel() {
  const [open, setOpen] = useState(false);
  const [biases, setBiases] = useState("");
  const [prompt, setPrompt] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState<"biases" | "prompt" | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  useEffect(() => {
    async function load() {
      try {
        const [b, p] = await Promise.all([fetchBiases(), fetchPrompt()]);
        setBiases(b.content);
        setPrompt(p.content);
      } catch {
        setMessage("Could not load settings.");
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  async function saveBiases() {
    setSaving("biases");
    setMessage(null);
    try {
      await updateBiases(biases);
      setMessage("Biases saved.");
    } catch {
      setMessage("Failed to save biases.");
    } finally {
      setSaving(null);
    }
  }

  async function savePrompt() {
    setSaving("prompt");
    setMessage(null);
    try {
      await updatePrompt(prompt);
      setMessage("Prompt saved.");
    } catch {
      setMessage("Failed to save prompt.");
    } finally {
      setSaving(null);
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
          <h3 className="font-display text-sm font-bold text-white">Lens & instructions</h3>
          <p className="mt-0.5 text-xs text-fog">Biases and analysis prompt — evolve over time.</p>
        </div>
        <ChevronDown className={cn("h-4 w-4 text-fog transition", open && "rotate-180")} />
      </button>

      {open && (
        <div className="border-t border-line px-5 pb-5">
          {loading ? (
            <p className="py-4 text-sm text-fog">Loading…</p>
          ) : (
            <div className="mt-4 grid gap-5 lg:grid-cols-2">
              <div>
                <label className="font-mono text-[10px] uppercase tracking-wider text-signal">
                  Biases
                </label>
                <textarea
                  value={biases}
                  onChange={(e) => setBiases(e.target.value)}
                  rows={14}
                  className="mt-2 w-full resize-y rounded-md border border-line bg-ink px-3 py-2 font-mono text-xs leading-relaxed text-fog-light focus:border-signal/50 focus:outline-none"
                />
                <button
                  type="button"
                  onClick={saveBiases}
                  disabled={saving === "biases"}
                  className="mt-2 inline-flex items-center gap-1.5 rounded border border-line px-3 py-1.5 text-xs text-white transition hover:border-signal/40 disabled:opacity-50"
                >
                  <Save className="h-3 w-3" />
                  {saving === "biases" ? "Saving…" : "Save biases"}
                </button>
              </div>
              <div>
                <label className="font-mono text-[10px] uppercase tracking-wider text-signal">
                  Analysis prompt
                </label>
                <textarea
                  value={prompt}
                  onChange={(e) => setPrompt(e.target.value)}
                  rows={14}
                  className="mt-2 w-full resize-y rounded-md border border-line bg-ink px-3 py-2 font-mono text-xs leading-relaxed text-fog-light focus:border-signal/50 focus:outline-none"
                />
                <button
                  type="button"
                  onClick={savePrompt}
                  disabled={saving === "prompt"}
                  className="mt-2 inline-flex items-center gap-1.5 rounded border border-line px-3 py-1.5 text-xs text-white transition hover:border-signal/40 disabled:opacity-50"
                >
                  <Save className="h-3 w-3" />
                  {saving === "prompt" ? "Saving…" : "Save prompt"}
                </button>
              </div>
            </div>
          )}
          {message && <p className="mt-3 font-mono text-xs text-fog">{message}</p>}
        </div>
      )}
    </div>
  );
}