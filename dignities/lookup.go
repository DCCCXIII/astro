package dignities

// ElementOf returns the element (Fire, Earth, Air, Water) of a sign.
func ElementOf(sign int) int { return sign % 4 }

// DomicileRuler returns the planet that rules (domiciles in) the sign. (+5)
func DomicileRuler(sign int) Planet { return domicile[sign] }

// DetrimentRuler returns the planet in its detriment in the sign. (−5)
func DetrimentRuler(sign int) Planet { return detriment[sign] }

// ExaltationRuler returns the planet exalted in the sign and its exaltation
// degree. ok is false for signs with no planetary exaltation.
func ExaltationRuler(sign int) (planet Planet, degree float64, ok bool) {
	e := exaltations[sign]
	if e.planet == None {
		return None, 0, false
	}
	return e.planet, e.degree, true
}

// FallRuler returns the planet in its fall in the sign (the planet exalted in
// the opposite sign). ok is false for signs with no associated fall.
func FallRuler(sign int) (planet Planet, ok bool) {
	opposite := (sign + 6) % 12
	e := exaltations[opposite]
	if e.planet == None {
		return None, false
	}
	return e.planet, true
}

// ActiveTriplicityRuler returns the active-sect triplicity lord of the sign:
// the day lord when diurnal, otherwise the night lord. When lillyWater is set,
// the water signs use Lilly's Mars/Mars override. (+3)
func ActiveTriplicityRuler(sign int, isDiurnal, lillyWater bool) Planet {
	t := triplicities[ElementOf(sign)]
	if lillyWater && ElementOf(sign) == Water {
		t = lillyWaterTriplicity
	}
	if isDiurnal {
		return t.day
	}
	return t.night
}

// ParticipatingTriplicityRuler returns the participating triplicity lord of the
// sign (used by Bonatti-style extensions; excluded from the default almuten).
func ParticipatingTriplicityRuler(sign int) Planet {
	return triplicities[ElementOf(sign)].participating
}

// TermRuler returns the Egyptian term (bound) lord governing degInSign, the
// degree within the sign in [0, 30). (+2)
func TermRuler(sign int, degInSign float64) Planet {
	for _, t := range terms[sign] {
		if degInSign < t.upTo {
			return t.lord
		}
	}
	return terms[sign][4].lord // degInSign == 30 falls in the last term
}

// FaceRuler returns the Chaldean face (decan) lord governing degInSign. (+1)
func FaceRuler(sign int, degInSign float64) Planet {
	idx := int(degInSign / 10.0)
	if idx > 2 {
		idx = 2
	}
	return faces[sign][idx]
}

// Dignities records which essential dignities a planet holds at a position.
type Dignities struct {
	Domicile bool
	Exalt    bool
	Trip     bool
	Term     bool
	Face     bool
}

// EssentialDignities reports the essential dignities held by planet at the given
// sign and degree, for the active sect. lillyWater selects Lilly's water
// triplicity override. Used for scoring and peregrine detection.
func EssentialDignities(planet Planet, sign int, degInSign float64, isDiurnal, lillyWater bool) Dignities {
	exaltP, _, hasExalt := ExaltationRuler(sign)
	return Dignities{
		Domicile: DomicileRuler(sign) == planet,
		Exalt:    hasExalt && exaltP == planet,
		Trip:     ActiveTriplicityRuler(sign, isDiurnal, lillyWater) == planet,
		Term:     TermRuler(sign, degInSign) == planet,
		Face:     FaceRuler(sign, degInSign) == planet,
	}
}
