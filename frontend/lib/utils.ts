import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function statusLabel(status: string): string {
  switch (status) {
    case "live":
      return "Live";
    case "testing":
      return "Testing";
    case "building":
      return "Building";
    case "idea":
      return "Idea";
    default:
      return status;
  }
}

export function statusColor(status: string): string {
  switch (status) {
    case "live":
      return "text-emerald-400 border-emerald-400/30 bg-emerald-400/10";
    case "testing":
      return "text-signal border-signal/30 bg-signal/10";
    case "building":
      return "text-sky-400 border-sky-400/30 bg-sky-400/10";
    default:
      return "text-fog border-line bg-ink-muted";
  }
}