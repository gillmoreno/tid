export function Footer() {
  return (
    <footer className="relative z-10 border-t border-line/60">
      <div className="mx-auto flex max-w-6xl flex-col gap-2 px-6 py-10 text-sm text-fog sm:flex-row sm:items-center sm:justify-between">
        <p>
          <span className="font-display font-semibold text-white">The Idea Guy</span>
          <span className="mx-2 text-line">·</span>
          Testing what&apos;s real in AI
        </p>
        <p className="font-mono text-xs text-fog/70">the-idea-guy.com</p>
      </div>
    </footer>
  );
}