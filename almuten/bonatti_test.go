package almuten

import (
	"math"
	"testing"

	"github.com/dcccxiii/astro/swisseph"
)

// TestScoreBonattiKnown pins down the dispositor-chasing arithmetic against a
// hand-computed chart: every planet sits at 0° of its own domicile sign, and
// the four angles/vital points are placed to coincide with those positions.
// That makes every dispositor chase resolve to a sign/degree already in this
// table, so each planet's expected total is a closed-form sum of domicile (5)
// + triplicity (3) + term (2) + face (1) + exaltation (4, where the sign has
// one), weighted by how many of the 20 scoring places land on it.
func TestScoreBonattiKnown(t *testing.T) {
	c := Chart{
		IsDiurnal: true,
		Houses: swisseph.HouseResult{
			Ascendant: 0,  // Aries 0
			MC:        90, // Cancer 0
		},
		Positions: map[int]swisseph.PlanetPos{
			swisseph.Sun:     {Longitude: 120}, // Leo 0
			swisseph.Moon:    {Longitude: 90},  // Cancer 0
			swisseph.Mercury: {Longitude: 60},  // Gemini 0
			swisseph.Venus:   {Longitude: 180}, // Libra 0
			swisseph.Mars:    {Longitude: 0},   // Aries 0
			swisseph.Jupiter: {Longitude: 240}, // Sagittarius 0
			swisseph.Saturn:  {Longitude: 270}, // Capricorn 0
		},
		PartFortune: 60,  // Gemini 0
		Syzygy:      240, // Sagittarius 0
	}

	score := make(Scorecard, len(Planets))
	c.scoreBonatti(score)

	want := Scorecard{
		swisseph.Sun:     47,
		swisseph.Moon:    23,
		swisseph.Mercury: 23,
		swisseph.Venus:   40,
		swisseph.Mars:    32,
		swisseph.Jupiter: 52,
		swisseph.Saturn:  51,
	}
	for _, p := range Planets {
		if score[p] != want[p] {
			t.Errorf("%s score = %d, want %d", swisseph.PlanetName(p), score[p], want[p])
		}
	}
}

// --- integration: BuildChart + AlmutenBonatti -------------------------------

func TestAlmutenBonattiIntegration(t *testing.T) {
	jd := swisseph.JulDay(1990, 5, 15, 8.5)
	c, err := BuildChart(jd, 40.4, -3.7, swisseph.HousePlacidus)
	if err != nil {
		t.Fatalf("BuildChart: %v", err)
	}

	score, win, err := AlmutenBonatti(c)
	if err != nil {
		t.Fatalf("AlmutenBonatti: %v", err)
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
			t.Errorf("winner %s has score %d, want max %d", swisseph.PlanetName(w), score[w], max)
		}
	}
}
