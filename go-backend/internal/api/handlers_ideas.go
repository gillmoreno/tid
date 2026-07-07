package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"tid/go-backend/internal/factory"
)

func (a *App) mountIdeaRoutes(r chi.Router) {
	r.Get("/factory/ideas", a.handleListIdeas)
	r.Post("/factory/ideas", a.handleCreateIdea)
	r.Get("/factory/ideas/{id}", a.handleGetIdea)
	r.Patch("/factory/ideas/{id}", a.handlePatchIdea)
	r.Delete("/factory/ideas/{id}", a.handleDeleteIdea)
}

func (a *App) handleListIdeas(w http.ResponseWriter, _ *http.Request) {
	items, err := a.factory.ListIdeas()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if items == nil {
		items = []factory.Idea{}
	}
	writeJSON(w, items)
}

func (a *App) handleGetIdea(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idea, err := a.factory.GetIdea(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "idea not found")
		return
	}
	writeJSON(w, idea)
}

func (a *App) handleCreateIdea(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title   string   `json:"title"`
		Slug    string   `json:"slug"`
		Kind    string   `json:"kind"`
		Status  string   `json:"status"`
		Summary string   `json:"summary"`
		Body    string   `json:"body"`
		XPost   string   `json:"x_post"`
		Tags    []string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if strings.TrimSpace(body.Title) == "" {
		writeError(w, http.StatusBadRequest, "title required")
		return
	}

	idea, err := a.factory.CreateIdea(factory.Idea{
		Title:   body.Title,
		Slug:    body.Slug,
		Kind:    body.Kind,
		Status:  body.Status,
		Summary: body.Summary,
		Body:    body.Body,
		XPost:   body.XPost,
		Tags:    body.Tags,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, idea)
}

func (a *App) handlePatchIdea(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Title   string   `json:"title"`
		Kind    string   `json:"kind"`
		Status  string   `json:"status"`
		Summary string   `json:"summary"`
		Body    string   `json:"body"`
		XPost   string   `json:"x_post"`
		Tags    []string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	current, err := a.factory.GetIdea(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "idea not found")
		return
	}
	if strings.TrimSpace(body.Title) != "" {
		current.Title = body.Title
	}
	if strings.TrimSpace(body.Kind) != "" {
		current.Kind = body.Kind
	}
	if strings.TrimSpace(body.Status) != "" {
		current.Status = body.Status
	}
	current.Summary = body.Summary
	current.Body = body.Body
	current.XPost = body.XPost
	if body.Tags != nil {
		current.Tags = body.Tags
	}

	idea, err := a.factory.UpdateIdea(id, current)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "idea not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, idea)
}

func (a *App) handleDeleteIdea(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := a.factory.DeleteIdea(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "idea not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, map[string]string{"deleted": id})
}