package api

import (
	"fmt"
	"strings"
	"time"

	"tid/go-backend/internal/factory"
)

func (a *App) IngestSource(url, podcast string) (factory.Source, error) {
	p := strings.TrimSpace(podcast)
	if p == "" {
		return factory.Source{}, fmt.Errorf("podcast required")
	}
	id := factory.NewSourceID(url, p)
	return a.factory.CreateSource(id, url, p)
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
	_, err := a.runClipCandidate(candidateID)
	return err
}

func (a *App) TrimCandidateCLI(candidateID, startTime, endTime string) (factory.Candidate, error) {
	return a.runTrimCandidate(candidateID, startTime, endTime)
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