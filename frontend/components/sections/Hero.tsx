import { Logo } from "@/components/brand/Logo";
import type { Site } from "@/types/site";

type HeroProps = {
  site: Site;
};

export function Hero({ site }: HeroProps) {
  return (
    <section className="mx-auto max-w-6xl px-6 pb-20 pt-16 md:pt-24">
      <div className="max-w-3xl">
        <div
          className="mb-8 animate-fade-up"
          style={{ animationDelay: "0ms" }}
        >
          <Logo
            size="hero"
            className="drop-shadow-[0_0_24px_rgba(0,255,65,0.35)]"
          />
        </div>

        <p
          className="mb-6 inline-flex items-center gap-2 font-mono text-xs uppercase tracking-[0.2em] text-signal animate-fade-up"
          style={{ animationDelay: "0ms" }}
        >
          <span className="h-1.5 w-1.5 rounded-full bg-signal animate-pulse-signal" />
          Signal vs noise
        </p>

        <h1
          className="font-display text-4xl font-extrabold leading-[1.05] tracking-tight text-white md:text-6xl animate-fade-up text-balance"
          style={{ animationDelay: "80ms", opacity: 0 }}
        >
          {site.name}
        </h1>

        <p
          className="mt-6 text-lg leading-relaxed text-fog-light md:text-xl animate-fade-up text-balance"
          style={{ animationDelay: "160ms", opacity: 0 }}
        >
          {site.tagline}
        </p>

        <p
          className="mt-4 max-w-2xl text-base leading-relaxed text-fog animate-fade-up"
          style={{ animationDelay: "240ms", opacity: 0 }}
        >
          {site.description}
        </p>
      </div>

      <div
        className="mt-12 flex flex-wrap gap-4 font-mono text-xs text-fog animate-fade-up"
        style={{ animationDelay: "320ms", opacity: 0 }}
      >
        <span className="rounded-full border border-line px-3 py-1">1 builder</span>
        <span className="rounded-full border border-line px-3 py-1">real loops</span>
        <span className="rounded-full border border-line px-3 py-1">no hype</span>
      </div>
    </section>
  );
}