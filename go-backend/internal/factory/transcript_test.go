package factory

import "testing"

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