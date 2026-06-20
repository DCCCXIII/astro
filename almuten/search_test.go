package almuten

import (
	"math"
	"testing"

	"github.com/dcccxiii/astro/swisseph"
)

// TestLastLongitude_SelfConsistent checks that searching backward for a
// longitude a planet is already known to occupy recovers that same instant
// (within the bisection tolerance), without relying on an external almanac.
func TestLastLongitude_SelfConsistent(t *testing.T) {
	ref := swisseph.JulDay(2024, 6, 1, 0)
	pos, _, err := swisseph.CalcPlanet(ref, swisseph.Mars)
	if err != nil {
		t.Fatalf("CalcPlanet: %v", err)
	}

	// Search from a few days after ref for the longitude Mars held at ref.
	crossingJD, _, _, err := LastLongitude(ref+5, swisseph.Mars, pos.Longitude, 10, searchStepDays)
	if err != nil {
		t.Fatalf("LastLongitude: %v", err)
	}

	gotPos, _, err := swisseph.CalcPlanet(crossingJD, swisseph.Mars)
	if err != nil {
		t.Fatalf("CalcPlanet at crossing: %v", err)
	}
	if diff := math.Abs(normalize180(gotPos.Longitude - pos.Longitude)); diff > 1e-3 {
		t.Errorf("longitude at crossing = %.6f, want %.6f (diff %.6f)", gotPos.Longitude, pos.Longitude, diff)
	}
	if math.Abs(crossingJD-ref) > 1.0/1440 { // within a minute
		t.Errorf("crossingJD = %.6f, want close to ref %.6f", crossingJD, ref)
	}
}

// TestLastLongitude_RetrogradeFlag checks that the retrograde flag returned
// always agrees with the sign of the speed at the crossing instant, by
// independently re-querying CalcPlanet.
func TestLastLongitude_RetrogradeFlag(t *testing.T) {
	jd := swisseph.JulDay(2024, 6, 1, 0)
	pos, _, err := swisseph.CalcPlanet(jd, swisseph.Mercury)
	if err != nil {
		t.Fatalf("CalcPlanet: %v", err)
	}

	crossingJD, speedLon, retrograde, err := LastLongitude(jd+3, swisseph.Mercury, pos.Longitude, 10, searchStepDays)
	if err != nil {
		t.Fatalf("LastLongitude: %v", err)
	}

	check, _, err := swisseph.CalcPlanet(crossingJD, swisseph.Mercury)
	if err != nil {
		t.Fatalf("CalcPlanet at crossing: %v", err)
	}
	if retrograde != (check.SpeedLon < 0) {
		t.Errorf("retrograde = %v, want %v (speed %.4f)", retrograde, check.SpeedLon < 0, check.SpeedLon)
	}
	if speedLon != check.SpeedLon {
		t.Errorf("speedLon = %.6f, want %.6f", speedLon, check.SpeedLon)
	}
}

// TestLastLongitude_NotFoundWithinWindow checks that a deliberately tiny
// window for a slow-moving planet returns an error rather than a false hit.
func TestLastLongitude_NotFoundWithinWindow(t *testing.T) {
	jd := swisseph.JulDay(2024, 6, 1, 0)
	pos, _, err := swisseph.CalcPlanet(jd, swisseph.Saturn)
	if err != nil {
		t.Fatalf("CalcPlanet: %v", err)
	}
	// Target a longitude far from Saturn's current position so a 1-day
	// window (much shorter than Saturn's ~0.034 deg/day motion could cover
	// to reach a longitude 90 degrees away) cannot find a crossing.
	target := normalize180(pos.Longitude + 90)

	_, _, _, err = LastLongitude(jd, swisseph.Saturn, target, 1, searchStepDays)
	if err == nil {
		t.Fatal("expected an error when no crossing exists within the window, got nil")
	}
}

// TestLastLongitude_MultipleCrossingsReturnsMostRecent validates that the
// backward scan stops at the chronologically most recent sign-change rather
// than an earlier one, for a target longitude Mercury is known to cross more
// than once within the window (its retrograde loop revisits longitudes it
// already passed through). The expected crossing is determined independently
// by brute-force fine sampling, not a hardcoded almanac date, and the test
// asserts at least two crossings actually exist in the window so it
// genuinely exercises the multiple-crossing case rather than degenerating
// into a single-crossing check.
func TestLastLongitude_MultipleCrossingsReturnsMostRecent(t *testing.T) {
	const window = 150.0 // > one Mercury synodic period (~116 days)
	const fineStep = 0.1

	// bruteForceCrossings independently fine-samples backward from ref over
	// window days and returns every sign change of the offset to target
	// (excluding the antipodal wraparound), most recent first.
	bruteForceCrossings := func(ref, target float64) ([]float64, error) {
		var crossings []float64
		prev, err := longitudeOffset(ref, swisseph.Mercury, target)
		if err != nil {
			return nil, err
		}
		tt := ref
		for i := 0; float64(i)*fineStep < window; i++ {
			tt -= fineStep
			cur, err := longitudeOffset(tt, swisseph.Mercury, target)
			if err != nil {
				return nil, err
			}
			if (prev <= 0) != (cur <= 0) && math.Abs(cur-prev) < 180 {
				crossings = append(crossings, tt+fineStep/2)
			}
			prev = cur
		}
		return crossings, nil
	}

	// Mercury retrogrades roughly 3 times a year; rather than assume a
	// specific calendar date contains a retrograde loop crossing a chosen
	// target twice, scan several candidate windows spread across two years
	// and use the first one that actually exhibits multiple crossings.
	var ref, target float64
	var crossings []float64
	found := false
	for month := 0; month < 24; month += 2 {
		candidateRef := swisseph.JulDay(2024, 1, 1, 0) + float64(month)*30
		basePos, _, err := swisseph.CalcPlanet(candidateRef-window, swisseph.Mercury)
		if err != nil {
			t.Fatalf("CalcPlanet: %v", err)
		}
		candidateTarget := basePos.Longitude
		c, err := bruteForceCrossings(candidateRef, candidateTarget)
		if err != nil {
			t.Fatalf("bruteForceCrossings: %v", err)
		}
		if len(c) >= 2 {
			ref, target, crossings, found = candidateRef, candidateTarget, c, true
			break
		}
	}
	if !found {
		t.Fatal("no candidate window in a 2-year scan exhibited multiple crossings; test setup invalid")
	}
	wantJD := crossings[0] // first found scanning backward = most recent

	gotJD, _, _, err := LastLongitude(ref, swisseph.Mercury, target, window, searchStepDays)
	if err != nil {
		t.Fatalf("LastLongitude: %v", err)
	}
	if math.Abs(gotJD-wantJD) > 1 { // within a day of the brute-force estimate
		t.Errorf("LastLongitude crossingJD = %.4f, brute-force most-recent crossing = %.4f (all crossings: %v)", gotJD, wantJD, crossings)
	}
}
