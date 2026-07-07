package factory

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	vttTimestampLine = regexp.MustCompile(`^(\d{2}:\d{2}:\d{2}\.\d{3})\s*-->\s*(\d{2}:\d{2}:\d{2}\.\d{3})`)
	vttTagPattern    = regexp.MustCompile(`<[^>]+>`)
)

// ParseVTTTimestamp converts HH:MM:SS.mmm to seconds.
func ParseVTTTimestamp(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty timestamp")
	}
	if dot := strings.Index(s, "."); dot >= 0 {
		s = s[:dot]
	}
	return ParseTimestamp(s)
}

func findVTTFile(sourceDir string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(sourceDir, "*.en.vtt"))
	if err != nil {
		return "", err
	}
	if len(matches) > 0 {
		return matches[0], nil
	}
	matches, err = filepath.Glob(filepath.Join(sourceDir, "*.vtt"))
	if err != nil {
		return "", err
	}
	if len(matches) > 0 {
		return matches[0], nil
	}
	return "", fmt.Errorf("no VTT captions for source")
}

func cleanVTTText(raw string) string {
	text := vttTagPattern.ReplaceAllString(raw, "")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.TrimSpace(text)
	return text
}

func parseVTTCues(data []byte) ([]struct {
	start, end float64
	text      string
}, error) {
	lines := strings.Split(string(data), "\n")
	var cues []struct {
		start, end float64
		text      string
	}
	var currentStart, currentEnd float64
	var textLines []string

	flush := func() {
		if len(textLines) == 0 {
			return
		}
		text := cleanVTTText(strings.Join(textLines, " "))
		if text == "" {
			textLines = nil
			return
		}
		cues = append(cues, struct {
			start, end float64
			text      string
		}{currentStart, currentEnd, text})
		textLines = nil
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "WEBVTT") || strings.HasPrefix(line, "Kind:") || strings.HasPrefix(line, "Language:") || strings.HasPrefix(line, "NOTE") {
			continue
		}
		if m := vttTimestampLine.FindStringSubmatch(line); len(m) == 3 {
			flush()
			start, err := ParseVTTTimestamp(m[1])
			if err != nil {
				continue
			}
			end, err := ParseVTTTimestamp(m[2])
			if err != nil {
				continue
			}
			currentStart, currentEnd = start, end
			continue
		}
		if strings.Contains(line, "-->") {
			continue
		}
		if strings.HasPrefix(line, "align:") || strings.HasPrefix(line, "position:") {
			continue
		}
		textLines = append(textLines, line)
	}
	flush()
	return cues, nil
}

func ExtractTranscriptSegment(loopsDir, sourceID, startTime, endTime string) (string, error) {
	startSec, err := ParseTimestamp(startTime)
	if err != nil {
		return "", err
	}
	endSec, err := ParseTimestamp(endTime)
	if err != nil {
		return "", err
	}
	if endSec <= startSec {
		return "", fmt.Errorf("invalid time range")
	}

	sourceDir := filepath.Join(loopsDir, "drafts", sourceID)
	vttPath, err := findVTTFile(sourceDir)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(vttPath)
	if err != nil {
		return "", err
	}

	cues, err := parseVTTCues(data)
	if err != nil {
		return "", err
	}

	const minCueDuration = 0.15
	const bucketWindow = 2.5

	type bucket struct {
		start float64
		text  string
	}
	var buckets []bucket

	for _, cue := range cues {
		if cue.end <= startSec || cue.start >= endSec {
			continue
		}
		if cue.end-cue.start < minCueDuration {
			continue
		}
		text := cue.text
		if text == "" {
			continue
		}
		if len(buckets) == 0 || cue.start-buckets[len(buckets)-1].start > bucketWindow {
			buckets = append(buckets, bucket{cue.start, text})
			continue
		}
		last := buckets[len(buckets)-1]
		if len(text) > len(last.text) {
			buckets[len(buckets)-1].text = text
		}
	}

	if len(buckets) == 0 {
		return "", nil
	}

	combined := buckets[0].text
	for _, b := range buckets[1:] {
		combined = mergeCaptionOverlap(combined, b.text)
	}
	return combined, nil
}

func mergeCaptionOverlap(left, right string) string {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	if left == "" {
		return right
	}
	if right == "" || strings.Contains(left, right) {
		return left
	}
	if strings.HasPrefix(right, left) {
		return right
	}

	leftWords := strings.Fields(left)
	rightWords := strings.Fields(right)
	best := 0
	limit := len(leftWords)
	if len(rightWords) < limit {
		limit = len(rightWords)
	}
	for i := 1; i <= limit; i++ {
		suffix := strings.Join(leftWords[len(leftWords)-i:], " ")
		prefix := strings.Join(rightWords[:i], " ")
		if suffix == prefix {
			best = i
		}
	}
	if best > 0 {
		return left + " " + strings.Join(rightWords[best:], " ")
	}
	return left + " " + right
}