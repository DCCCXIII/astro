package almuten

import (
	"fmt"
	"math"

	"github.com/dcccxiii/astro/swisseph"
)

// searchWindowDays bounds how far back LastLongitudeDefault searches before
// giving up, keyed by swisseph planet ID. Sized generously against each
// planet's synodic/orbital period so a target longitude is guaranteed to
// recur within the window even accounting for retrograde loops.
var searchWindowDays = map[int]float64{
	swisseph.Sun:     400,
	swisseph.Moon:    40,
	swisseph.Mercury: 400,
	swisseph.Venus:   700,
	swisseph.Mars:    1000,
	swisseph.Jupiter: 4800,
	swisseph.Saturn:  11500,
}

// searchStepDays is the coarse backward-scan step. It is conservative
// relative to PrenatalSyzygy's proven 0.5-day step, since non-luminary
// planets can have faster instantaneous retrograde speeds than the Moon's
// elongation rate.
const searchStepDays = 0.25

// longitudeOffset returns planet's longitude minus targetLon, reduced to
// (-180, 180], at the given Julian Day.
func longitudeOffset(jd float64, planet int, targetLon float64) (float64, error) {
	pos, _, err := swisseph.CalcPlanet(jd, planet)
	if err != nil {
		return 0, err
	}
	return normalize180(pos.Longitude - targetLon), nil
}

// Direction selects which way FindLongitude scans from the reference Julian
// Day: Backward for the most recent prior occurrence, Forward for the next
// occurrence.
type Direction int

const (
	Backward Direction = iota
	Forward
)

// String returns "before" for Backward and "after" for Forward, used to
// phrase the not-found error message.
func (d Direction) String() string {
	if d == Forward {
		return "after"
	}
	return "before"
}

// FindLongitude finds the nearest moment, in the given direction from jd,
// at which planet crosses targetLon (degrees, [0,360)). Backward searches
// for the most recent prior occurrence (at or before jd); Forward searches
// for the next occurrence (at or after jd). It returns that crossing's
// exact Julian Day, the planet's signed daily speed in longitude at the
// crossing (negative means retrograde), and whether it was retrograde.
// windowDays bounds how far to search; if no crossing is found within the
// window it returns an error.
func FindLongitude(jd float64, planet int, targetLon, windowDays, stepDays float64, dir Direction) (crossingJD, speedLon float64, retrograde bool, err error) {
	prev, err := longitudeOffset(jd, planet, targetLon)
	if err != nil {
		return 0, 0, false, err
	}

	maxIter := int(windowDays/stepDays) + 1
	t := jd
	for i := 0; i < maxIter; i++ {
		var lo, hi float64
		if dir == Forward {
			t += stepDays
			lo, hi = t-stepDays, t
		} else {
			t -= stepDays
			lo, hi = t, t+stepDays
		}
		cur, err := longitudeOffset(t, planet, targetLon)
		if err != nil {
			return 0, 0, false, err
		}

		// A real crossing has prev and cur both near 0. Excluding the
		// large-jump case rejects the offset itself wrapping at ±180 (i.e.
		// the planet passing targetLon+180, the offset's antipode), which is
		// not a crossing of targetLon and would otherwise be a false
		// positive on every such pass.
		if (prev <= 0) != (cur <= 0) && math.Abs(cur-prev) < 180 {
			inst, err := refineLongitudeCrossing(lo, hi, planet, targetLon)
			if err != nil {
				return 0, 0, false, err
			}
			pos, _, err := swisseph.CalcPlanet(inst, planet)
			if err != nil {
				return 0, 0, false, err
			}
			return inst, pos.SpeedLon, pos.SpeedLon < 0, nil
		}
		prev = cur
	}
	return 0, 0, false, fmt.Errorf("no occurrence of %s at %.4f° found within %.0f days %s jd %.4f", swisseph.PlanetName(planet), targetLon, windowDays, dir, jd)
}

// FindLongitudeDefault calls FindLongitude using this package's recommended
// search window and step size for planet.
func FindLongitudeDefault(jd float64, planet int, targetLon float64, dir Direction) (crossingJD, speedLon float64, retrograde bool, err error) {
	window, ok := searchWindowDays[planet]
	if !ok {
		return 0, 0, false, fmt.Errorf("no default search window for planet %s", swisseph.PlanetName(planet))
	}
	return FindLongitude(jd, planet, targetLon, window, searchStepDays, dir)
}

// refineLongitudeCrossing bisects the bracket [lo, hi] to ~1-second precision,
// locating the instant planet's longitude equals targetLon. It is agnostic to
// whether the planet is moving direct or retrograde across the bracket.
func refineLongitudeCrossing(lo, hi float64, planet int, targetLon float64) (float64, error) {
	return bisectZeroCrossing(lo, hi, func(jd float64) (float64, error) {
		return longitudeOffset(jd, planet, targetLon)
	})
}
