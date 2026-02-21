package swisseph_test

import (
	"math"
	"os"
	"testing"

	"github.com/dcccxiii/astro/swisseph"
)

// TestMain sets the ephemeris path before any test runs and closes the
// library afterwards. The path points one level up from the package
// directory, where the ephe/ directory lives in the repo. If the files
// are absent the library falls back to the built-in Moshier approximation
// (lower precision but still correct to well within the tolerances used below).
func TestMain(m *testing.M) {
	swisseph.SetEphePath("../ephe")
	code := m.Run()
	swisseph.Close()
	os.Exit(code)
}

// ---------------------------------------------------------------------------
// ZodiacSign
// ---------------------------------------------------------------------------

func TestZodiacSign(t *testing.T) {
	cases := []struct {
		lon     float64
		sign    string
		degrees float64
	}{
		// Sign boundaries (each sign starts at a multiple of 30°)
		{0.0, "Aries", 0.0},
		{30.0, "Taurus", 0.0},
		{60.0, "Gemini", 0.0},
		{90.0, "Cancer", 0.0},
		{120.0, "Leo", 0.0},
		{150.0, "Virgo", 0.0},
		{180.0, "Libra", 0.0},
		{210.0, "Scorpio", 0.0},
		{240.0, "Sagittarius", 0.0},
		{270.0, "Capricorn", 0.0},
		{300.0, "Aquarius", 0.0},
		{330.0, "Pisces", 0.0},
		// Interior of a sign
		{15.0, "Aries", 15.0},
		{45.5, "Taurus", 15.5},
		// Just before the next sign boundary
		{29.999, "Aries", 29.999},
		{359.9, "Pisces", 29.9},
		// Exactly 360° — idx clamps to 11 (Pisces), degree = 30.0
		{360.0, "Pisces", 30.0},
	}

	for _, tc := range cases {
		sign, deg := swisseph.ZodiacSign(tc.lon)
		if sign != tc.sign {
			t.Errorf("ZodiacSign(%.3f) sign = %q, want %q", tc.lon, sign, tc.sign)
		}
		if math.Abs(deg-tc.degrees) > 1e-9 {
			t.Errorf("ZodiacSign(%.3f) degrees = %v, want %v", tc.lon, deg, tc.degrees)
		}
	}
}

// ---------------------------------------------------------------------------
// JulDay
// ---------------------------------------------------------------------------

func TestJulDay_KnownEpochs(t *testing.T) {
	const epsilon = 1e-5 // well under a second of time

	cases := []struct {
		name        string
		year, month, day int
		hour        float64
		want        float64
	}{
		// J2000.0 is defined as JD 2451545.0 at 2000-01-01 12:00 UT.
		{"J2000.0", 2000, 1, 1, 12.0, 2451545.0},
		// Unix epoch: 1970-01-01 00:00 UTC = JD 2440587.5
		// (J2000.0 minus 10957.5 days: 30 years with 7 leap years + 0.5-day offset)
		{"Unix epoch", 1970, 1, 1, 0.0, 2440587.5},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := swisseph.JulDay(tc.year, tc.month, tc.day, tc.hour)
			if math.Abs(got-tc.want) > epsilon {
				t.Errorf("JulDay = %.6f, want %.6f (diff %.2e)", got, tc.want, math.Abs(got-tc.want))
			}
		})
	}
}

// ---------------------------------------------------------------------------
// PlanetName
// ---------------------------------------------------------------------------

func TestPlanetName(t *testing.T) {
	cases := []struct {
		id   int
		name string
	}{
		{swisseph.Sun, "Sun"},
		{swisseph.Moon, "Moon"},
		{swisseph.Mercury, "Mercury"},
		{swisseph.Venus, "Venus"},
		{swisseph.Mars, "Mars"},
		{swisseph.Jupiter, "Jupiter"},
		{swisseph.Saturn, "Saturn"},
	}

	for _, tc := range cases {
		got := swisseph.PlanetName(tc.id)
		if got != tc.name {
			t.Errorf("PlanetName(%d) = %q, want %q", tc.id, got, tc.name)
		}
	}
}

// ---------------------------------------------------------------------------
// CalcPlanet
// ---------------------------------------------------------------------------

// TestCalcPlanet_J2000 checks the Sun's position against the well-known
// J2000.0 reference epoch (2000-01-01 12:00 UT, JD 2451545.0).
// Sun ecliptic longitude ≈ 280.46° (Capricorn ~10.46°), daily speed ≈ 1.0°/day.
// A tolerance of 1° accommodates the Moshier fallback when ephemeris files are absent.
func TestCalcPlanet_J2000(t *testing.T) {
	jd := swisseph.JulDay(2000, 1, 1, 12.0)

	pos, err := swisseph.CalcPlanet(jd, swisseph.Sun)
	if err != nil {
		t.Fatalf("CalcPlanet(Sun) unexpected error: %v", err)
	}

	const (
		wantLon    = 280.46
		lonTol     = 1.0  // degrees; Moshier precision
		wantSpeed  = 1.0  // degrees/day
		speedTol   = 0.05 // degrees/day
	)

	if math.Abs(pos.Longitude-wantLon) > lonTol {
		t.Errorf("Sun longitude = %.4f°, want %.4f° ± %.1f°", pos.Longitude, wantLon, lonTol)
	}
	if math.Abs(pos.SpeedLon-wantSpeed) > speedTol {
		t.Errorf("Sun speed = %.4f°/day, want %.4f ± %.2f°/day", pos.SpeedLon, wantSpeed, speedTol)
	}

	sign, _ := swisseph.ZodiacSign(pos.Longitude)
	if sign != "Capricorn" {
		t.Errorf("Sun sign at J2000.0 = %q, want Capricorn", sign)
	}
}

// TestCalcPlanet_AllPlanets verifies that all seven classical planets return
// a valid position (no error, longitude in [0, 360)) at J2000.0.
func TestCalcPlanet_AllPlanets(t *testing.T) {
	jd := swisseph.JulDay(2000, 1, 1, 12.0)

	planets := []struct {
		id   int
		name string
	}{
		{swisseph.Sun, "Sun"},
		{swisseph.Moon, "Moon"},
		{swisseph.Mercury, "Mercury"},
		{swisseph.Venus, "Venus"},
		{swisseph.Mars, "Mars"},
		{swisseph.Jupiter, "Jupiter"},
		{swisseph.Saturn, "Saturn"},
	}

	for _, p := range planets {
		t.Run(p.name, func(t *testing.T) {
			pos, err := swisseph.CalcPlanet(jd, p.id)
			if err != nil {
				t.Fatalf("CalcPlanet(%s) error: %v", p.name, err)
			}
			if pos.Longitude < 0 || pos.Longitude >= 360 {
				t.Errorf("CalcPlanet(%s) longitude = %.4f°, want [0, 360)", p.name, pos.Longitude)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// CalcHouses
// ---------------------------------------------------------------------------

// TestCalcHouses_ValidRanges checks that all six supported house systems
// return angles in the valid [0, 360) range for a representative location
// and time (London, J2000.0).
func TestCalcHouses_ValidRanges(t *testing.T) {
	jd := swisseph.JulDay(2000, 1, 1, 12.0)
	lat, lon := 51.5074, -0.1278 // London

	systems := []struct {
		label string
		code  byte
	}{
		{"Placidus", swisseph.HousePlacidus},
		{"Koch", swisseph.HouseKoch},
		{"WholeSign", swisseph.HouseWholeSign},
		{"Regiomontanus", swisseph.HouseRegiomontanus},
		{"Equal", swisseph.HouseEqual},
		{"Campanus", swisseph.HouseCampanus},
	}

	for _, sys := range systems {
		t.Run(sys.label, func(t *testing.T) {
			res, err := swisseph.CalcHouses(jd, lat, lon, sys.code)
			if err != nil {
				t.Fatalf("CalcHouses error: %v", err)
			}

			inRange := func(v float64) bool { return v >= 0 && v < 360 }

			if !inRange(res.Ascendant) {
				t.Errorf("Ascendant = %.4f° out of [0, 360)", res.Ascendant)
			}
			if !inRange(res.MC) {
				t.Errorf("MC = %.4f° out of [0, 360)", res.MC)
			}
			for i := 1; i <= 12; i++ {
				if !inRange(res.Cusps[i]) {
					t.Errorf("Cusps[%d] = %.4f° out of [0, 360)", i, res.Cusps[i])
				}
			}
		})
	}
}

// TestCalcHouses_WholeSign verifies the defining property of Whole Sign houses:
// each successive cusp is exactly 30° ahead of the previous one (mod 360).
func TestCalcHouses_WholeSign(t *testing.T) {
	jd := swisseph.JulDay(2000, 1, 1, 12.0)

	res, err := swisseph.CalcHouses(jd, 51.5074, -0.1278, swisseph.HouseWholeSign)
	if err != nil {
		t.Fatalf("CalcHouses(WholeSign) error: %v", err)
	}

	for i := 2; i <= 12; i++ {
		want := math.Mod(res.Cusps[i-1]+30.0, 360.0)
		if math.Abs(res.Cusps[i]-want) > 1e-6 {
			t.Errorf("Cusps[%d] = %.6f°, want %.6f° (Cusps[%d] + 30° mod 360)", i, res.Cusps[i], want, i-1)
		}
	}
}

// TestCalcHouses_ASCMatchesCusp1 checks the Ascendant matches Cusps[1],
// which holds for every house system except (arguably) Whole Sign.
// We test it for Placidus as the canonical case.
func TestCalcHouses_ASCMatchesCusp1(t *testing.T) {
	jd := swisseph.JulDay(2000, 1, 1, 12.0)

	res, err := swisseph.CalcHouses(jd, 51.5074, -0.1278, swisseph.HousePlacidus)
	if err != nil {
		t.Fatalf("CalcHouses(Placidus) error: %v", err)
	}

	if math.Abs(res.Ascendant-res.Cusps[1]) > 1e-6 {
		t.Errorf("Ascendant (%.6f°) does not match Cusps[1] (%.6f°)", res.Ascendant, res.Cusps[1])
	}
}
