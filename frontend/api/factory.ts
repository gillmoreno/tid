import { apiClient } from "@/api/client";
import type {
  AnalyzeResult,
  BiasProfile,
  Candidate,
  MentionDictionaryProfile,
  PodcastOption,
  PromptTemplate,
  ScheduledPost,
  SchedulerTickResult,
  Source,
} from "@/types/factory";

export async function fetchBiases(): Promise<BiasProfile> {
  const { data } = await apiClient.get<BiasProfile>("/factory/biases");
  return data;
}

export async function updateBiases(content: string): Promise<BiasProfile> {
  const { data } = await apiClient.put<BiasProfile>("/factory/biases", { content });
  return data;
}

export async function fetchPrompt(): Promise<PromptTemplate> {
  const { data } = await apiClient.get<PromptTemplate>("/factory/prompt");
  return data;
}

export async function updatePrompt(content: string): Promise<PromptTemplate> {
  const { data } = await apiClient.put<PromptTemplate>("/factory/prompt", { content });
  return data;
}

export async function fetchMentions(): Promise<MentionDictionaryProfile> {
  const { data } = await apiClient.get<MentionDictionaryProfile>("/factory/mentions");
  return data;
}

export async function updateMentions(content: string): Promise<MentionDictionaryProfile> {
  const { data } = await apiClient.put<MentionDictionaryProfile>("/factory/mentions", { content });
  return data;
}

export async function fetchSources(): Promise<Source[]> {
  const { data } = await apiClient.get<Source[]>("/factory/sources");
  return data;
}

export async function fetchPodcasts(): Promise<PodcastOption[]> {
  const { data } = await apiClient.get<PodcastOption[]>("/factory/podcasts");
  return data;
}

export async function createSource(input: {
  youtube_url: string;
  podcast: string;
}): Promise<Source> {
  const { data } = await apiClient.post<Source>("/factory/sources", input);
  return data;
}

export async function deleteSource(id: string): Promise<void> {
  await apiClient.delete(`/factory/sources/${id}`);
}

export async function analyzeSource(sourceId: string): Promise<AnalyzeResult> {
  const { data } = await apiClient.post<AnalyzeResult>(`/factory/sources/${sourceId}/analyze`);
  return data;
}

export async function fetchCandidates(sourceId?: string): Promise<Candidate[]> {
  const params = sourceId ? { source_id: sourceId } : undefined;
  const { data } = await apiClient.get<Candidate[]>("/factory/candidates", { params });
  return data;
}

export async function deleteCandidate(id: string): Promise<void> {
  await apiClient.delete(`/factory/candidates/${id}`);
}

export async function updateCandidate(
  id: string,
  patch: Partial<Pick<Candidate, "post_text" | "status">>
): Promise<Candidate> {
  const { data } = await apiClient.patch<Candidate>(`/factory/candidates/${id}`, patch);
  return data;
}

export async function clipCandidate(id: string): Promise<Candidate> {
  const { data } = await apiClient.post<Candidate>(`/factory/candidates/${id}/clip`);
  return data;
}

export async function trimCandidate(
  id: string,
  patch: Partial<Pick<Candidate, "start_time" | "end_time">>
): Promise<Candidate> {
  const { data } = await apiClient.post<Candidate>(`/factory/candidates/${id}/trim`, patch);
  return data;
}

export async function rewriteCandidate(
  id: string,
  instruction: string,
  patch?: Partial<Pick<Candidate, "post_text">>
): Promise<Candidate> {
  const { data } = await apiClient.post<Candidate>(`/factory/candidates/${id}/rewrite`, {
    instruction,
    ...patch,
  });
  return data;
}

export async function postNowCandidate(
  id: string,
  patch?: Partial<Pick<Candidate, "post_text">>
): Promise<Candidate> {
  const { data } = await apiClient.post<Candidate>(`/factory/candidates/${id}/post-now`, patch ?? {});
  return data;
}

export async function scheduleCandidate(id: string, scheduledAt: string): Promise<ScheduledPost> {
  const { data } = await apiClient.post<ScheduledPost>(`/factory/candidates/${id}/schedule`, {
    scheduled_at: scheduledAt,
  });
  return data;
}

export async function fetchScheduled(): Promise<ScheduledPost[]> {
  const { data } = await apiClient.get<ScheduledPost[]>("/factory/scheduled");
  return data;
}

export async function tickScheduler(): Promise<SchedulerTickResult> {
  const { data } = await apiClient.post<SchedulerTickResult>("/factory/scheduler/tick");
  return data;
}