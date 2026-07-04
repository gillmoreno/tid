import { apiClient } from "@/api/client";
import type { Item, Section, SectionDetail, Site } from "@/types/site";

export async function fetchSite(): Promise<Site> {
  const { data } = await apiClient.get<Site>("/site");
  return data;
}

export async function fetchSections(): Promise<Section[]> {
  const { data } = await apiClient.get<Section[]>("/sections");
  return data;
}

export async function fetchSection(slug: string): Promise<SectionDetail> {
  const { data } = await apiClient.get<SectionDetail>(`/sections/${slug}`);
  return data;
}

export async function fetchItem(slug: string): Promise<Item> {
  const { data } = await apiClient.get<Item>(`/items/${slug}`);
  return data;
}