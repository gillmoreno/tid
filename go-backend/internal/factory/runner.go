package factory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Runner struct {
	RepoRoot string
	LoopsDir string
}

func NewRunner(repoRoot string) Runner {
	loops := filepath.Join(repoRoot, "loops", "clip-to-post")
	return Runner{RepoRoot: repoRoot, LoopsDir: loops}
}

func (r Runner) Run(ctxDir string, name string, args ...string) (string, error) {
	cmd := exec.Command(filepath.Join(r.LoopsDir, name), args...)
	cmd.Dir = r.LoopsDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s: %w\n%s", name, err, stderr.String())
	}
	return stdout.String(), nil
}

func (r Runner) FetchTranscript(sourceID, url string) error {
	workDir := filepath.Join(r.LoopsDir, "drafts", sourceID)
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return err
	}
	_, err := r.Run(workDir, "transcript.sh", "--url", url, "--id", sourceID, "--out", "drafts")
	return err
}

func (r Runner) Analyze(sourceID string, bias, prompt, mentions string) (AnalysisResult, error) {
	transcriptPath := filepath.Join(r.LoopsDir, "drafts", sourceID, "transcript.txt")
	transcript, err := os.ReadFile(transcriptPath)
	if err != nil {
		return AnalysisResult{}, fmt.Errorf("read transcript: %w", err)
	}

	inputPath := filepath.Join(r.LoopsDir, "drafts", sourceID, "analyze-input.json")
	payload := map[string]string{
		"biases":     bias,
		"prompt":     prompt,
		"mentions":   mentions,
		"transcript": string(transcript),
	}
	raw, _ := json.Marshal(payload)
	if err := os.WriteFile(inputPath, raw, 0o644); err != nil {
		return AnalysisResult{}, err
	}

	out, err := r.Run(r.LoopsDir, "analyze.sh", "--input", inputPath, "--out", filepath.Join("drafts", sourceID, "analysis.json"))
	if err != nil {
		return AnalysisResult{}, err
	}

	resultPath := filepath.Join(r.LoopsDir, "drafts", sourceID, "analysis.json")
	if data, readErr := os.ReadFile(resultPath); readErr == nil {
		var result AnalysisResult
		if json.Unmarshal(data, &result) == nil && len(result.Candidates) > 0 {
			return result, nil
		}
	}

	// Fallback: parse stdout if file missing
	trim := strings.TrimSpace(out)
	start := strings.Index(trim, "{")
	end := strings.LastIndex(trim, "}")
	if start >= 0 && end > start {
		var result AnalysisResult
		if err := json.Unmarshal([]byte(trim[start:end+1]), &result); err == nil {
			return result, nil
		}
	}
	return AnalysisResult{}, fmt.Errorf("could not parse analysis output")
}

func speakerPodcast(source Source) (speaker, podcast string) {
	speaker = source.Title
	if speaker == "" {
		speaker = source.Podcast
	}
	if speaker == "" {
		speaker = "Speaker"
	}
	podcast = source.Podcast
	if podcast == "" {
		podcast = "Podcast"
	}
	return speaker, podcast
}

func (r Runner) WriteCandidateDraft(source Source, c Candidate, clipPath string, dict MentionDictionary) error {
	speaker, podcast := speakerPodcast(source)
	draftDir := filepath.Join(r.LoopsDir, "drafts", c.ID)
	if err := os.MkdirAll(draftDir, 0o755); err != nil {
		return err
	}

	postText := c.PostText
	if postText == "" {
		postText = fmt.Sprintf("%s: %s\n\n%s", speaker, c.Hook, c.Take)
	}
	postText = EnsurePostTextAttribution(postText, podcast, dict)

	meta := map[string]any{
		"id":          c.ID,
		"source_url":  source.YouTubeURL,
		"speaker":     speaker,
		"podcast":     podcast,
		"start":       c.StartTime,
		"end":         c.EndTime,
		"status":      "approved",
		"post_text":   postText,
		"clip_path":   clipPath,
		"created_at":  c.CreatedAt,
	}
	metaBytes, _ := json.MarshalIndent(meta, "", "  ")
	if err := os.WriteFile(filepath.Join(draftDir, "meta.json"), metaBytes, 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(draftDir, "post.txt"), []byte(postText), 0o644)
}

func (r Runner) ClipCandidate(source Source, c Candidate, dict MentionDictionary) (string, error) {
	_, err := r.Run(r.LoopsDir, "clip.sh",
		"--url", source.YouTubeURL,
		"--start", c.StartTime,
		"--end", c.EndTime,
		"--id", c.ID,
		"--out", "drafts")
	if err != nil {
		return "", err
	}

	clipPath := filepath.Join("drafts", c.ID, "clip.mp4")
	if err := r.WriteCandidateDraft(source, c, clipPath, dict); err != nil {
		return "", err
	}
	return clipPath, nil
}

func (r Runner) PreparePost(candidateID string) error {
	_, err := r.Run(r.LoopsDir, "prepare-post.sh", "--draft", candidateID)
	return err
}

func (r Runner) RewriteCandidate(bias, mentions string, source Source, c Candidate, instruction string) (RewriteResult, error) {
	workDir := filepath.Join(r.LoopsDir, "drafts", c.ID)
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return RewriteResult{}, err
	}

	podcast := source.Podcast
	if podcast == "" {
		podcast = sourcePodcastFromPostText(c.PostText)
	}
	if podcast == "" {
		podcast = "Podcast"
	}

	inputPath := filepath.Join(workDir, "rewrite-input.json")
	podcastHandle := ParseMentionDictionary(mentions).ResolvePodcastHandle(podcast)
	payload := map[string]string{
		"biases":         bias,
		"mentions":       mentions,
		"instruction":    instruction,
		"hook":           c.Hook,
		"take":           c.Take,
		"post_text":      c.PostText,
		"podcast":        podcast,
		"podcast_handle": podcastHandle,
	}
	if podcastHandle == "" {
		payload["podcast_handle"] = "theallinpod"
	}
	raw, _ := json.Marshal(payload)
	if err := os.WriteFile(inputPath, raw, 0o644); err != nil {
		return RewriteResult{}, err
	}

	outPath := filepath.Join(workDir, "rewrite-output.json")
	if _, err := r.Run(r.LoopsDir, "rewrite.sh", "--input", inputPath, "--out", outPath); err != nil {
		return RewriteResult{}, err
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		return RewriteResult{}, fmt.Errorf("read rewrite output: %w", err)
	}
	var result RewriteResult
	if err := json.Unmarshal(data, &result); err != nil {
		return RewriteResult{}, fmt.Errorf("parse rewrite output: %w", err)
	}
	result.PostText = EnsurePostTextAttribution(result.PostText, podcast, ParseMentionDictionary(mentions))
	return result, nil
}