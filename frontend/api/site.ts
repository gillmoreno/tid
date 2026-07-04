import { apiClient } from "@/api/client";
import * as seed from "@/content/seed";
import type { Item, Section, SectionDetail, Site } from "@/types/site";

const useStatic = import.meta.env.VITE_STATIC_CONTENT === "true";

export async function fetchSite(): Promise<Site> {
  if (useStatic) return seed.site;
  const { data } = await apiClient.get<Site>("/site");
  return data;
}

export async function fetchSections(): Promise<Section[]> {
  if (useStatic) return seed.sections;
  const { data } = await apiClient.get<Section[]>("/sections");
  return data;
}

export async function fetchSection(slug: string): Promise<SectionDetail> {
  if (useStatic) {
    const detail = seed.sectionDetail(slug);
    if (!detail) throw new Error("section not found");
    return detail;
  }
  const { data } = await apiClient.get<SectionDetail>(`/sections/${slug}`);
  return data;
}

export async function fetchItem(slug: string): Promise<Item> {
  if (useStatic) {
    const item = seed.itemBySlug(slug);
    if (!item) throw new Error("item not found");
    return item;
  }
  const { data } = await apiClient.get<Item>(`/items/${slug}`);
  return data;
}

export async function fetchFeaturedItems(): Promise<Item[]> {
  if (useStatic) return seed.items.filter((i) => i.featured);
  const sectionsData = await fetchSections();
  const sectionDetails = await Promise.all(
    sectionsData.map((s) => apiClient.get<{ items: Item[] }>(`/sections/${s.slug}`))
  );
  return sectionDetails.flatMap((r) => r.data.items).filter((i) => i.featured);
}