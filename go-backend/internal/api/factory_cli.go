package api

import (
	"time"

	"tid/go-backend/internal/factory"
)

func (a *App) IngestSource(url, title, podcast string) (factory.Source, error) {
	id := factory.NewSourceID(url, podcast)
	return a.factory.CreateSource(id, url, title, podcast)
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
	clipPath, err := a.runner.ClipCandidate(src, c)
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