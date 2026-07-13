import { apiClient } from "@/api/client";
import type {
  ArticleAnalyzeResult,
  ArticleCandidate,
  ArticleSource,
  Dictionary,
  MentionDictionaryProfile,
  PromptTemplate,
  PublicationOption,
} from "@/types/factory";

const EMPTY_DICTIONARY: Dictionary = {
  people: [],
  companies: [],
  podcasts: [],
  news_feeds: [],
};

export function parseDictionary(content: string): Dictionary {
  try {
    const parsed = JSON.parse(content) as Partial<Dictionary>;
    return {
      people: parsed.people ?? [],
      companies: parsed.companies ?? [],
      podcasts: parsed.podcasts ?? [],
      news_feeds: parsed.news_feeds ?? [],
    };
  } catch {
    return { ...EMPTY_DICTIONARY };
  }
}

export async function fetchDictionary(): Promise<{ profile: MentionDictionaryProfile; dictionary: Dictionary }> {
  const { data } = await apiClient.get<MentionDictionaryProfile>("/factory/mentions");
  return { profile: data, dictionary: parseDictionary(data.content) };
}

export async function saveDictionary(dictionary: Dictionary): Promise<MentionDictionaryProfile> {
  const content = JSON.stringify(dictionary, null, 2);
  const { data } = await apiClient.put<MentionDictionaryProfile>("/factory/mentions", { content });
  return data;
}

export async function fetchPublications(): Promise<PublicationOption[]> {
  const { data } = await apiClient.get<PublicationOption[]>("/factory/publications");
  return data;
}

export async function fetchArticlePrompt(): Promise<PromptTemplate> {
  const { data } = await apiClient.get<PromptTemplate>("/factory/articles/prompt");
  return data;
}

export async function updateArticlePrompt(content: string): Promise<PromptTemplate> {
  const { data } = await apiClient.put<PromptTemplate>("/factory/articles/prompt", { content });
  return data;
}

export async function fetchArticleSources(): Promise<ArticleSource[]> {
  const { data } = await apiClient.get<ArticleSource[]>("/factory/articles/sources");
  return data;
}

export async function createArticleSource(input: {
  url: string;
  publication: string;
}): Promise<ArticleSource> {
  const { data } = await apiClient.post<ArticleSource>("/factory/articles/sources", input);
  return data;
}

export async function deleteArticleSource(id: string): Promise<void> {
  await apiClient.delete(`/factory/articles/sources/${id}`);
}

export async function analyzeArticleSource(sourceId: string): Promise<ArticleAnalyzeResult> {
  const { data } = await apiClient.post<ArticleAnalyzeResult>(`/factory/articles/sources/${sourceId}/analyze`);
  return data;
}

export async function fetchArticleCandidates(sourceId?: string): Promise<ArticleCandidate[]> {
  const params = sourceId ? { source_id: sourceId } : undefined;
  const { data } = await apiClient.get<ArticleCandidate[]>("/factory/articles/candidates", { params });
  return data;
}

export async function updateArticleCandidate(
  id: string,
  patch: Partial<Pick<ArticleCandidate, "post_text" | "status">>
): Promise<ArticleCandidate> {
  const { data } = await apiClient.patch<ArticleCandidate>(`/factory/articles/candidates/${id}`, patch);
  return data;
}

export async function deleteArticleCandidate(id: string): Promise<void> {
  await apiClient.delete(`/factory/articles/candidates/${id}`);
}

export async function rewriteArticleCandidate(
  id: string,
  instruction: string,
  patch?: Partial<Pick<ArticleCandidate, "post_text">>
): Promise<ArticleCandidate> {
  const { data } = await apiClient.post<ArticleCandidate>(`/factory/articles/candidates/${id}/rewrite`, {
    instruction,
    ...patch,
  });
  return data;
}

export async function scheduleArticleCandidate(id: string, scheduledAt: string): Promise<ArticleCandidate> {
  const { data } = await apiClient.post<ArticleCandidate>(`/factory/articles/candidates/${id}/schedule`, {
    scheduled_at: scheduledAt,
  });
  return data;
}

export async function postNowArticleCandidate(
  id: string,
  patch?: Partial<Pick<ArticleCandidate, "post_text">>
): Promise<ArticleCandidate> {
  const { data } = await apiClient.post<ArticleCandidate>(
    `/factory/articles/candidates/${id}/post-now`,
    patch ?? {}
  );
  return data;
}

export async function fetchArticleScheduled(): Promise<ArticleCandidate[]> {
  const { data } = await apiClient.get<ArticleCandidate[]>("/factory/articles/scheduled");
  return data;
}

export async function tickArticleScheduler(): Promise<{ prepared: string[] }> {
  const { data } = await apiClient.post<{ prepared: string[] }>("/factory/articles/scheduler/tick");
  return data;
}
