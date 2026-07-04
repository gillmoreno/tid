import type { Item, Section, SectionDetail, Site } from "@/types/site";

export const site: Site = {
  name: "The Idea Guy",
  tagline: "Testing what's real in AI — one person, real loops, no hype.",
  domain: "the-idea-guy.com",
  description:
    "A personal lab for testing AI ideas — separating signal from hype, and showing what one builder can actually ship.",
  mission: "Cut through AI noise. Build small loops. Share what works.",
};

export const sections: Section[] = [
  {
    slug: "ai-guild",
    title: "AI Guild of Dev",
    subtitle: "aigil.dev",
    description:
      "A developer guild for people building with AI — community, tools, and experiments in the open.",
    icon: "guild",
    order: 1,
    external_url: "https://aigil.dev/",
  },
  {
    slug: "loops-with-taste",
    title: "Loops with Taste",
    subtitle: "Content automation, human-filtered",
    description:
      "AI loops that turn raw signal into posts and shorts — but only after my taste filter. The hook is the output, not the pipeline.",
    icon: "loop",
    order: 2,
  },
  {
    slug: "software-ideas",
    title: "Software Ideas",
    subtitle: "Things I'm building",
    description:
      "Software experiments and products — small bets on what AI makes possible for a solo builder today.",
    icon: "code",
    order: 3,
  },
];

export const items: Item[] = [
  {
    slug: "aigil",
    section_slug: "ai-guild",
    title: "AIGil",
    tagline: "Developer guild for the AI era",
    description:
      "Community and tooling for developers navigating AI — what's worth learning, what's worth building, and what's just noise.",
    status: "live",
    tags: ["community", "developers", "ai"],
    external_url: "https://aigil.dev/",
    featured: true,
  },
  {
    slug: "clip-to-post",
    section_slug: "loops-with-taste",
    title: "Clip → Post",
    tagline: "YouTube signal, X output",
    description:
      "Take a small slice of an AI YouTube clip that caught my attention, extract the insight, and rewrite it as tweets or X posts — filtered through my taste, not generic AI slop.",
    status: "testing",
    tags: ["x", "youtube", "content", "loop"],
    featured: true,
  },
  {
    slug: "grok-shorts",
    section_slug: "loops-with-taste",
    title: "Grok Imagine Shorts",
    tagline: "AI visuals, tight edits",
    description:
      "A loop for creating short-form video using Grok Imagine — fast visual generation paired with tight editing for reels and shorts.",
    status: "testing",
    tags: ["video", "shorts", "grok", "imagine"],
    featured: true,
  },
  {
    slug: "roger",
    section_slug: "software-ideas",
    title: "Roger",
    tagline: "Restaurant reservations, rebuilt",
    description:
      "Multi-restaurant reservation platform — Go API, React widget, the full stack. Proof that one person can ship production software with modern AI-assisted development.",
    status: "live",
    tags: ["go", "react", "saas", "hospitality"],
    featured: true,
  },
  {
    slug: "tid",
    section_slug: "software-ideas",
    title: "The Idea Guy",
    tagline: "This site",
    description:
      "The meta-project — a personal lab site to document and test AI ideas in public. Go backend, React frontend, Roger-style deploy.",
    status: "building",
    tags: ["go", "react", "brand", "lab"],
    featured: true,
  },
];

export function sectionDetail(slug: string): SectionDetail | null {
  const section = sections.find((s) => s.slug === slug);
  if (!section) return null;
  return {
    ...section,
    items: items.filter((i) => i.section_slug === slug),
  };
}

export function itemBySlug(slug: string): Item | null {
  return items.find((i) => i.slug === slug) ?? null;
}