export type Site = {
  name: string;
  tagline: string;
  domain: string;
  description: string;
  mission: string;
};

export type Section = {
  slug: string;
  title: string;
  subtitle: string;
  description: string;
  icon: string;
  order: number;
  external_url?: string;
};

export type Item = {
  slug: string;
  section_slug: string;
  title: string;
  tagline: string;
  description: string;
  status: "live" | "testing" | "building" | "idea";
  tags: string[];
  external_url?: string;
  featured: boolean;
};

export type SectionDetail = Section & {
  items: Item[];
};