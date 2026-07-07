import { cn } from "@/lib/utils";

export function factoryStatusColor(status: string): string {
  switch (status) {
    case "analyzed":
    case "approved":
    case "prepared":
    case "clipped":
      return "text-emerald-400 border-emerald-400/30 bg-emerald-400/10";
    case "analyzing":
    case "clipping":
    case "scheduled":
    case "pending":
      return "text-signal border-signal/30 bg-signal/10";
    case "failed":
      return "text-red-400 border-red-400/30 bg-red-400/10";
    default:
      return "text-fog border-line bg-ink-muted";
  }
}

export function factoryStatusBadge(status: string): string {
  return cn(
    "inline-flex rounded border px-2 py-0.5 font-mono text-[10px] uppercase tracking-wider",
    factoryStatusColor(status)
  );
}

export function toRFC3339(localDatetime: string): string {
  if (!localDatetime) return "";
  return new Date(localDatetime).toISOString();
}

export function fromRFC3339(iso: string): string {
  if (!iso) return "";
  const d = new Date(iso);
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

export function formatScheduledAt(iso: string): string {
  return new Date(iso).toLocaleString(undefined, {
    weekday: "short",
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}

export function parseTimestamp(s: string): number {
  const parts = s.trim().split(":").map((p) => Number(p));
  if (parts.length === 3) return parts[0] * 3600 + parts[1] * 60 + parts[2];
  if (parts.length === 2) return parts[0] * 60 + parts[1];
  return 0;
}

export function formatClipDuration(start: string, end: string): string {
  const seconds = Math.max(0, Math.round(parseTimestamp(end) - parseTimestamp(start)));
  const m = Math.floor(seconds / 60);
  const s = seconds % 60;
  return `${m}:${String(s).padStart(2, "0")}`;
}

export function defaultScheduleTime(): string {
  const d = new Date();
  d.setHours(16, 0, 0, 0);
  if (d.getTime() <= Date.now()) {
    d.setDate(d.getDate() + 1);
  }
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}