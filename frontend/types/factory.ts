export interface BiasProfile {
  id: number;
  name: string;
  content: string;
  is_active: boolean;
  updated_at: string;
}

export interface PromptTemplate {
  id: number;
  name: string;
  content: string;
  is_active: boolean;
  updated_at: string;
}

export interface MentionDictionaryProfile {
  id: number;
  name: string;
  content: string;
  is_active: boolean;
  updated_at: string;
}

export interface PodcastOption {
  name: string;
  handle: string;
}

export interface Source {
  id: string;
  youtube_url: string;
  podcast: string;
  status: string;
  error_message?: string;
  created_at: string;
  analyzed_at?: string;
}

export interface Candidate {
  id: string;
  source_id: string;
  rank: number;
  start_time: string;
  end_time: string;
  post_text: string;
  why_interesting: string;
  confidence: number;
  clip_path?: string;
  status: string;
  created_at: string;
  updated_at: string;
  scheduled_at?: string;
}

export interface ScheduledPost {
  id: string;
  candidate_id: string;
  scheduled_at: string;
  status: string;
  prepared_at?: string;
  created_at: string;
  candidate?: Candidate;
}

export interface AnalyzeResult {
  source_id: string;
  candidates: Candidate[];
}

export interface SchedulerTickResult {
  prepared: string[];
}