package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"tid/go-backend/internal/factory"
)

func (a *App) mountFactoryRoutes(r chi.Router) {
	r.Get("/factory/biases", a.handleGetBiases)
	r.Put("/factory/biases", a.handlePutBiases)
	r.Get("/factory/prompt", a.handleGetPrompt)
	r.Put("/factory/prompt", a.handlePutPrompt)
	r.Get("/factory/mentions", a.handleGetMentions)
	r.Put("/factory/mentions", a.handlePutMentions)
	r.Get("/factory/podcasts", a.handleListPodcasts)

	r.Get("/factory/sources", a.handleListSources)
	r.Post("/factory/sources", a.handleCreateSource)
	r.Get("/factory/sources/{id}", a.handleGetSource)
	r.Delete("/factory/sources/{id}", a.handleDeleteSource)
	r.Post("/factory/sources/{id}/analyze", a.handleAnalyzeSource)

	r.Get("/factory/candidates", a.handleListCandidates)
	r.Get("/factory/candidates/{id}", a.handleGetCandidate)
	r.Patch("/factory/candidates/{id}", a.handlePatchCandidate)
	r.Delete("/factory/candidates/{id}", a.handleDeleteCandidate)
	r.Post("/factory/candidates/{id}/clip", a.handleClipCandidate)
	r.Post("/factory/candidates/{id}/trim", a.handleTrimCandidate)
	r.Post("/factory/candidates/{id}/rewrite", a.handleRewriteCandidate)
	r.Post("/factory/candidates/{id}/post-now", a.handlePostNowCandidate)
	r.Post("/factory/candidates/{id}/schedule", a.handleScheduleCandidate)

	r.Get("/factory/scheduled", a.handleListScheduled)
	r.Post("/factory/scheduler/tick", a.handleSchedulerTick)
}

func (a *App) handleGetBiases(w http.ResponseWriter, _ *http.Request) {
	b, err := a.factory.GetActiveBias()
	if err != nil {
		writeError(w, http.StatusNotFound, "biases not found")
		return
	}
	writeJSON(w, b)
}

func (a *App) handlePutBiases(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Content) == "" {
		writeError(w, http.StatusBadRequest, "content required")
		return
	}
	b, err := a.factory.UpdateActiveBias(body.Content)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, b)
}

func (a *App) handleGetPrompt(w http.ResponseWriter, _ *http.Request) {
	p, err := a.factory.GetActivePrompt()
	if err != nil {
		writeError(w, http.StatusNotFound, "prompt not found")
		return
	}
	writeJSON(w, p)
}

func (a *App) handlePutPrompt(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Content) == "" {
		writeError(w, http.StatusBadRequest, "content required")
		return
	}
	p, err := a.factory.UpdateActivePrompt(body.Content)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, p)
}

func (a *App) handleGetMentions(w http.ResponseWriter, _ *http.Request) {
	m, err := a.factory.GetActiveMentions()
	if err != nil {
		writeError(w, http.StatusNotFound, "mentions not found")
		return
	}
	writeJSON(w, m)
}

func (a *App) handlePutMentions(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Content) == "" {
		writeError(w, http.StatusBadRequest, "content required")
		return
	}
	m, err := a.factory.UpdateActiveMentions(body.Content)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, m)
}

func (a *App) activeMentionDict() (factory.MentionDictionary, error) {
	m, err := a.factory.GetActiveMentions()
	if err != nil {
		return factory.MentionDictionary{}, err
	}
	return factory.ParseMentionDictionary(m.Content), nil
}

func (a *App) handleListPodcasts(w http.ResponseWriter, _ *http.Request) {
	m, err := a.factory.GetActiveMentions()
	if err != nil {
		writeError(w, http.StatusNotFound, "mentions not found")
		return
	}
	dict := factory.ParseMentionDictionary(m.Content)
	opts := dict.PodcastOptions()
	if opts == nil {
		opts = []factory.PodcastOption{}
	}
	writeJSON(w, opts)
}

func (a *App) handleListSources(w http.ResponseWriter, _ *http.Request) {
	items, err := a.factory.ListSources()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if items == nil {
		items = []factory.Source{}
	}
	writeJSON(w, items)
}

func (a *App) handleCreateSource(w http.ResponseWriter, r *http.Request) {
	var body struct {
		YouTubeURL string `json:"youtube_url"`
		Podcast    string `json:"podcast"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.YouTubeURL) == "" {
		writeError(w, http.StatusBadRequest, "youtube_url required")
		return
	}

	podcast := strings.TrimSpace(body.Podcast)
	if podcast == "" {
		writeError(w, http.StatusBadRequest, "podcast required")
		return
	}

	dict, err := a.activeMentionDict()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if dict.ResolvePodcastHandle(podcast) == "" {
		writeError(w, http.StatusBadRequest, "podcast must match a name in the mentions dictionary")
		return
	}

	id := factory.NewSourceID(body.YouTubeURL, podcast)
	src, err := a.factory.CreateSource(id, body.YouTubeURL, podcast)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, src)
}

func (a *App) handleGetSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	src, err := a.factory.GetSource(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "source not found")
		return
	}
	writeJSON(w, src)
}

func (a *App) handleDeleteSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := a.factory.DeleteSource(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "source not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, map[string]string{"deleted": id})
}

func (a *App) handleAnalyzeSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	result, err := a.runAnalyze(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, result)
}

func (a *App) runAnalyze(sourceID string) (map[string]any, error) {
	src, err := a.factory.GetSource(sourceID)
	if err != nil {
		return nil, err
	}
	_ = a.factory.SetSourceStatus(sourceID, "analyzing", "")

	if err := a.runner.FetchTranscript(sourceID, src.YouTubeURL); err != nil {
		_ = a.factory.SetSourceStatus(sourceID, "failed", err.Error())
		return nil, err
	}

	bias, err := a.factory.GetActiveBias()
	if err != nil {
		_ = a.factory.SetSourceStatus(sourceID, "failed", err.Error())
		return nil, err
	}
	prompt, err := a.factory.GetActivePrompt()
	if err != nil {
		_ = a.factory.SetSourceStatus(sourceID, "failed", err.Error())
		return nil, err
	}

	mentions, err := a.factory.GetActiveMentions()
	if err != nil {
		_ = a.factory.SetSourceStatus(sourceID, "failed", err.Error())
		return nil, err
	}

	analysis, err := a.runner.Analyze(sourceID, bias.Content, prompt.Content, mentions.Content)
	if err != nil {
		_ = a.factory.SetSourceStatus(sourceID, "failed", err.Error())
		return nil, err
	}

	candidates, err := a.factory.InsertCandidates(sourceID, analysis.Candidates)
	if err != nil {
		_ = a.factory.SetSourceStatus(sourceID, "failed", err.Error())
		return nil, err
	}
	_ = a.factory.MarkSourceAnalyzed(sourceID)
	return map[string]any{"source_id": sourceID, "candidates": candidates}, nil
}

func (a *App) handleListCandidates(w http.ResponseWriter, r *http.Request) {
	sourceID := r.URL.Query().Get("source_id")
	items, err := a.factory.ListCandidates(sourceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if items == nil {
		items = []factory.Candidate{}
	}
	writeJSON(w, items)
}

func (a *App) handleGetCandidate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	c, err := a.factory.GetCandidateEnriched(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "candidate not found")
		return
	}
	writeJSON(w, c)
}

func (a *App) handlePatchCandidate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		PostText string `json:"post_text"`
		Status   string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	postText := body.PostText
	if strings.TrimSpace(postText) != "" {
		cand, err := a.factory.GetCandidate(id)
		if err != nil {
			writeError(w, http.StatusNotFound, "candidate not found")
			return
		}
		src, err := a.factory.GetSource(cand.SourceID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		dict, err := a.activeMentionDict()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		postText = factory.EnsurePostTextAttribution(postText, src.Podcast, dict)
	}
	c, err := a.factory.UpdateCandidate(id, postText, body.Status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, c)
}

func (a *App) handleDeleteCandidate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := a.factory.DeleteCandidate(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "candidate not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, map[string]string{"deleted": id})
}

func (a *App) handleClipCandidate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	c, err := a.runClipCandidate(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "candidate not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, c)
}

func (a *App) runClipCandidate(id string) (factory.Candidate, error) {
	c, err := a.factory.GetCandidate(id)
	if err != nil {
		return factory.Candidate{}, err
	}
	src, err := a.factory.GetSource(c.SourceID)
	if err != nil {
		return factory.Candidate{}, err
	}
	dict, err := a.activeMentionDict()
	if err != nil {
		return factory.Candidate{}, err
	}
	clipPath, err := a.runner.ClipCandidate(src, c, dict)
	if err != nil {
		return factory.Candidate{}, err
	}
	_ = a.factory.SetCandidateClip(id, clipPath)
	return a.factory.GetCandidateEnriched(id)
}

func (a *App) handleTrimCandidate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if strings.TrimSpace(body.StartTime) == "" && strings.TrimSpace(body.EndTime) == "" {
		writeError(w, http.StatusBadRequest, "start_time or end_time required")
		return
	}

	c, err := a.runTrimCandidate(id, body.StartTime, body.EndTime)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "candidate not found")
			return
		}
		if strings.Contains(err.Error(), "invalid trim") || strings.Contains(err.Error(), "invalid timestamp") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, c)
}

func (a *App) runTrimCandidate(id, startTime, endTime string) (factory.Candidate, error) {
	c, err := a.factory.GetCandidate(id)
	if err != nil {
		return factory.Candidate{}, err
	}
	oldStart := c.StartTime
	oldEnd := c.EndTime

	newStart := strings.TrimSpace(startTime)
	if newStart == "" {
		newStart = c.StartTime
	}
	newEnd := strings.TrimSpace(endTime)
	if newEnd == "" {
		newEnd = c.EndTime
	}
	if newStart == oldStart && newEnd == oldEnd {
		return factory.Candidate{}, fmt.Errorf("no timestamp changes")
	}

	c, err = a.factory.UpdateCandidateTimes(id, newStart, newEnd)
	if err != nil {
		return factory.Candidate{}, err
	}

	src, err := a.factory.GetSource(c.SourceID)
	if err != nil {
		return factory.Candidate{}, err
	}
	dict, err := a.activeMentionDict()
	if err != nil {
		return factory.Candidate{}, err
	}

	clipPath, err := a.runner.TrimCandidate(src, c, oldStart, oldEnd, dict)
	if err != nil {
		_, _ = a.factory.UpdateCandidateTimes(id, oldStart, oldEnd)
		return factory.Candidate{}, err
	}
	_ = a.factory.SetCandidateClip(id, clipPath)
	return a.factory.GetCandidateEnriched(id)
}

func (a *App) handleRewriteCandidate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Instruction string `json:"instruction"`
		PostText    string `json:"post_text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Instruction) == "" {
		writeError(w, http.StatusBadRequest, "instruction required")
		return
	}

	c, err := a.factory.GetCandidate(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "candidate not found")
		return
	}
	if strings.TrimSpace(body.PostText) != "" {
		c.PostText = body.PostText
	}

	src, err := a.factory.GetSource(c.SourceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	bias, err := a.factory.GetActiveBias()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	mentions, err := a.factory.GetActiveMentions()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	rewritten, err := a.runner.RewriteCandidate(bias.Content, mentions.Content, src, c, body.Instruction)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	updated, err := a.factory.UpdateCandidate(id, rewritten.PostText, "")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, updated)
}

func (a *App) handlePostNowCandidate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		PostText string `json:"post_text"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	if strings.TrimSpace(body.PostText) != "" {
		cand, err := a.factory.GetCandidate(id)
		if err != nil {
			writeError(w, http.StatusNotFound, "candidate not found")
			return
		}
		src, err := a.factory.GetSource(cand.SourceID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		dict, derr := a.activeMentionDict()
		if derr != nil {
			writeError(w, http.StatusInternalServerError, derr.Error())
			return
		}
		postText := factory.EnsurePostTextAttribution(body.PostText, src.Podcast, dict)
		if _, err := a.factory.UpdateCandidate(id, postText, ""); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	c, err := a.runPrepareCandidate(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, c)
}

func (a *App) runPrepareCandidate(id string) (factory.Candidate, error) {
	c, err := a.factory.GetCandidate(id)
	if err != nil {
		return factory.Candidate{}, err
	}
	src, err := a.factory.GetSource(c.SourceID)
	if err != nil {
		return factory.Candidate{}, err
	}

	dict, err := a.activeMentionDict()
	if err != nil {
		return factory.Candidate{}, err
	}

	if c.ClipPath == "" {
		clipPath, err := a.runner.ClipCandidate(src, c, dict)
		if err != nil {
			return factory.Candidate{}, err
		}
		if err := a.factory.SetCandidateClip(id, clipPath); err != nil {
			return factory.Candidate{}, err
		}
		c.ClipPath = clipPath
	} else if err := a.runner.WriteCandidateDraft(src, c, c.ClipPath, dict); err != nil {
		return factory.Candidate{}, err
	}

	if err := a.runner.PreparePost(id); err != nil {
		return factory.Candidate{}, err
	}
	return a.factory.GetCandidateEnriched(id)
}

func (a *App) handleScheduleCandidate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		ScheduledAt string `json:"scheduled_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ScheduledAt == "" {
		writeError(w, http.StatusBadRequest, "scheduled_at required (RFC3339)")
		return
	}
	at, err := time.Parse(time.RFC3339, body.ScheduledAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid scheduled_at")
		return
	}
	sp, err := a.factory.ScheduleCandidate(id, at)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, sp)
}

func (a *App) handleListScheduled(w http.ResponseWriter, _ *http.Request) {
	items, err := a.factory.ListScheduled()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if items == nil {
		items = []factory.ScheduledPost{}
	}
	writeJSON(w, items)
}

func (a *App) handleSchedulerTick(w http.ResponseWriter, _ *http.Request) {
	prepared, err := a.runSchedulerTick()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, map[string]any{"prepared": prepared})
}

func (a *App) runSchedulerTick() ([]string, error) {
	due, err := a.factory.ListDueScheduled(time.Now())
	if err != nil {
		return nil, err
	}
	var prepared []string
	for _, sp := range due {
		if sp.Candidate == nil {
			continue
		}
		if _, err := a.runPrepareCandidate(sp.Candidate.ID); err != nil {
			a.logger.Printf("prepare-post failed for %s: %v", sp.Candidate.ID, err)
			continue
		}
		_ = a.factory.MarkScheduledPrepared(sp.ID)
		prepared = append(prepared, sp.Candidate.ID)
	}
	return prepared, nil
}

func loadSeedFile(repoRoot, rel string) string {
	data, err := os.ReadFile(filepath.Join(repoRoot, rel))
	if err != nil {
		return ""
	}
	return string(data)
}

func seedFactoryStore(store *factory.Store, repoRoot string) error {
	bias := loadSeedFile(repoRoot, "loops/clip-to-post/biases.default.md")
	prompt := loadSeedFile(repoRoot, "loops/clip-to-post/prompt.default.md")
	mentions := loadSeedFile(repoRoot, "loops/clip-to-post/mentions.default.json")
	if bias == "" || prompt == "" {
		return fmt.Errorf("missing seed files")
	}
	if err := store.SeedDefaults(bias, prompt); err != nil {
		return err
	}
	if mentions != "" {
		if err := store.SeedMentions(mentions); err != nil {
			return err
		}
	}
	return nil
}