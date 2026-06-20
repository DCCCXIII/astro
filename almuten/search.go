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

// LastLongitude finds the most recent moment at or before jd that planet
// crossed targetLon (degrees, [0,360)). It returns that crossing's exact
// Julian Day, the planet's signed daily speed in longitude at the crossing
// (negative means retrograde), and whether it was retrograde. windowDays
// bounds how far back to search; if no crossing is found within the window
// it returns an error.
func LastLongitude(jd float64, planet int, targetLon, windowDays, stepDays float64) (crossingJD, speedLon float64, retrograde bool, err error) {
	prev, err := longitudeOffset(jd, planet, targetLon)
	if err != nil {
		return 0, 0, false, err
	}

	maxIter := int(windowDays/stepDays) + 1
	t := jd
	for i := 0; i < maxIter; i++ {
		t -= stepDays
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
			inst, err := refineLongitudeCrossing(t, t+stepDays, planet, targetLon)
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
	return 0, 0, false, fmt.Errorf("no occurrence of %s at %.4f° found within %.0f days before jd %.4f", swisseph.PlanetName(planet), targetLon, windowDays, jd)
}

// LastLongitudeDefault calls LastLongitude using this package's recommended
// search window and step size for planet.
func LastLongitudeDefault(jd float64, planet int, targetLon float64) (crossingJD, speedLon float64, retrograde bool, err error) {
	return LastLongitude(jd, planet, targetLon, searchWindowDays[planet], searchStepDays)
}

// refineLongitudeCrossing bisects the bracket [lo, hi] to ~1-second precision,
// locating the instant planet's longitude equals targetLon. It mirrors
// refineCrossing in syzygy.go but is generalized to any planet/target and is
// agnostic to whether the planet is moving direct or retrograde across the
// bracket.
func refineLongitudeCrossing(lo, hi float64, planet int, targetLon float64) (float64, error) {
	f := func(jd float64) (float64, error) {
		return longitudeOffset(jd, planet, targetLon)
	}

	flo, err := f(lo)
	if err != nil {
		return 0, err
	}

	const tol = 1.0 / 86400.0 // one second in days
	for hi-lo > tol {
		mid := (lo + hi) / 2
		fmid, err := f(mid)
		if err != nil {
			return 0, err
		}
		if (flo <= 0) == (fmid <= 0) {
			lo, flo = mid, fmid
		} else {
			hi = mid
		}
	}
	return (lo + hi) / 2, nil
}
