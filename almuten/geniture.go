package almuten

import (
	"math"

	"github.com/dcccxiii/astro/dignities"
	"github.com/dcccxiii/astro/swisseph"
)

// lillyHousePoints is Lilly's accidental house valuation (Christian Astrology
// p.115), using the plain containing house.
var lillyHousePoints = map[int]int{
	1: 5, 10: 5,
	4: 4, 7: 4, 11: 4,
	2: 3, 5: 3,
	9:  2,
	3:  1,
	12: -5,
	6:  -2, 8: -2,
}

// fixedStars lists the three stars Lilly scores and their point values.
var fixedStars = []struct {
	name   string
	points int
}{
	{"Regulus", 6},
	{"Spica", 5},
	{"Algol", -4},
}

// LordOfGeniture computes Lilly's Lord of the Geniture: a fixed essential +
// accidental scorecard applied to each planet at its own position. It returns
// the full scorecard and every planet tied for the highest net total.
func LordOfGeniture(c Chart, opts Options) (Scorecard, []int) {
	score := newScorecard()
	sunLon := c.Positions[swisseph.Sun].Longitude

	for _, p := range Planets {
		score[p] = c.scoreGeniturePlanet(p, sunLon, opts)
	}

	return score, winners(score)
}

func (c Chart) scoreGeniturePlanet(p int, sunLon float64, opts Options) int {
	pos := c.Positions[p]
	lon := pos.Longitude
	sign, deg := signDeg(lon)
	dp := dignities.Planet(p)
	s := 0

	// --- essential dignities / debilities ---
	dig := dignities.EssentialDignities(dp, sign, deg, c.IsDiurnal, opts.LillyWater)
	inDomicile := dig.Domicile || c.mutualReceptionDomicile(p)
	inExalt := dig.Exalt || c.mutualReceptionExalt(p)

	if inDomicile {
		s += 5
	}
	if inExalt {
		s += 4
	}
	if dig.Trip {
		s += 3
	}

	if int(dignities.DetrimentRuler(sign)) == p {
		s -= 5
	}
	if fallP, ok := dignities.FallRuler(sign); ok && int(fallP) == p {
		s -= 4
	}

	// Peregrine: lacking essential dignity. With PeregrineSuppressesTermFace a
	// planet without any major dignity (domicile/exalt/triplicity, incl. mutual
	// reception) is peregrine even if it holds term/face, and the +2/+1 are
	// suppressed; otherwise term/face still count and only a wholly undignified
	// planet is peregrine.
	hasMajor := inDomicile || inExalt || dig.Trip
	peregrine := !hasMajor && !dig.Term && !dig.Face
	if opts.PeregrineSuppressesTermFace {
		peregrine = !hasMajor
	}
	if peregrine {
		s -= 5
	} else {
		if dig.Term {
			s += 2
		}
		if dig.Face {
			s += 1
		}
	}

	// --- motion (non-luminaries only) ---
	if p != swisseph.Sun && p != swisseph.Moon {
		if pos.SpeedLon < 0 {
			s -= 5 // retrograde
		} else {
			s += 4 // direct
		}
		if math.Abs(pos.SpeedLon) >= dignities.MeanDailyMotion[dp] {
			s += 2 // swift
		} else {
			s -= 2 // slow
		}
	}

	// --- solar relationship (non-Sun) ---
	if p != swisseph.Sun {
		switch d := elongation(lon, sunLon); {
		case d <= 0.283: // 17′ -> cazimi
			s += 5
		case d <= 8.5: // 8°30′ -> combust
			s -= 5
		case d <= 17: // under the Sun's beams
			s -= 4
		default: // free of the Sun
			s += opts.FreeOfSunBonus
		}
	}

	// --- oriental / occidental ---
	// oriental = rising before the Sun: planet at lower longitude than the Sun,
	// i.e. (planetLon - sunLon) normalized to (-180,180] is negative.
	if p != swisseph.Sun && p != swisseph.Moon {
		oriental := normalize180(lon-sunLon) < 0
		switch p {
		case swisseph.Saturn, swisseph.Jupiter, swisseph.Mars:
			if oriental {
				s += 2
			} else {
				s -= 2
			}
		case swisseph.Mercury, swisseph.Venus:
			if oriental {
				s -= 2
			} else {
				s += 2
			}
		}
	}

	// --- Moon waxing / waning ---
	if p == swisseph.Moon {
		if normalize180(lon-sunLon) > 0 { // Moon ahead of Sun by 0–180° -> waxing
			s += 2
		} else {
			s -= 2
		}
	}

	// --- accidental house placement ---
	s += lillyHousePoints[houseOf(lon, c.Houses.Cusps)]

	// --- partile aspects (orb < 1°) ---
	for _, ben := range []int{swisseph.Jupiter, swisseph.Venus} {
		if ben == p {
			continue
		}
		if asp, ok := partileAspect(lon, c.Positions[ben].Longitude); ok {
			switch asp {
			case 0:
				s += 5
			case 120:
				s += 4
			case 60:
				s += 3
			}
		}
	}
	for _, mal := range []int{swisseph.Saturn, swisseph.Mars} {
		if mal == p {
			continue
		}
		if asp, ok := partileAspect(lon, c.Positions[mal].Longitude); ok {
			switch asp {
			case 0:
				s -= 5
			case 180, 90:
				s -= 4
			}
		}
	}
	if asp, ok := partileAspect(lon, c.NorthNode); ok && asp == 0 {
		s += 4
	}
	if asp, ok := partileAspect(lon, c.NorthNode+180); ok && asp == 0 {
		s -= 4
	}

	// --- besiegement by Saturn and Mars ---
	if c.besiegedBySaturnMars(p) {
		s -= 4
	}

	// --- fixed stars ---
	for _, star := range fixedStars {
		sp, err := swisseph.FixStar(star.name, c.JD)
		if err != nil {
			continue // catalogue unavailable -> skip the bonus rather than fail
		}
		if elongation(lon, sp.Longitude) <= opts.FixedStarOrb {
			s += star.points
		}
	}

	return s
}

// partileAspect reports the Ptolemaic aspect (0, 60, 90, 120, 180) between two
// longitudes when they are partile (within a 1° orb).
func partileAspect(a, b float64) (aspect int, ok bool) {
	d := elongation(a, b)
	for _, asp := range []int{0, 60, 90, 120, 180} {
		if math.Abs(d-float64(asp)) < 1.0 {
			return asp, true
		}
	}
	return 0, false
}

// mutualReceptionDomicile reports whether p and its dispositor (by domicile)
// each occupy a sign ruled by the other.
func (c Chart) mutualReceptionDomicile(p int) bool {
	sign, _ := signDeg(c.Positions[p].Longitude)
	q := int(dignities.DomicileRuler(sign))
	if q == p {
		return false // simple domicile, not reception
	}
	qSign, _ := signDeg(c.Positions[q].Longitude)
	return int(dignities.DomicileRuler(qSign)) == p
}

// mutualReceptionExalt reports whether p and the planet exalted in its sign each
// occupy a sign in which the other is exalted.
func (c Chart) mutualReceptionExalt(p int) bool {
	sign, _ := signDeg(c.Positions[p].Longitude)
	exP, _, ok := dignities.ExaltationRuler(sign)
	if !ok || int(exP) == p {
		return false
	}
	q := int(exP)
	qSign, _ := signDeg(c.Positions[q].Longitude)
	qexP, _, ok := dignities.ExaltationRuler(qSign)
	return ok && int(qexP) == p
}

// besiegedBySaturnMars reports whether p sits on the minor arc between Saturn
// and Mars (≤ 30° wide, so the malefics genuinely enclose it) with no other
// classical planet intervening. This is a pragmatic body-besiegement test; the
// fuller "by ray" form is not modelled.
func (c Chart) besiegedBySaturnMars(p int) bool {
	if p == swisseph.Saturn || p == swisseph.Mars {
		return false
	}
	sat := c.Positions[swisseph.Saturn].Longitude
	mar := c.Positions[swisseph.Mars].Longitude
	lon := c.Positions[p].Longitude

	arc := elongation(sat, mar)
	if arc == 0 || arc > 30 {
		return false
	}
	// Determine the start of the minor arc (going zodiacally) from sat to mar.
	start, end := sat, mar
	if norm360(mar-sat) > 180 {
		start, end = mar, sat
	}
	span := norm360(end - start)
	if norm360(lon-start) >= span {
		return false // p not between the malefics
	}
	// No other classical planet may lie between them.
	for _, q := range Planets {
		if q == p || q == swisseph.Saturn || q == swisseph.Mars {
			continue
		}
		if norm360(c.Positions[q].Longitude-start) < span {
			return false
		}
	}
	return true
}
