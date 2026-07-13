package factory

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseTimestamp converts HH:MM:SS or MM:SS to seconds.
func ParseTimestamp(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty timestamp")
	}
	parts := strings.Split(s, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return 0, fmt.Errorf("invalid timestamp %q", s)
	}

	var hours, minutes, seconds float64
	var err error

	if len(parts) == 3 {
		hours, err = strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return 0, fmt.Errorf("invalid timestamp %q", s)
		}
		minutes, err = strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return 0, fmt.Errorf("invalid timestamp %q", s)
		}
		seconds, err = strconv.ParseFloat(parts[2], 64)
		if err != nil {
			return 0, fmt.Errorf("invalid timestamp %q", s)
		}
	} else {
		minutes, err = strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return 0, fmt.Errorf("invalid timestamp %q", s)
		}
		seconds, err = strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return 0, fmt.Errorf("invalid timestamp %q", s)
		}
	}

	total := hours*3600 + minutes*60 + seconds
	if total < 0 {
		return 0, fmt.Errorf("invalid timestamp %q", s)
	}
	return total, nil
}

const (
	MaxMomentClipSec   = 300  // proposed clip max duration: 5 min
	MaxMomentWindowSec = 1200 // transcript window max: 20 min
)

// ValidateTimeRange normalizes start/end and ensures end is after start within limits.
func ValidateTimeRange(startTime, endTime string) (start, end string, err error) {
	startSec, err := ParseTimestamp(startTime)
	if err != nil {
		return "", "", fmt.Errorf("invalid start_time: %w", err)
	}
	endSec, err := ParseTimestamp(endTime)
	if err != nil {
		return "", "", fmt.Errorf("invalid end_time: %w", err)
	}
	if endSec <= startSec {
		return "", "", fmt.Errorf("end_time must be after start_time")
	}
	if endSec-startSec > MaxMomentWindowSec {
		return "", "", fmt.Errorf("range too long (max %d minutes)", int(MaxMomentWindowSec/60))
	}
	return FormatTimestamp(startSec), FormatTimestamp(endSec), nil
}

// ClampClipToWindow keeps clip timestamps inside the user range and max duration.
func ClampClipToWindow(clipStart, clipEnd, windowStart, windowEnd string, maxSec float64) (string, string, error) {
	ws, err := ParseTimestamp(windowStart)
	if err != nil {
		return "", "", err
	}
	we, err := ParseTimestamp(windowEnd)
	if err != nil {
		return "", "", err
	}
	cs, err := ParseTimestamp(clipStart)
	if err != nil {
		cs = ws
	}
	ce, err := ParseTimestamp(clipEnd)
	if err != nil {
		ce = we
	}
	if cs < ws {
		cs = ws
	}
	if ce > we {
		ce = we
	}
	if ce <= cs {
		cs, ce = ws, we
	}
	return ClampClipDuration(FormatTimestamp(cs), FormatTimestamp(ce), maxSec)
}

// ClampClipDuration shortens end_time so the clip is at most maxSec long.
func ClampClipDuration(startTime, endTime string, maxSec float64) (string, string, error) {
	start, err := ParseTimestamp(startTime)
	if err != nil {
		return "", "", err
	}
	end, err := ParseTimestamp(endTime)
	if err != nil {
		return "", "", err
	}
	if end <= start {
		return "", "", fmt.Errorf("end_time %q must be after start_time %q", endTime, startTime)
	}
	if end-start > maxSec {
		end = start + maxSec
	}
	return FormatTimestamp(start), FormatTimestamp(end), nil
}

func FormatTimestamp(seconds float64) string {
	if seconds < 0 {
		seconds = 0
	}
	h := int(seconds) / 3600
	m := (int(seconds) % 3600) / 60
	s := int(seconds) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
