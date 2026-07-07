package factory

import (
	"strings"
	"testing"
)

func TestExtractTranscriptSegment(t *testing.T) {
	loopsDir := "../../../loops/clip-to-post"
	text, err := ExtractTranscriptSegment(loopsDir, "20260707-silicon-valley-girl", "00:06:45", "00:10:15")
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if text == "" {
		t.Fatal("expected non-empty transcript segment")
	}
	if len(text) < 100 {
		t.Fatalf("transcript too short: %q", text)
	}
	t.Logf("sample: %.200s...", text)
}

func TestExtractTimestampedTranscriptSegment(t *testing.T) {
	loopsDir := "../../../loops/clip-to-post"
	text, err := ExtractTimestampedTranscriptSegment(loopsDir, "20260707-silicon-valley-girl", "00:02:10", "00:05:35")
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if !strings.Contains(text, "[00:03:") {
		t.Fatalf("expected timestamped lines, got: %.200s", text)
	}
	if !strings.Contains(strings.ToLower(text), "intel") {
		t.Fatal("expected Intel in c01 segment")
	}
}

func TestExportTimestampedTranscript(t *testing.T) {
	loopsDir := "../../../loops/clip-to-post"
	text, err := ExportTimestampedTranscript(loopsDir, "20260707-silicon-valley-girl")
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if !strings.Contains(text, "[00:03:") {
		t.Fatalf("expected timestamped lines around 00:03, got: %.200s", text)
	}
	if !strings.Contains(strings.ToLower(text), "intel") {
		t.Fatal("expected Intel mention with timestamp in exported transcript")
	}
}