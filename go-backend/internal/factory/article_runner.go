package factory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FetchArticle extracts readable text from an article URL into drafts/{id}/article.txt.
// Returns the extracted title when available.
func (r Runner) FetchArticle(sourceID, articleURL string) (string, error) {
	workDir := filepath.Join(r.ArticleLoopsDir, "drafts", sourceID)
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return "", err
	}
	out, err := runIn(r.ArticleLoopsDir, "fetch-article.sh", "--url", articleURL, "--id", sourceID, "--out", "drafts")
	if err != nil {
		return "", err
	}

	// fetch-article.sh writes article.json {title, text}; prefer that for the title.
	metaPath := filepath.Join(workDir, "article.json")
	if data, readErr := os.ReadFile(metaPath); readErr == nil {
		var meta struct {
			Title string `json:"title"`
		}
		if json.Unmarshal(data, &meta) == nil {
			return strings.TrimSpace(meta.Title), nil
		}
	}
	return strings.TrimSpace(out), nil
}

func (r Runner) articleText(sourceID string) (string, error) {
	path := filepath.Join(r.ArticleLoopsDir, "drafts", sourceID, "article.txt")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("no extracted article text for %s: %w", sourceID, err)
	}
	text := strings.TrimSpace(string(data))
	if text == "" {
		return "", fmt.Errorf("extracted article for %s is empty", sourceID)
	}
	return text, nil
}

// AnalyzeArticle runs the article analyzer over the extracted text.
func (r Runner) AnalyzeArticle(sourceID, publication string, bias, prompt, mentions string) (ArticleAnalysisResult, error) {
	article, err := r.articleText(sourceID)
	if err != nil {
		return ArticleAnalysisResult{}, err
	}

	workDir := filepath.Join(r.ArticleLoopsDir, "drafts", sourceID)
	inputPath := filepath.Join(workDir, "analyze-input.json")
	payload := map[string]string{
		"biases":      bias,
		"prompt":      prompt,
		"mentions":    mentions,
		"publication": publication,
		"article":     article,
	}
	raw, _ := json.Marshal(payload)
	if err := os.WriteFile(inputPath, raw, 0o644); err != nil {
		return ArticleAnalysisResult{}, err
	}

	outRel := filepath.Join("drafts", sourceID, "analysis.json")
	out, err := runIn(r.ArticleLoopsDir, "analyze.sh", "--input", inputPath, "--out", outRel)
	if err != nil {
		return ArticleAnalysisResult{}, err
	}

	resultPath := filepath.Join(workDir, "analysis.json")
	if data, readErr := os.ReadFile(resultPath); readErr == nil {
		var result ArticleAnalysisResult
		if json.Unmarshal(data, &result) == nil && len(result.Candidates) > 0 {
			return result, nil
		}
	}

	trim := strings.TrimSpace(out)
	start := strings.Index(trim, "{")
	end := strings.LastIndex(trim, "}")
	if start >= 0 && end > start {
		var result ArticleAnalysisResult
		if err := json.Unmarshal([]byte(trim[start:end+1]), &result); err == nil && len(result.Candidates) > 0 {
			return result, nil
		}
	}
	return ArticleAnalysisResult{}, fmt.Errorf("could not parse article analysis output")
}

// RewriteArticlePost applies Gil's lens + an instruction to an article post.
func (r Runner) RewriteArticlePost(bias, mentions string, src ArticleSource, c ArticleCandidate, instruction string) (RewriteResult, error) {
	workDir := filepath.Join(r.ArticleLoopsDir, "drafts", c.ID)
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return RewriteResult{}, err
	}

	publication := strings.TrimSpace(src.Publication)
	if publication == "" {
		publication = "Source"
	}
	pubHandle := ParseMentionDictionary(mentions).ResolvePublicationHandle(publication)

	inputPath := filepath.Join(workDir, "rewrite-input.json")
	payload := map[string]string{
		"biases":             bias,
		"mentions":           mentions,
		"instruction":        instruction,
		"post_text":          c.PostText,
		"publication":        publication,
		"publication_handle": pubHandle,
	}
	raw, _ := json.Marshal(payload)
	if err := os.WriteFile(inputPath, raw, 0o644); err != nil {
		return RewriteResult{}, err
	}

	outPath := filepath.Join(workDir, "rewrite-output.json")
	if _, err := runIn(r.ArticleLoopsDir, "rewrite.sh", "--input", inputPath, "--out", outPath); err != nil {
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
	result.PostText = EnsureArticlePostAttribution(result.PostText, publication, ParseMentionDictionary(mentions))
	return result, nil
}

// WriteArticleDraft writes meta.json + post.txt for an article candidate.
func (r Runner) WriteArticleDraft(src ArticleSource, c ArticleCandidate, dict MentionDictionary) error {
	draftDir := filepath.Join(r.ArticleLoopsDir, "drafts", c.ID)
	if err := os.MkdirAll(draftDir, 0o755); err != nil {
		return err
	}

	publication := strings.TrimSpace(src.Publication)
	if publication == "" {
		publication = "Source"
	}
	postText := EnsureArticlePostAttribution(c.PostText, publication, dict)

	meta := map[string]any{
		"id":          c.ID,
		"source_url":  src.URL,
		"publication": publication,
		"title":       src.Title,
		"status":      "approved",
		"post_text":   postText,
		"created_at":  c.CreatedAt,
	}
	metaBytes, _ := json.MarshalIndent(meta, "", "  ")
	if err := os.WriteFile(filepath.Join(draftDir, "meta.json"), metaBytes, 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(draftDir, "post.txt"), []byte(postText), 0o644)
}

// PrepareArticlePost runs the semi-automated posting prep (clipboard + browser).
func (r Runner) PrepareArticlePost(candidateID string) error {
	_, err := runIn(r.ArticleLoopsDir, "prepare-post.sh", "--draft", candidateID)
	return err
}
