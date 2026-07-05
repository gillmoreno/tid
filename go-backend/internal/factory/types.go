package factory

type BiasProfile struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Content   string `json:"content"`
	IsActive  bool   `json:"is_active"`
	UpdatedAt string `json:"updated_at"`
}

type PromptTemplate struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Content   string `json:"content"`
	IsActive  bool   `json:"is_active"`
	UpdatedAt string `json:"updated_at"`
}

type MentionDictionaryProfile struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Content   string `json:"content"`
	IsActive  bool   `json:"is_active"`
	UpdatedAt string `json:"updated_at"`
}

type Source struct {
	ID           string  `json:"id"`
	YouTubeURL   string  `json:"youtube_url"`
	Title        string  `json:"title"`
	Podcast      string  `json:"podcast"`
	Status       string  `json:"status"`
	ErrorMessage string  `json:"error_message,omitempty"`
	CreatedAt    string  `json:"created_at"`
	AnalyzedAt   *string `json:"analyzed_at,omitempty"`
}

type Candidate struct {
	ID             string  `json:"id"`
	SourceID       string  `json:"source_id"`
	Rank           int     `json:"rank"`
	StartTime      string  `json:"start_time"`
	EndTime        string  `json:"end_time"`
	Hook           string  `json:"hook"`
	Take           string  `json:"take"`
	PostText       string  `json:"post_text"`
	WhyInteresting string  `json:"why_interesting"`
	Confidence     float64 `json:"confidence"`
	ClipPath       string  `json:"clip_path,omitempty"`
	Status         string  `json:"status"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
	ScheduledAt    *string `json:"scheduled_at,omitempty"`
}

type ScheduledPost struct {
	ID          string  `json:"id"`
	CandidateID string  `json:"candidate_id"`
	ScheduledAt string  `json:"scheduled_at"`
	Status      string  `json:"status"`
	PreparedAt  *string `json:"prepared_at,omitempty"`
	CreatedAt   string  `json:"created_at"`
	Candidate   *Candidate `json:"candidate,omitempty"`
}

type AnalysisCandidate struct {
	StartTime      string  `json:"start_time"`
	EndTime        string  `json:"end_time"`
	Hook           string  `json:"hook"`
	Take           string  `json:"take"`
	PostText       string  `json:"post_text"`
	WhyInteresting string  `json:"why_interesting"`
	Confidence     float64 `json:"confidence"`
}

type AnalysisResult struct {
	Candidates []AnalysisCandidate `json:"candidates"`
}

type RewriteResult struct {
	Hook     string `json:"hook"`
	Take     string `json:"take"`
	PostText string `json:"post_text"`
}