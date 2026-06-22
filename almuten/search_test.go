package almuten

import (
	"math"
	"testing"

	"github.com/dcccxiii/astro/swisseph"
)

// TestFindLongitude_Backward_SelfConsistent checks that searching backward
// for a longitude a planet is already known to occupy recovers that same
// instant (within the bisection tolerance), without relying on an external
// almanac.
func TestFindLongitude_Backward_SelfConsistent(t *testing.T) {
	ref := swisseph.JulDay(2024, 6, 1, 0)
	pos, _, err := swisseph.CalcPlanet(ref, swisseph.Mars)
	if err != nil {
		t.Fatalf("CalcPlanet: %v", err)
	}

	// Search from a few days after ref for the longitude Mars held at ref.
	crossingJD, _, _, err := FindLongitude(ref+5, swisseph.Mars, pos.Longitude, 10, searchStepDays, Backward)
	if err != nil {
		t.Fatalf("FindLongitude: %v", err)
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

// TestFindLongitude_Forward_SelfConsistent mirrors
// TestFindLongitude_Backward_SelfConsistent but searches forward from
// before the reference instant.
func TestFindLongitude_Forward_SelfConsistent(t *testing.T) {
	ref := swisseph.JulDay(2024, 6, 1, 0)
	pos, _, err := swisseph.CalcPlanet(ref, swisseph.Mars)
	if err != nil {
		t.Fatalf("CalcPlanet: %v", err)
	}

	// Search from a few days before ref for the longitude Mars held at ref.
	crossingJD, _, _, err := FindLongitude(ref-5, swisseph.Mars, pos.Longitude, 10, searchStepDays, Forward)
	if err != nil {
		t.Fatalf("FindLongitude: %v", err)
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

// TestFindLongitude_Backward_RetrogradeFlag checks that the retrograde flag
// returned always agrees with the sign of the speed at the crossing
// instant, by independently re-querying CalcPlanet.
func TestFindLongitude_Backward_RetrogradeFlag(t *testing.T) {
	jd := swisseph.JulDay(2024, 6, 1, 0)
	pos, _, err := swisseph.CalcPlanet(jd, swisseph.Mercury)
	if err != nil {
		t.Fatalf("CalcPlanet: %v", err)
	}

	crossingJD, speedLon, retrograde, err := FindLongitude(jd+3, swisseph.Mercury, pos.Longitude, 10, searchStepDays, Backward)
	if err != nil {
		t.Fatalf("FindLongitude: %v", err)
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

// TestFindLongitude_Forward_RetrogradeFlag mirrors
// TestFindLongitude_Backward_RetrogradeFlag but searches forward.
func TestFindLongitude_Forward_RetrogradeFlag(t *testing.T) {
	jd := swisseph.JulDay(2024, 6, 1, 0)
	pos, _, err := swisseph.CalcPlanet(jd, swisseph.Mercury)
	if err != nil {
		t.Fatalf("CalcPlanet: %v", err)
	}

	crossingJD, speedLon, retrograde, err := FindLongitude(jd-3, swisseph.Mercury, pos.Longitude, 10, searchStepDays, Forward)
	if err != nil {
		t.Fatalf("FindLongitude: %v", err)
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

// TestFindLongitudeDefault_UnsupportedPlanet checks that a planet with no
// entry in searchWindowDays (e.g. MeanNode, which the CLI never routes here)
// fails loudly rather than silently searching a zero-day window.
func TestFindLongitudeDefault_UnsupportedPlanet(t *testing.T) {
	_, _, _, err := FindLongitudeDefault(swisseph.JulDay(2024, 6, 1, 0), swisseph.MeanNode, 0, Backward)
	if err == nil {
		t.Fatal("expected an error for a planet with no default search window, got nil")
	}
}

// TestFindLongitude_Backward_NotFoundWithinWindow checks that a
// deliberately tiny window for a slow-moving planet returns an error rather
// than a false hit.
func TestFindLongitude_Backward_NotFoundWithinWindow(t *testing.T) {
	jd := swisseph.JulDay(2024, 6, 1, 0)
	pos, _, err := swisseph.CalcPlanet(jd, swisseph.Saturn)
	if err != nil {
		t.Fatalf("CalcPlanet: %v", err)
	}
	// Target a longitude far from Saturn's current position so a 1-day
	// window (much shorter than Saturn's ~0.034 deg/day motion could cover
	// to reach a longitude 90 degrees away) cannot find a crossing.
	target := normalize180(pos.Longitude + 90)

	_, _, _, err = FindLongitude(jd, swisseph.Saturn, target, 1, searchStepDays, Backward)
	if err == nil {
		t.Fatal("expected an error when no crossing exists within the window, got nil")
	}
}

// TestFindLongitude_Forward_NotFoundWithinWindow mirrors
// TestFindLongitude_Backward_NotFoundWithinWindow but searches forward.
func TestFindLongitude_Forward_NotFoundWithinWindow(t *testing.T) {
	jd := swisseph.JulDay(2024, 6, 1, 0)
	pos, _, err := swisseph.CalcPlanet(jd, swisseph.Saturn)
	if err != nil {
		t.Fatalf("CalcPlanet: %v", err)
	}
	target := normalize180(pos.Longitude + 90)

	_, _, _, err = FindLongitude(jd, swisseph.Saturn, target, 1, searchStepDays, Forward)
	if err == nil {
		t.Fatal("expected an error when no crossing exists within the window, got nil")
	}
}

// TestFindLongitude_BackwardForwardSymmetry checks that searching Backward
// from just after a known crossing and Forward from just before it both
// converge on the same crossing instant, exercising the per-direction
// bisection bracket.
func TestFindLongitude_BackwardForwardSymmetry(t *testing.T) {
	ref := swisseph.JulDay(2024, 6, 1, 0)
	pos, _, err := swisseph.CalcPlanet(ref, swisseph.Venus)
	if err != nil {
		t.Fatalf("CalcPlanet: %v", err)
	}

	backJD, _, _, err := FindLongitude(ref+2, swisseph.Venus, pos.Longitude, 10, searchStepDays, Backward)
	if err != nil {
		t.Fatalf("FindLongitude(Backward): %v", err)
	}
	fwdJD, _, _, err := FindLongitude(ref-2, swisseph.Venus, pos.Longitude, 10, searchStepDays, Forward)
	if err != nil {
		t.Fatalf("FindLongitude(Forward): %v", err)
	}

	if math.Abs(backJD-fwdJD) > 1.0/1440 { // within a minute
		t.Errorf("backward crossing = %.6f, forward crossing = %.6f, want them to match", backJD, fwdJD)
	}
}

// TestFindLongitude_Backward_MultipleCrossingsReturnsMostRecent validates
// that the backward scan stops at the chronologically most recent
// sign-change rather than an earlier one, for a target longitude Mercury is
// known to cross more than once within the window (its retrograde loop
// revisits longitudes it already passed through). The expected crossing is
// determined independently by brute-force fine sampling, not a hardcoded
// almanac date, and the test asserts at least two crossings actually exist
// in the window so it genuinely exercises the multiple-crossing case rather
// than degenerating into a single-crossing check.
func TestFindLongitude_Backward_MultipleCrossingsReturnsMostRecent(t *testing.T) {
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

	gotJD, _, _, err := FindLongitude(ref, swisseph.Mercury, target, window, searchStepDays, Backward)
	if err != nil {
		t.Fatalf("FindLongitude: %v", err)
	}
	if math.Abs(gotJD-wantJD) > 1 { // within a day of the brute-force estimate
		t.Errorf("FindLongitude crossingJD = %.4f, brute-force most-recent crossing = %.4f (all crossings: %v)", gotJD, wantJD, crossings)
	}
}

// TestFindLongitude_Forward_MultipleCrossingsReturnsSoonest mirrors
// TestFindLongitude_Backward_MultipleCrossingsReturnsMostRecent but scans
// forward, asserting the forward search returns the soonest upcoming
// crossing rather than a later one.
func TestFindLongitude_Forward_MultipleCrossingsReturnsSoonest(t *testing.T) {
	const window = 150.0 // > one Mercury synodic period (~116 days)
	const fineStep = 0.1

	// bruteForceCrossings independently fine-samples forward from ref over
	// window days and returns every sign change of the offset to target
	// (excluding the antipodal wraparound), soonest first.
	bruteForceCrossings := func(ref, target float64) ([]float64, error) {
		var crossings []float64
		prev, err := longitudeOffset(ref, swisseph.Mercury, target)
		if err != nil {
			return nil, err
		}
		tt := ref
		for i := 0; float64(i)*fineStep < window; i++ {
			tt += fineStep
			cur, err := longitudeOffset(tt, swisseph.Mercury, target)
			if err != nil {
				return nil, err
			}
			if (prev <= 0) != (cur <= 0) && math.Abs(cur-prev) < 180 {
				crossings = append(crossings, tt-fineStep/2)
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
		basePos, _, err := swisseph.CalcPlanet(candidateRef+window, swisseph.Mercury)
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
	wantJD := crossings[0] // first found scanning forward = soonest

	gotJD, _, _, err := FindLongitude(ref, swisseph.Mercury, target, window, searchStepDays, Forward)
	if err != nil {
		t.Fatalf("FindLongitude: %v", err)
	}
	if math.Abs(gotJD-wantJD) > 1 { // within a day of the brute-force estimate
		t.Errorf("FindLongitude crossingJD = %.4f, brute-force soonest crossing = %.4f (all crossings: %v)", gotJD, wantJD, crossings)
	}
}
