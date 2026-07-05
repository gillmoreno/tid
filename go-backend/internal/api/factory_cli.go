package api

import (
	"strings"
	"time"

	"tid/go-backend/internal/factory"
)

func (a *App) IngestSource(url, title, podcast string) (factory.Source, error) {
	t := strings.TrimSpace(title)
	p := strings.TrimSpace(podcast)
	if t == "" || p == "" {
		tt, pp := factory.FetchYouTubeMetadata(url)
		if t == "" {
			t = tt
		}
		if p == "" {
			p = pp
		}
	}
	id := factory.NewSourceID(url, p)
	return a.factory.CreateSource(id, url, t, p)
}

func (a *App) AnalyzeSourceCLI(sourceID string) ([]factory.Candidate, error) {
	result, err := a.runAnalyze(sourceID)
	if err != nil {
		return nil, err
	}
	candidates, _ := result["candidates"].([]factory.Candidate)
	return candidates, nil
}

func (a *App) ClipCandidateCLI(candidateID string) error {
	c, err := a.factory.GetCandidate(candidateID)
	if err != nil {
		return err
	}
	src, err := a.factory.GetSource(c.SourceID)
	if err != nil {
		return err
	}
	mentions, err := a.factory.GetActiveMentions()
	if err != nil {
		return err
	}
	dict := factory.ParseMentionDictionary(mentions.Content)
	clipPath, err := a.runner.ClipCandidate(src, c, dict)
	if err != nil {
		return err
	}
	return a.factory.SetCandidateClip(candidateID, clipPath)
}

func (a *App) ScheduleCandidateCLI(candidateID string, at time.Time) (factory.ScheduledPost, error) {
	return a.factory.ScheduleCandidate(candidateID, at)
}

func (a *App) SchedulerTickCLI() ([]string, error) {
	return a.runSchedulerTick()
}

func (a *App) PostNowCLI(candidateID string) (factory.Candidate, error) {
	return a.runPrepareCandidate(candidateID)
}