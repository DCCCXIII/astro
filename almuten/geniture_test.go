package almuten

import (
	"math"
	"os"
	"testing"

	"github.com/dcccxiii/astro/swisseph"
)

// TestMain sets the ephemeris path so the cgo-backed calls (CalcPlanet,
// CalcHouses, FixStar) work, mirroring the swisseph package's test setup.
func TestMain(m *testing.M) {
	swisseph.SetEphePath("../ephe")
	code := m.Run()
	swisseph.Close()
	os.Exit(code)
}

// --- pure geometric helpers ------------------------------------------------

func TestNormalize180(t *testing.T) {
	cases := []struct{ in, want float64 }{
		{0, 0}, {180, 180}, {-180, 180}, {190, -170}, {350, -10}, {370, 10},
	}
	for _, c := range cases {
		if got := normalize180(c.in); math.Abs(got-c.want) > 1e-9 {
			t.Errorf("normalize180(%.1f) = %.4f, want %.4f", c.in, got, c.want)
		}
	}
}

func TestElongation(t *testing.T) {
	cases := []struct{ a, b, want float64 }{
		{10, 10, 0}, {0, 180, 180}, {350, 10, 20}, {100, 5, 95},
	}
	for _, c := range cases {
		if got := elongation(c.a, c.b); math.Abs(got-c.want) > 1e-9 {
			t.Errorf("elongation(%.1f,%.1f) = %.4f, want %.4f", c.a, c.b, got, c.want)
		}
	}
}

func TestPartileAspect(t *testing.T) {
	cases := []struct {
		a, b   float64
		aspect int
		ok     bool
	}{
		{0, 0.5, 0, true},   // partile conjunction
		{0, 60.9, 60, true}, // partile sextile
		{0, 122, 0, false},  // 2° off trine -> not partile
		{0, 180, 180, true}, // partile opposition
		{0, 89.5, 90, true}, // partile square
	}
	for _, c := range cases {
		asp, ok := partileAspect(c.a, c.b)
		if ok != c.ok || (ok && asp != c.aspect) {
			t.Errorf("partileAspect(%.1f,%.1f) = (%d,%v), want (%d,%v)", c.a, c.b, asp, ok, c.aspect, c.ok)
		}
	}
}

// TestHouseOfWholeSign uses 30°-wide cusps so each house equals one sign.
func TestHouseOfWholeSign(t *testing.T) {
	var cusps [13]float64
	for h := 1; h <= 12; h++ {
		cusps[h] = float64((h - 1) * 30)
	}
	cases := []struct {
		lon  float64
		want int
	}{
		{5, 1}, {35, 2}, {95, 4}, {100, 4}, {355, 12},
	}
	for _, c := range cases {
		if got := HouseOf(c.lon, cusps); got != c.want {
			t.Errorf("HouseOf(%.1f) = %d, want %d", c.lon, got, c.want)
		}
	}
}

// --- mutual reception (pure) ----------------------------------------------

func TestMutualReceptionDomicile(t *testing.T) {
	// Sun in Aries (Mars' domicile) and Mars in Leo (Sun's domicile) -> MR.
	c := Chart{Positions: map[int]swisseph.PlanetPos{
		swisseph.Sun:  {Longitude: 5},   // Aries
		swisseph.Mars: {Longitude: 125}, // Leo
	}}
	if !c.mutualReceptionDomicile(swisseph.Sun) {
		t.Error("expected Sun-Mars mutual reception by domicile")
	}
	if !c.mutualReceptionDomicile(swisseph.Mars) {
		t.Error("expected Mars-Sun mutual reception by domicile")
	}
}

// --- besiegement (pure) ---------------------------------------------------

func TestBesiegedBySaturnMars(t *testing.T) {
	// "far" positions keep non-relevant planets off the Saturn–Mars arc.
	base := map[int]float64{
		swisseph.Sun: 100, swisseph.Moon: 200, swisseph.Venus: 300, swisseph.Jupiter: 170,
	}
	withBase := func(extra map[int]float64) map[int]swisseph.PlanetPos {
		pos := map[int]swisseph.PlanetPos{}
		for p, l := range base {
			pos[p] = swisseph.PlanetPos{Longitude: l}
		}
		for p, l := range extra {
			pos[p] = swisseph.PlanetPos{Longitude: l}
		}
		return pos
	}

	cases := []struct {
		name   string
		extra  map[int]float64
		planet int
		want   bool
	}{
		{"enclosed on minor arc", map[int]float64{swisseph.Saturn: 10, swisseph.Mars: 25, swisseph.Mercury: 15}, swisseph.Mercury, true},
		{"intervening planet breaks siege", map[int]float64{swisseph.Saturn: 10, swisseph.Mars: 25, swisseph.Mercury: 15, swisseph.Venus: 18}, swisseph.Mercury, false},
		{"candidate outside arc", map[int]float64{swisseph.Saturn: 10, swisseph.Mars: 25, swisseph.Mercury: 40}, swisseph.Mercury, false},
		{"arc too wide (>30°)", map[int]float64{swisseph.Saturn: 10, swisseph.Mars: 50, swisseph.Mercury: 30}, swisseph.Mercury, false},
		{"a malefic itself is never besieged", map[int]float64{swisseph.Saturn: 10, swisseph.Mars: 25, swisseph.Mercury: 15}, swisseph.Saturn, false},
		{"wrap across 0°", map[int]float64{swisseph.Saturn: 350, swisseph.Mars: 5, swisseph.Mercury: 0}, swisseph.Mercury, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			chart := Chart{Positions: withBase(c.extra)}
			if got := chart.besiegedBySaturnMars(c.planet); got != c.want {
				t.Errorf("besiegedBySaturnMars(%s) = %v, want %v", swisseph.PlanetName(c.planet), got, c.want)
			}
		})
	}
}

// --- golden scorecard for a single hand-built planet ----------------------

// TestScoreMarsKnown checks the full per-planet branch sum for Mars placed at
// Aries 5° in a controlled chart, with fixed stars disabled (orb 0).
func TestScoreMarsKnown(t *testing.T) {
	var cusps [13]float64
	for h := 1; h <= 12; h++ {
		cusps[h] = float64((h - 1) * 30) // whole-sign cusps
	}
	c := Chart{
		JD:        swisseph.JulDay(2000, 1, 1, 12.0),
		IsDiurnal: true,
		Houses:    swisseph.HouseResult{Cusps: cusps},
		NorthNode: 80,
		Positions: map[int]swisseph.PlanetPos{
			swisseph.Sun:     {Longitude: 100},
			swisseph.Moon:    {Longitude: 200},
			swisseph.Mercury: {Longitude: 250},
			swisseph.Venus:   {Longitude: 310},
			swisseph.Mars:    {Longitude: 5, SpeedLon: 0.6}, // Aries 5°, direct & swift
			swisseph.Jupiter: {Longitude: 170},
			swisseph.Saturn:  {Longitude: 230},
		},
	}
	opts := DefaultOptions()
	opts.FixedStarOrb = 0 // suppress star bonuses for a deterministic total

	// Expected for Mars: domicile +5, face +1, direct +4, swift +2,
	// free of Sun +5, oriental +2, house 1 +5 = 24.
	if got := c.scoreGeniturePlanet(swisseph.Mars, 100, opts); got != 24 {
		t.Errorf("Mars score = %d, want 24", got)
	}
}

// --- integration: BuildChart + LordOfGeniture -----------------------------

func TestLordOfGenitureIntegration(t *testing.T) {
	jd := swisseph.JulDay(1990, 5, 15, 8.5)
	c, err := BuildChart(jd, 40.4, -3.7, swisseph.HousePlacidus)
	if err != nil {
		t.Fatalf("BuildChart: %v", err)
	}

	score, win := LordOfGeniture(c, DefaultOptions())
	if len(score) != len(Planets) {
		t.Fatalf("scorecard has %d planets, want %d", len(score), len(Planets))
	}
	if len(win) == 0 {
		t.Fatal("no winner returned")
	}

	// The winner(s) must hold the maximum score.
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

	// Determinism: same inputs -> same scorecard.
	score2, _ := LordOfGeniture(c, DefaultOptions())
	for _, p := range Planets {
		if score[p] != score2[p] {
			t.Errorf("non-deterministic score for %s: %d vs %d", swisseph.PlanetName(p), score[p], score2[p])
		}
	}
}
