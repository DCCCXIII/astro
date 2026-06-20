package almuten

import (
	"fmt"
	"math"

	"github.com/dcccxiii/astro/swisseph"
)

// lunarElongation returns the Moon−Sun elongation reduced to (-180, 180] at the
// given Julian Day. It is 0 at a new moon and ±180 at a full moon.
func lunarElongation(jd float64) (float64, error) {
	moon, _, err := swisseph.CalcPlanet(jd, swisseph.Moon)
	if err != nil {
		return 0, err
	}
	sun, _, err := swisseph.CalcPlanet(jd, swisseph.Sun)
	if err != nil {
		return 0, err
	}
	return normalize180(moon.Longitude - sun.Longitude), nil
}

// PrenatalSyzygy finds the most recent lunation (new or full moon) at or before
// jd and returns its ecliptic longitude and type ("new" | "full").
//
// For a new moon the longitude is the conjunction degree (Sun = Moon). For a
// full moon it is the longitude of whichever luminary was above the horizon at
// the syzygy instant, evaluated at the birth location (lat, lon).
func PrenatalSyzygy(jd, lat, lon float64) (syzygyLon float64, kind string, err error) {
	const step = 0.5 // coarse backward step in days

	prev, err := lunarElongation(jd)
	if err != nil {
		return 0, "", err
	}

	t := jd
	for i := 0; i < 80; i++ { // up to 40 days back — more than a synodic month
		t -= step
		cur, err := lunarElongation(t)
		if err != nil {
			return 0, "", err
		}

		switch {
		case math.Abs(cur-prev) > 180: // wrap across ±180 -> full moon
			inst, err := refineCrossing(t, t+step, true)
			if err != nil {
				return 0, "", err
			}
			return fullMoonLongitude(inst, lat, lon)
		case (prev <= 0) != (cur <= 0): // sign change across 0 -> new moon
			inst, err := refineCrossing(t, t+step, false)
			if err != nil {
				return 0, "", err
			}
			sun, _, err := swisseph.CalcPlanet(inst, swisseph.Sun)
			if err != nil {
				return 0, "", err
			}
			return norm360(sun.Longitude), "new", nil
		}
		prev = cur
	}
	return 0, "", fmt.Errorf("no prenatal syzygy found within 40 days before jd %.4f", jd)
}

// refineCrossing bisects the bracket [lo, hi] to ~1-second precision. When full
// is true the target is a full moon (elongation crossing ±180); otherwise a new
// moon (elongation crossing 0). It returns the crossing instant.
func refineCrossing(lo, hi float64, full bool) (float64, error) {
	return bisectZeroCrossing(lo, hi, func(jd float64) (float64, error) {
		e, err := lunarElongation(jd)
		if err != nil {
			return 0, err
		}
		if full {
			return normalize180(e - 180), nil
		}
		return e, nil
	})
}

// fullMoonLongitude returns the longitude of the luminary above the horizon at a
// full-moon instant, evaluated at the given location.
func fullMoonLongitude(jd, lat, lon float64) (float64, string, error) {
	sun, _, err := swisseph.CalcPlanet(jd, swisseph.Sun)
	if err != nil {
		return 0, "", err
	}
	moon, _, err := swisseph.CalcPlanet(jd, swisseph.Moon)
	if err != nil {
		return 0, "", err
	}
	houses, err := swisseph.CalcHouses(jd, lat, lon, swisseph.HousePlacidus)
	if err != nil {
		return 0, "", err
	}
	// The luminary in houses 7–12 is above the horizon.
	if HouseOf(sun.Longitude, houses.Cusps) >= 7 {
		return norm360(sun.Longitude), "full", nil
	}
	return norm360(moon.Longitude), "full", nil
}
