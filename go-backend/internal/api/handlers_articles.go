package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"tid/go-backend/internal/factory"
)

func (a *App) mountArticleRoutes(r chi.Router) {
	r.Get("/factory/articles/prompt", a.handleGetArticlePrompt)
	r.Put("/factory/articles/prompt", a.handlePutArticlePrompt)

	r.Get("/factory/articles/sources", a.handleListArticleSources)
	r.Post("/factory/articles/sources", a.handleCreateArticleSource)
	r.Get("/factory/articles/sources/{id}", a.handleGetArticleSource)
	r.Delete("/factory/articles/sources/{id}", a.handleDeleteArticleSource)
	r.Post("/factory/articles/sources/{id}/analyze", a.handleAnalyzeArticleSource)

	r.Get("/factory/articles/candidates", a.handleListArticleCandidates)
	r.Get("/factory/articles/candidates/{id}", a.handleGetArticleCandidate)
	r.Patch("/factory/articles/candidates/{id}", a.handlePatchArticleCandidate)
	r.Delete("/factory/articles/candidates/{id}", a.handleDeleteArticleCandidate)
	r.Post("/factory/articles/candidates/{id}/rewrite", a.handleRewriteArticleCandidate)
	r.Post("/factory/articles/candidates/{id}/schedule", a.handleScheduleArticleCandidate)
	r.Post("/factory/articles/candidates/{id}/post-now", a.handlePostNowArticleCandidate)

	r.Get("/factory/articles/scheduled", a.handleListArticleScheduled)
	r.Post("/factory/articles/scheduler/tick", a.handleArticleSchedulerTick)
}

func (a *App) handleGetArticlePrompt(w http.ResponseWriter, _ *http.Request) {
	p, err := a.factory.GetPromptByName(factory.ArticlePromptName)
	if err != nil {
		writeError(w, http.StatusNotFound, "article prompt not found")
		return
	}
	writeJSON(w, p)
}

func (a *App) handlePutArticlePrompt(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Content) == "" {
		writeError(w, http.StatusBadRequest, "content required")
		return
	}
	p, err := a.factory.UpsertPromptByName(factory.ArticlePromptName, body.Content)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, p)
}

func (a *App) handleListArticleSources(w http.ResponseWriter, _ *http.Request) {
	items, err := a.factory.ListArticleSources()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if items == nil {
		items = []factory.ArticleSource{}
	}
	writeJSON(w, items)
}

func (a *App) handleCreateArticleSource(w http.ResponseWriter, r *http.Request) {
	var body struct {
		URL         string `json:"url"`
		Publication string `json:"publication"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.URL) == "" {
		writeError(w, http.StatusBadRequest, "url required")
		return
	}
	publication := strings.TrimSpace(body.Publication)
	if publication == "" {
		writeError(w, http.StatusBadRequest, "publication required")
		return
	}

	dict, err := a.activeMentionDict()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if dict.ResolvePublicationHandle(publication) == "" {
		writeError(w, http.StatusBadRequest, "publication must match a news feed in the sources dictionary")
		return
	}

	id := factory.NewArticleSourceID(body.URL, publication)
	src, err := a.factory.CreateArticleSource(id, strings.TrimSpace(body.URL), publication)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, src)
}

func (a *App) handleGetArticleSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	src, err := a.factory.GetArticleSource(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "article source not found")
		return
	}
	writeJSON(w, src)
}

func (a *App) handleDeleteArticleSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := a.factory.DeleteArticleSource(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "article source not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, map[string]string{"deleted": id})
}

func (a *App) handleAnalyzeArticleSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	result, err := a.runAnalyzeArticle(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, result)
}

func (a *App) runAnalyzeArticle(sourceID string) (map[string]any, error) {
	src, err := a.factory.GetArticleSource(sourceID)
	if err != nil {
		return nil, err
	}
	_ = a.factory.SetArticleSourceStatus(sourceID, "analyzing", "")

	title, err := a.runner.FetchArticle(sourceID, src.URL)
	if err != nil {
		_ = a.factory.SetArticleSourceStatus(sourceID, "failed", err.Error())
		return nil, err
	}
	_ = a.factory.SetArticleSourceTitle(sourceID, title)

	bias, err := a.factory.GetActiveBias()
	if err != nil {
		_ = a.factory.SetArticleSourceStatus(sourceID, "failed", err.Error())
		return nil, err
	}
	prompt, err := a.factory.GetPromptByName(factory.ArticlePromptName)
	if err != nil {
		_ = a.factory.SetArticleSourceStatus(sourceID, "failed", err.Error())
		return nil, err
	}
	mentions, err := a.factory.GetActiveMentions()
	if err != nil {
		_ = a.factory.SetArticleSourceStatus(sourceID, "failed", err.Error())
		return nil, err
	}

	analysis, err := a.runner.AnalyzeArticle(sourceID, src.Publication, bias.Content, prompt.Content, mentions.Content)
	if err != nil {
		_ = a.factory.SetArticleSourceStatus(sourceID, "failed", err.Error())
		return nil, err
	}
	if strings.TrimSpace(analysis.Title) != "" {
		_ = a.factory.SetArticleSourceTitle(sourceID, analysis.Title)
	}

	candidates, err := a.factory.InsertArticleCandidates(sourceID, analysis.Candidates)
	if err != nil {
		_ = a.factory.SetArticleSourceStatus(sourceID, "failed", err.Error())
		return nil, err
	}
	_ = a.factory.MarkArticleSourceAnalyzed(sourceID)
	return map[string]any{"source_id": sourceID, "candidates": candidates}, nil
}

func (a *App) handleListArticleCandidates(w http.ResponseWriter, r *http.Request) {
	sourceID := r.URL.Query().Get("source_id")
	items, err := a.factory.ListArticleCandidates(sourceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if items == nil {
		items = []factory.ArticleCandidate{}
	}
	writeJSON(w, items)
}

func (a *App) handleGetArticleCandidate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	c, err := a.factory.GetArticleCandidate(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "article candidate not found")
		return
	}
	writeJSON(w, c)
}

func (a *App) handlePatchArticleCandidate(w http.ResponseWriter, r *http.Request) {
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
		cand, err := a.factory.GetArticleCandidate(id)
		if err != nil {
			writeError(w, http.StatusNotFound, "article candidate not found")
			return
		}
		src, err := a.factory.GetArticleSource(cand.SourceID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		dict, err := a.activeMentionDict()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		postText = factory.EnsureArticlePostAttribution(postText, src.Publication, dict)
	}
	c, err := a.factory.UpdateArticleCandidate(id, postText, body.Status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, c)
}

func (a *App) handleDeleteArticleCandidate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := a.factory.DeleteArticleCandidate(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "article candidate not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, map[string]string{"deleted": id})
}

func (a *App) handleRewriteArticleCandidate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Instruction string `json:"instruction"`
		PostText    string `json:"post_text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Instruction) == "" {
		writeError(w, http.StatusBadRequest, "instruction required")
		return
	}

	c, err := a.factory.GetArticleCandidate(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "article candidate not found")
		return
	}
	if strings.TrimSpace(body.PostText) != "" {
		c.PostText = body.PostText
	}

	src, err := a.factory.GetArticleSource(c.SourceID)
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

	rewritten, err := a.runner.RewriteArticlePost(bias.Content, mentions.Content, src, c, body.Instruction)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	updated, err := a.factory.UpdateArticleCandidate(id, rewritten.PostText, "")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, updated)
}

func (a *App) handleScheduleArticleCandidate(w http.ResponseWriter, r *http.Request) {
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
	c, err := a.factory.ScheduleArticleCandidate(id, at)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, c)
}

func (a *App) handlePostNowArticleCandidate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		PostText string `json:"post_text"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	if strings.TrimSpace(body.PostText) != "" {
		cand, err := a.factory.GetArticleCandidate(id)
		if err != nil {
			writeError(w, http.StatusNotFound, "article candidate not found")
			return
		}
		src, err := a.factory.GetArticleSource(cand.SourceID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		dict, derr := a.activeMentionDict()
		if derr != nil {
			writeError(w, http.StatusInternalServerError, derr.Error())
			return
		}
		postText := factory.EnsureArticlePostAttribution(body.PostText, src.Publication, dict)
		if _, err := a.factory.UpdateArticleCandidate(id, postText, ""); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	c, err := a.runPrepareArticleCandidate(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, c)
}

func (a *App) runPrepareArticleCandidate(id string) (factory.ArticleCandidate, error) {
	c, err := a.factory.GetArticleCandidate(id)
	if err != nil {
		return factory.ArticleCandidate{}, err
	}
	src, err := a.factory.GetArticleSource(c.SourceID)
	if err != nil {
		return factory.ArticleCandidate{}, err
	}
	dict, err := a.activeMentionDict()
	if err != nil {
		return factory.ArticleCandidate{}, err
	}
	if err := a.runner.WriteArticleDraft(src, c, dict); err != nil {
		return factory.ArticleCandidate{}, err
	}
	if err := a.runner.PrepareArticlePost(id); err != nil {
		return factory.ArticleCandidate{}, err
	}
	return a.factory.GetArticleCandidate(id)
}

func (a *App) handleListArticleScheduled(w http.ResponseWriter, _ *http.Request) {
	items, err := a.factory.ListScheduledArticleCandidates()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if items == nil {
		items = []factory.ArticleCandidate{}
	}
	writeJSON(w, items)
}

func (a *App) handleArticleSchedulerTick(w http.ResponseWriter, _ *http.Request) {
	prepared, err := a.runArticleSchedulerTick()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, map[string]any{"prepared": prepared})
}

func (a *App) runArticleSchedulerTick() ([]string, error) {
	due, err := a.factory.ListDueArticleCandidates(time.Now())
	if err != nil {
		return nil, err
	}
	var prepared []string
	for _, c := range due {
		if _, err := a.runPrepareArticleCandidate(c.ID); err != nil {
			a.logger.Printf("prepare article post failed for %s: %v", c.ID, err)
			continue
		}
		_ = a.factory.MarkArticleCandidatePrepared(c.ID)
		prepared = append(prepared, c.ID)
	}
	return prepared, nil
}
