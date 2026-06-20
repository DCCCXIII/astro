package almuten

import (
	"errors"
	"math"

	"github.com/dcccxiii/astro/dignities"
	"github.com/dcccxiii/astro/swisseph"
)

// figurisHousePoints is Ibn Ezra's non-standard accidental house valuation
// (§4.3), distinct from Lilly's.
var figurisHousePoints = map[int]int{
	1: 12, 10: 11, 7: 10, 4: 9, 11: 8, 5: 7,
	2: 6, 9: 5, 8: 4, 3: 3, 12: 2, 6: 1,
}

// firstStationElongation is the approximate elongation (degrees) at which each
// superior planet turns retrograde, bounding the synodic banding (§4.5).
var firstStationElongation = map[int]float64{
	swisseph.Saturn:  109,
	swisseph.Jupiter: 115,
	swisseph.Mars:    136,
}

// AlmutenFiguris computes Ibn Ezra's Almuten Figuris: the seven planets scored
// over five vital places plus accidental, temporal, and synodic bonuses. It
// returns the full scorecard and every planet tied for the highest total.
// The chart's Figuris-specific inputs (Part of Fortune, prenatal syzygy, day and
// hour rulers) are computed here, so a plain BuildChart result suffices.
func AlmutenFiguris(c Chart, opts Options) (Scorecard, []int, error) {
	if err := c.prepareFiguris(); err != nil {
		return nil, nil, err
	}

	score := make(Scorecard, len(Planets))
	for _, p := range Planets {
		score[p] = 0
	}

	// --- essential scoring over the five vital places ---
	places := []float64{
		c.Houses.Ascendant,
		c.Positions[swisseph.Sun].Longitude,
		c.Positions[swisseph.Moon].Longitude,
		c.PartFortune,
		c.Syzygy,
	}
	for _, place := range places {
		addEssentialDignities(score, place, c.IsDiurnal)
	}

	// --- accidental house placement (5° pre-cusp orb) ---
	for _, p := range Planets {
		h := adjustedHouse(c.Positions[p].Longitude, c.Houses.Cusps)
		score[p] += figurisHousePoints[h]
	}

	// --- temporal points ---
	score[c.DayRuler] += 7
	score[c.HourRuler] += 6

	// --- synodic points (superiors separating from the Sun) ---
	sunLon := c.Positions[swisseph.Sun].Longitude
	for _, p := range []int{swisseph.Saturn, swisseph.Jupiter, swisseph.Mars} {
		lon := c.Positions[p].Longitude
		if normalize180(sunLon-lon) <= 0 {
			continue // Sun not ahead -> not separating
		}
		switch d := elongation(sunLon, lon); {
		case d >= 15 && d < 60:
			score[p] += 3
		case d >= 60 && d < 90:
			score[p] += 2
		case d >= 90 && d < firstStationElongation[p]:
			score[p] += 1
		}
	}

	return score, winners(score), nil
}

// addEssentialDignities adds the five dignity weights (5/4/3/2/1) at a longitude
// to the corresponding rulers. The triplicity uses the active-sect lord only
// (no participating lord, per Ibn Ezra).
func addEssentialDignities(score Scorecard, lon float64, isDiurnal bool) {
	sign, deg := signDeg(lon)

	score[int(dignities.DomicileRuler(sign))] += 5
	if exP, _, ok := dignities.ExaltationRuler(sign); ok {
		score[int(exP)] += 4
	}
	score[int(dignities.ActiveTriplicityRuler(sign, isDiurnal, false))] += 3
	score[int(dignities.TermRuler(sign, deg))] += 2
	score[int(dignities.FaceRuler(sign, deg))] += 1
}

// adjustedHouse returns the house a planet occupies, treating a planet within 5°
// before a cusp (in zodiacal order) as already in the next house (§4.3).
func adjustedHouse(lon float64, cusps [13]float64) int {
	for h := 1; h <= 12; h++ {
		next := h%12 + 1
		if norm360(cusps[next]-lon) <= 5.0 {
			return next
		}
	}
	return houseOf(lon, cusps)
}

// prepareFiguris fills the Figuris-specific chart inputs (weekday and
// planetary-hour rulers) on top of the shared vital points (Part of Fortune,
// prenatal syzygy) computed by computeVitalPoints.
func (c *Chart) prepareFiguris() error {
	if err := c.computeVitalPoints(); err != nil {
		return err
	}
	return c.computeTemporalRulers()
}

// computeTemporalRulers derives the weekday ruler (at the sunrise opening the
// astrological day) and the planetary-hour ruler containing the birth instant.
// A genuine polar day/night (ErrNoRiseSet) falls back to an equal-hour split;
// any other rise/set failure is a real error and is propagated rather than
// masked as a polar case.
func (c *Chart) computeTemporalRulers() error {
	sunrise, sunset, nextSunrise, err := c.sunRiseSetWindow()
	if err != nil {
		if errors.Is(err, swisseph.ErrNoRiseSet) {
			return c.temporalFallback()
		}
		return err
	}

	c.DayRuler = int(dignities.WeekdayRulers[weekday(sunrise)])

	// Locate the unequal hour containing the birth instant.
	var hourIndex int  // 0..23, hour 0 is the first hour after sunrise
	if c.JD < sunset { // daytime
		dayHour := (sunset - sunrise) / 12
		hourIndex = int((c.JD - sunrise) / dayHour)
	} else { // nighttime
		nightHour := (nextSunrise - sunset) / 12
		hourIndex = 12 + int((c.JD-sunset)/nightHour)
	}
	if hourIndex > 23 {
		hourIndex = 23
	}

	// Hours follow the Chaldean order starting from the day ruler.
	start := chaldeanIndex(c.DayRuler)
	c.HourRuler = int(dignities.ChaldeanOrder[(start+hourIndex)%7])
	return nil
}

// sunRiseSetWindow returns the sunrise opening the astrological day, the
// following sunset, and the next sunrise. The opening sunrise is the first
// sunrise after the instant exactly 24h before birth: it lands on the birth
// day's sunrise when born after sunrise, and the previous day's when born
// before it. Any rise/set error short-circuits and is returned to the caller.
func (c *Chart) sunRiseSetWindow() (sunrise, sunset, nextSunrise float64, err error) {
	if sunrise, err = swisseph.RiseTrans(c.JD-1.0, swisseph.Sun, c.Lat, c.Lon, 0, true); err != nil {
		return 0, 0, 0, err
	}
	if sunset, err = swisseph.RiseTrans(sunrise, swisseph.Sun, c.Lat, c.Lon, 0, false); err != nil {
		return 0, 0, 0, err
	}
	if nextSunrise, err = swisseph.RiseTrans(sunset, swisseph.Sun, c.Lat, c.Lon, 0, true); err != nil {
		return 0, 0, 0, err
	}
	return sunrise, sunset, nextSunrise, nil
}

// temporalFallback handles locations/times without a normal sunrise/sunset
// (polar). It assigns the weekday ruler from the birth instant and splits the
// civil day into 12 equal day and night hours starting at 06:00 UT.
func (c *Chart) temporalFallback() error {
	c.DayRuler = int(dignities.WeekdayRulers[weekday(c.JD)])
	sunrise := math.Floor(c.JD-0.5) + 0.5 + 0.25 // 06:00 UT of the civil day
	hourIndex := int(norm24((c.JD - sunrise) * 24))
	start := chaldeanIndex(c.DayRuler)
	c.HourRuler = int(dignities.ChaldeanOrder[(start+hourIndex)%7])
	return nil
}

// weekday returns the day of week (0=Sunday … 6=Saturday) for a Julian Day,
// rolling over at 00:00 UT.
func weekday(jd float64) int {
	return (int(math.Floor(jd+0.5)) + 1) % 7
}

// chaldeanIndex returns the position of a planet in the Chaldean order.
func chaldeanIndex(planet int) int {
	for i, p := range dignities.ChaldeanOrder {
		if int(p) == planet {
			return i
		}
	}
	return 0
}

// norm24 reduces an hour count to [0, 24).
func norm24(h float64) float64 {
	h = math.Mod(h, 24)
	if h < 0 {
		h += 24
	}
	return h
}
