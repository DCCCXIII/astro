package almuten

import (
	"github.com/dcccxiii/astro/dignities"
	"github.com/dcccxiii/astro/swisseph"
)

// AlmutenBonatti computes Guido Bonatti's Almudebit (Liber Astronomiae,
// c.1277): essential dignity scoring over two groups of points, each followed
// by a second round at that point's dispositor(s):
//   - four radix angles (Asc, MC, Desc, IC): chase the active-sect and
//     participating triplicity lords
//   - four vital points (Sun, Moon, Part of Fortune, prenatal syzygy): chase
//     the domicile lord
//
// Unlike Figuris, Bonatti's method has no accidental house, temporal, or
// synodic component.
func AlmutenBonatti(c Chart) (Scorecard, []int, error) {
	if err := c.computeVitalPoints(); err != nil {
		return nil, nil, err
	}

	score := newScorecard()
	c.scoreBonatti(score)

	return score, winners(score), nil
}

// scoreBonatti adds essential dignity points for the four radix angles and
// four vital points, each followed by a second round of scoring at their
// dispositor(s). PartFortune and Syzygy must already be populated on c (see
// computeVitalPoints).
func (c Chart) scoreBonatti(score Scorecard) {
	c.scoreAnglesWithDispositors(score)
	c.scoreVitalPointsWithDispositor(score)
}

// scoreAnglesWithDispositors scores the four radix angles (Asc, MC, Desc,
// IC), then chases each angle's active-sect and participating triplicity
// lords to their own chart position for a second round of scoring.
func (c Chart) scoreAnglesWithDispositors(score Scorecard) {
	desc := norm360(c.Houses.Ascendant + 180)
	ic := norm360(c.Houses.MC + 180)
	for _, angle := range []float64{c.Houses.Ascendant, c.Houses.MC, desc, ic} {
		addEssentialDignities(score, angle, c.IsDiurnal)
		sign, _ := signDeg(angle)
		for _, lord := range []int{
			int(dignities.ActiveTriplicityRuler(sign, c.IsDiurnal, false)),
			int(dignities.ParticipatingTriplicityRuler(sign)),
		} {
			addEssentialDignities(score, c.Positions[lord].Longitude, c.IsDiurnal)
		}
	}
}

// scoreVitalPointsWithDispositor scores the four vital points (Sun, Moon,
// Part of Fortune, prenatal syzygy), then chases each point's domicile lord
// to its own chart position for a second round of scoring.
func (c Chart) scoreVitalPointsWithDispositor(score Scorecard) {
	vitalPoints := []float64{
		c.Positions[swisseph.Sun].Longitude,
		c.Positions[swisseph.Moon].Longitude,
		c.PartFortune,
		c.Syzygy,
	}
	for _, v := range vitalPoints {
		addEssentialDignities(score, v, c.IsDiurnal)
		sign, _ := signDeg(v)
		lord := int(dignities.DomicileRuler(sign))
		addEssentialDignities(score, c.Positions[lord].Longitude, c.IsDiurnal)
	}
}
