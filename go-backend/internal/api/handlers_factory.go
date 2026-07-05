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

	r.Get("/factory/sources", a.handleListSources)
	r.Post("/factory/sources", a.handleCreateSource)
	r.Get("/factory/sources/{id}", a.handleGetSource)
	r.Post("/factory/sources/{id}/analyze", a.handleAnalyzeSource)

	r.Get("/factory/candidates", a.handleListCandidates)
	r.Get("/factory/candidates/{id}", a.handleGetCandidate)
	r.Patch("/factory/candidates/{id}", a.handlePatchCandidate)
	r.Post("/factory/candidates/{id}/clip", a.handleClipCandidate)
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
		Title      string `json:"title"`
		Podcast    string `json:"podcast"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.YouTubeURL) == "" {
		writeError(w, http.StatusBadRequest, "youtube_url required")
		return
	}

	title := strings.TrimSpace(body.Title)
	podcast := strings.TrimSpace(body.Podcast)
	if title == "" || podcast == "" {
		t, p := factory.FetchYouTubeMetadata(body.YouTubeURL)
		if title == "" {
			title = t
		}
		if podcast == "" {
			podcast = p
		}
	}

	id := factory.NewSourceID(body.YouTubeURL, podcast)
	src, err := a.factory.CreateSource(id, body.YouTubeURL, title, podcast)
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

	analysis, err := a.runner.Analyze(sourceID, bias.Content, prompt.Content)
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
		Hook     string `json:"hook"`
		Take     string `json:"take"`
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
		postText = factory.EnsurePostTextAttribution(postText, src.Podcast, src.YouTubeURL)
	}
	c, err := a.factory.UpdateCandidate(id, body.Hook, body.Take, postText, body.Status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, c)
}

func (a *App) handleClipCandidate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	c, err := a.factory.GetCandidate(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "candidate not found")
		return
	}
	src, err := a.factory.GetSource(c.SourceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	clipPath, err := a.runner.ClipCandidate(src, c)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	_ = a.factory.SetCandidateClip(id, clipPath)
	c, _ = a.factory.GetCandidate(id)
	writeJSON(w, c)
}

func (a *App) handlePostNowCandidate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Hook     string `json:"hook"`
		Take     string `json:"take"`
		PostText string `json:"post_text"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	if strings.TrimSpace(body.Hook) != "" || strings.TrimSpace(body.Take) != "" || strings.TrimSpace(body.PostText) != "" {
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
			postText = factory.EnsurePostTextAttribution(postText, src.Podcast, src.YouTubeURL)
		}
		if _, err := a.factory.UpdateCandidate(id, body.Hook, body.Take, postText, ""); err != nil {
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

	if c.ClipPath == "" {
		clipPath, err := a.runner.ClipCandidate(src, c)
		if err != nil {
			return factory.Candidate{}, err
		}
		if err := a.factory.SetCandidateClip(id, clipPath); err != nil {
			return factory.Candidate{}, err
		}
		c.ClipPath = clipPath
	} else if err := a.runner.WriteCandidateDraft(src, c, c.ClipPath); err != nil {
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
	if bias == "" || prompt == "" {
		return fmt.Errorf("missing seed files")
	}
	return store.SeedDefaults(bias, prompt)
}