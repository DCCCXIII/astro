package output

import (
	"testing"

	"github.com/dcccxiii/astro/swisseph"
)

func TestJulianDayTimestamp(t *testing.T) {
	cases := []struct {
		name             string
		year, month, day int
		hour             float64
		want             string
	}{
		{"exact midnight", 2024, 3, 20, 12.0, "2024-03-20T12:00:00Z"},
		// A decimal hour whose seconds round up to 60 must roll over into
		// the next minute (and, here, the next day) via time.Date's
		// overflow normalization rather than producing an invalid time.
		{"seconds rollover", 2024, 1, 1, 23.0 + 59.0/60 + 59.9996/3600, "2024-01-02T00:00:00Z"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			jd := swisseph.JulDay(tc.year, tc.month, tc.day, tc.hour)
			got := julianDayTimestamp(jd)
			if got != tc.want {
				t.Errorf("julianDayTimestamp(%v) = %q, want %q", jd, got, tc.want)
			}
		})
	}
}
