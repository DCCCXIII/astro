package almuten

import (
	"math"
	"testing"

	"github.com/dcccxiii/astro/swisseph"
)

func TestWeekday(t *testing.T) {
	cases := []struct {
		jd   float64
		want int // 0=Sunday … 6=Saturday
	}{
		{swisseph.JulDay(2000, 1, 1, 12.0), 6}, // Saturday
		{swisseph.JulDay(2000, 1, 2, 12.0), 0}, // Sunday
		{swisseph.JulDay(1990, 5, 15, 8.5), 2}, // Tuesday
	}
	for _, c := range cases {
		if got := weekday(c.jd); got != c.want {
			t.Errorf("weekday(%.2f) = %d, want %d", c.jd, got, c.want)
		}
	}
}

func TestAdjustedHouse(t *testing.T) {
	var cusps [13]float64
	for h := 1; h <= 12; h++ {
		cusps[h] = float64((h - 1) * 30) // whole-sign cusps
	}
	cases := []struct {
		lon  float64
		want int
	}{
		{5, 1},   // well inside house 1
		{24, 1},  // 6° before cusp 2 -> still house 1
		{26, 2},  // 4° before cusp 2 -> promoted to house 2
		{27, 2},  // within 5° pre-cusp orb
		{359, 1}, // 1° before cusp 1 (wrap) -> house 1
	}
	for _, c := range cases {
		if got := adjustedHouse(c.lon, cusps); got != c.want {
			t.Errorf("adjustedHouse(%.0f) = %d, want %d", c.lon, got, c.want)
		}
	}
}

// TestAddEssentialDignities checks the five dignity weights for Aries 5° (diurnal).
func TestAddEssentialDignities(t *testing.T) {
	score := make(Scorecard)
	addEssentialDignities(score, 5, true) // Aries 5°, diurnal

	// Aries: domicile Mars(+5), exalt Sun(+4), fire-day triplicity Sun(+3),
	// term Jupiter(+2), face Mars(+1).
	want := map[int]int{
		swisseph.Mars:    6, // domicile + face
		swisseph.Sun:     7, // exaltation + triplicity
		swisseph.Jupiter: 2, // term
	}
	for p, w := range want {
		if score[p] != w {
			t.Errorf("score[%s] = %d, want %d", swisseph.PlanetName(p), score[p], w)
		}
	}
}

// TestPrenatalSyzygy checks the lunation immediately before a known date. The
// full moon of 1990-05-09 precedes the 1990-05-15 birth (next new moon is
// 1990-05-24), so the prenatal syzygy is a full moon.
func TestPrenatalSyzygy(t *testing.T) {
	jd := swisseph.JulDay(1990, 5, 15, 8.5)
	lon, kind, err := PrenatalSyzygy(jd, 40.4, -3.7)
	if err != nil {
		t.Fatalf("PrenatalSyzygy: %v", err)
	}
	if kind != "full" {
		t.Errorf("syzygy type = %q, want \"full\"", kind)
	}
	if lon < 0 || lon >= 360 {
		t.Errorf("syzygy longitude %.4f out of [0, 360)", lon)
	}
}

func TestAlmutenFigurisIntegration(t *testing.T) {
	jd := swisseph.JulDay(1990, 5, 15, 8.5)
	c, err := BuildChart(jd, 40.4, -3.7, swisseph.HousePlacidus)
	if err != nil {
		t.Fatalf("BuildChart: %v", err)
	}

	score, win, err := AlmutenFiguris(c)
	if err != nil {
		t.Fatalf("AlmutenFiguris: %v", err)
	}
	if len(score) != len(Planets) {
		t.Fatalf("scorecard has %d planets, want %d", len(score), len(Planets))
	}
	if len(win) == 0 {
		t.Fatal("no winner returned")
	}

	max := math.MinInt
	for _, p := range Planets {
		if score[p] > max {
			max = score[p]
		}
	}
	for _, w := range win {
		if score[w] != max {
			t.Errorf("winner %s score %d != max %d", swisseph.PlanetName(w), score[w], max)
		}
	}

	// Determinism.
	score2, _, _ := AlmutenFiguris(c)
	for _, p := range Planets {
		if score[p] != score2[p] {
			t.Errorf("non-deterministic score for %s: %d vs %d", swisseph.PlanetName(p), score[p], score2[p])
		}
	}
}
