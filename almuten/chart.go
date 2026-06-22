// Package almuten computes traditional "chart victor" rulers: Lilly's Lord of
// the Geniture, Ibn Ezra's Almuten Figuris, and Guido Bonatti's Almudebit. All
// three build a per-planet integer scorecard and return its argmax. swisseph
// is the only cgo boundary; the scoring tables live in the pure-Go dignities
// package.
package almuten

import (
	"fmt"
	"math"

	"github.com/dcccxiii/astro/swisseph"
)

// Planets is the set of seven traditional planets scored by both algorithms,
// in a stable order used for deterministic tie reporting.
var Planets = []int{
	swisseph.Sun, swisseph.Moon, swisseph.Mercury, swisseph.Venus,
	swisseph.Mars, swisseph.Jupiter, swisseph.Saturn,
}

// Scorecard maps a planet ID (swisseph constant) to its integer score.
type Scorecard map[int]int

// newScorecard returns an empty Scorecard sized for all seven Planets.
func newScorecard() Scorecard {
	return make(Scorecard, len(Planets))
}

// Chart gathers every input the scoring algorithms need, computed once.
// PartFortune, Syzygy, and SyzygyType are populated by computeVitalPoints,
// shared by AlmutenFiguris and AlmutenBonatti. DayRuler and HourRuler
// (temporal rulers) are populated only by the Almuten Figuris path.
type Chart struct {
	JD        float64
	Lat, Lon  float64
	IsDiurnal bool
	Positions map[int]swisseph.PlanetPos
	NorthNode float64 // ecliptic longitude of the Moon's mean north node
	Houses    swisseph.HouseResult

	PartFortune float64
	Syzygy      float64
	SyzygyType  string // "new" | "full"; informational — scoring uses the Syzygy longitude only
	DayRuler    int
	HourRuler   int
}

// Options exposes the spec's documented variant toggles (§8). DefaultOptions
// returns the recommended defaults.
type Options struct {
	FreeOfSunBonus              int     // points when a planet is free of the Sun (default 5)
	PeregrineSuppressesTermFace bool    // treat a planet lacking major dignity as peregrine, suppress term/face (default true)
	FixedStarOrb                float64 // conjunction orb for Regulus/Spica/Algol in degrees (default 5)
	LillyWater                  bool    // use Lilly's Mars/Mars water triplicity (used only by LordOfGeniture)
}

// DefaultOptions returns the recommended defaults for the Lord of the Geniture.
func DefaultOptions() Options {
	return Options{
		FreeOfSunBonus:              5,
		PeregrineSuppressesTermFace: true,
		FixedStarOrb:                5,
		LillyWater:                  true,
	}
}

// BuildChart computes planetary positions, the lunar node, houses, and sect for
// a given moment and location. It is shared by both scoring algorithms.
func BuildChart(jd, lat, lon float64, hsys byte) (Chart, error) {
	c := Chart{JD: jd, Lat: lat, Lon: lon, Positions: make(map[int]swisseph.PlanetPos, len(Planets))}

	for _, p := range Planets {
		pos, _, err := swisseph.CalcPlanet(jd, p)
		if err != nil {
			return Chart{}, fmt.Errorf("calculating %s: %w", swisseph.PlanetName(p), err)
		}
		c.Positions[p] = pos
	}

	node, _, err := swisseph.CalcPlanet(jd, swisseph.MeanNode)
	if err != nil {
		return Chart{}, fmt.Errorf("calculating north node: %w", err)
	}
	c.NorthNode = node.Longitude

	houses, err := swisseph.CalcHouses(jd, lat, lon, hsys)
	if err != nil {
		return Chart{}, fmt.Errorf("calculating houses: %w", err)
	}
	c.Houses = houses

	// Diurnal when the Sun is above the horizon (houses 7–12).
	c.IsDiurnal = HouseOf(c.Positions[swisseph.Sun].Longitude, houses.Cusps) >= 7

	return c, nil
}

// computeVitalPoints fills the Part of Fortune and prenatal syzygy, shared by
// any algorithm that scores essential dignities at those points.
func (c *Chart) computeVitalPoints() error {
	asc := c.Houses.Ascendant
	sun := c.Positions[swisseph.Sun].Longitude
	moon := c.Positions[swisseph.Moon].Longitude

	if c.IsDiurnal {
		c.PartFortune = norm360(asc + moon - sun)
	} else {
		c.PartFortune = norm360(asc + sun - moon)
	}

	syz, kind, err := PrenatalSyzygy(c.JD, c.Lat, c.Lon)
	if err != nil {
		return err
	}
	c.Syzygy, c.SyzygyType = syz, kind
	return nil
}

// --- geometric helpers -----------------------------------------------------

// norm360 reduces an angle to [0, 360).
func norm360(x float64) float64 {
	x = math.Mod(x, 360)
	if x < 0 {
		x += 360
	}
	return x
}

// normalize180 reduces an angle to (-180, 180].
func normalize180(x float64) float64 {
	x = math.Mod(x, 360)
	if x <= -180 {
		x += 360
	}
	if x > 180 {
		x -= 360
	}
	return x
}

// elongation returns the absolute angular separation of two longitudes, [0, 180].
func elongation(a, b float64) float64 {
	return math.Abs(normalize180(a - b))
}

// bisectZeroCrossing bisects [lo, hi] to ~1-second-of-JD precision, locating
// the instant where f crosses zero. f must be continuous and change sign
// somewhere within [lo, hi]. Shared by syzygy.go's lunation search and
// search.go's longitude search.
func bisectZeroCrossing(lo, hi float64, f func(jd float64) (float64, error)) (float64, error) {
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

// signDeg returns the sign index (0=Aries … 11=Pisces) and the degree within the
// sign [0, 30) for an ecliptic longitude.
func signDeg(lon float64) (sign int, deg float64) {
	l := norm360(lon)
	s := int(l / 30)
	if s >= 12 {
		s = 11
	}
	return s, l - float64(s)*30
}

// HouseOf returns the quadrant house (1–12) containing the longitude, using the
// plain containing house defined by the cusps (no orb).
func HouseOf(lon float64, cusps [13]float64) int {
	lon = norm360(lon)
	for h := 1; h <= 12; h++ {
		start := norm360(cusps[h])
		end := norm360(cusps[h%12+1])
		span := norm360(end - start)
		if span == 0 {
			continue
		}
		if norm360(lon-start) < span {
			return h
		}
	}
	// Defensive fallback for degenerate cusp configurations (e.g. duplicate
	// adjacent cusps at extreme latitudes); not a meaningful house assignment.
	return 12
}

// winners returns every planet tied for the maximum score, in Planets order.
func winners(s Scorecard) []int {
	max := math.MinInt
	for _, p := range Planets {
		if s[p] > max {
			max = s[p]
		}
	}
	var w []int
	for _, p := range Planets {
		if s[p] == max {
			w = append(w, p)
		}
	}
	return w
}
