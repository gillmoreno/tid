package factory

import "testing"

func TestValidateTimeRange(t *testing.T) {
	start, end, err := ValidateTimeRange("42:10", "47:30")
	if err != nil {
		t.Fatal(err)
	}
	if start != "00:42:10" || end != "00:47:30" {
		t.Fatalf("got %s → %s", start, end)
	}

	_, _, err = ValidateTimeRange("47:30", "42:10")
	if err == nil {
		t.Fatal("expected error when end before start")
	}
}

func TestClampClipToWindow(t *testing.T) {
	start, end, err := ClampClipToWindow("00:14:00", "00:20:00", "00:12:30", "00:22:30", MaxMomentClipSec)
	if err != nil {
		t.Fatal(err)
	}
	if start != "00:14:00" || end != "00:19:00" {
		t.Fatalf("got %s → %s, want 00:14:00 → 00:19:00", start, end)
	}

	start, end, err = ClampClipToWindow("00:10:00", "00:25:00", "00:12:30", "00:22:30", MaxMomentClipSec)
	if err != nil {
		t.Fatal(err)
	}
	if start != "00:12:30" || end != "00:17:30" {
		t.Fatalf("clamped to window: got %s → %s", start, end)
	}
}

func TestClampClipDuration(t *testing.T) {
	start, end, err := ClampClipDuration("00:14:00", "00:20:00", MaxMomentClipSec)
	if err != nil {
		t.Fatal(err)
	}
	if start != "00:14:00" || end != "00:19:00" {
		t.Fatalf("got %s → %s, want 00:14:00 → 00:19:00", start, end)
	}
}