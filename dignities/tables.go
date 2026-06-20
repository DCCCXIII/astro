// Package dignities holds the static essential-dignity tables of traditional
// astrology (domicile, exaltation, triplicity, term, face) and pure lookups
// over them. It has no cgo dependency and is fully unit-testable.
//
// Sign indices run 0=Aries … 11=Pisces. Planet IDs match the Swiss Ephemeris
// integer IDs used by the swisseph package (Sun=0 … Saturn=6), so callers can
// convert freely between dignities.Planet and swisseph planet constants.
package dignities

// Planet identifies one of the seven traditional planets. The underlying
// integer values match the Swiss Ephemeris SE_* planet constants.
type Planet int

const (
	Sun     Planet = 0
	Moon    Planet = 1
	Mercury Planet = 2
	Venus   Planet = 3
	Mars    Planet = 4
	Jupiter Planet = 5
	Saturn  Planet = 6

	// None marks the absence of a ruler (e.g. a sign with no exaltation).
	None Planet = -1
)

// Element classifies a sign. element = sign % 4.
const (
	Fire  = 0
	Earth = 1
	Air   = 2
	Water = 3
)

// domicile[sign] is the planet that rules (is at home in) the sign. (+5)
var domicile = [12]Planet{
	Mars,    // Aries
	Venus,   // Taurus
	Mercury, // Gemini
	Moon,    // Cancer
	Sun,     // Leo
	Mercury, // Virgo
	Venus,   // Libra
	Mars,    // Scorpio
	Jupiter, // Sagittarius
	Saturn,  // Capricorn
	Saturn,  // Aquarius
	Jupiter, // Pisces
}

// detriment[sign] is the planet in its detriment in the sign (opposite of
// domicile). (−5, Lilly only)
var detriment = [12]Planet{
	Venus,   // Aries
	Mars,    // Taurus
	Jupiter, // Gemini
	Saturn,  // Cancer
	Saturn,  // Leo
	Jupiter, // Virgo
	Mars,    // Libra
	Venus,   // Scorpio
	Mercury, // Sagittarius
	Moon,    // Capricorn
	Sun,     // Aquarius
	Mercury, // Pisces
}

// exalt holds the planet exalted in a sign and the exact degree of exaltation.
// Planet is None for the four signs with no planetary exaltation.
type exalt struct {
	planet Planet
	degree float64
}

// exaltations[sign] — presence in the sign suffices for the +4; the degree is
// only needed for accidental "exact exaltation" niceties.
var exaltations = [12]exalt{
	{Sun, 19},     // Aries
	{Moon, 3},     // Taurus
	{None, 0},     // Gemini
	{Jupiter, 15}, // Cancer
	{None, 0},     // Leo
	{Mercury, 15}, // Virgo
	{Saturn, 21},  // Libra
	{None, 0},     // Scorpio
	{None, 0},     // Sagittarius
	{Mars, 28},    // Capricorn
	{None, 0},     // Aquarius
	{Venus, 27},   // Pisces
}

// triplicity holds the three Dorothean lords of an element's triplicity.
type triplicity struct {
	day, night, participating Planet
}

// triplicities is indexed by element (Fire, Earth, Air, Water).
var triplicities = [4]triplicity{
	{day: Sun, night: Jupiter, participating: Saturn},     // Fire
	{day: Venus, night: Moon, participating: Mars},        // Earth
	{day: Saturn, night: Mercury, participating: Jupiter}, // Air
	{day: Venus, night: Mars, participating: Moon},        // Water
}

// lillyWaterTriplicity overrides the water triplicity for Lilly, who assigns
// Mars as both day and night lord of the water signs.
var lillyWaterTriplicity = triplicity{day: Mars, night: Mars, participating: Moon}

// term is one Egyptian bound: lord rules from the previous boundary up to upTo
// (degrees within the sign).
type term struct {
	lord Planet
	upTo float64
}

// terms[sign] holds the five Egyptian bounds, ordered by ascending upper bound,
// collectively covering 0–30°. (+2)
var terms = [12][5]term{
	{{Jupiter, 6}, {Venus, 12}, {Mercury, 20}, {Mars, 25}, {Saturn, 30}},  // Aries
	{{Venus, 8}, {Mercury, 14}, {Jupiter, 22}, {Saturn, 27}, {Mars, 30}},  // Taurus
	{{Mercury, 6}, {Jupiter, 12}, {Venus, 17}, {Mars, 24}, {Saturn, 30}},  // Gemini
	{{Mars, 7}, {Venus, 13}, {Mercury, 19}, {Jupiter, 26}, {Saturn, 30}},  // Cancer
	{{Jupiter, 6}, {Venus, 11}, {Saturn, 18}, {Mercury, 24}, {Mars, 30}},  // Leo
	{{Mercury, 7}, {Venus, 17}, {Jupiter, 21}, {Mars, 28}, {Saturn, 30}},  // Virgo
	{{Saturn, 6}, {Mercury, 14}, {Jupiter, 21}, {Venus, 28}, {Mars, 30}},  // Libra
	{{Mars, 7}, {Venus, 11}, {Mercury, 19}, {Jupiter, 24}, {Saturn, 30}},  // Scorpio
	{{Jupiter, 12}, {Venus, 17}, {Mercury, 21}, {Saturn, 26}, {Mars, 30}}, // Sagittarius
	{{Mercury, 7}, {Jupiter, 14}, {Venus, 22}, {Saturn, 26}, {Mars, 30}},  // Capricorn
	{{Mercury, 7}, {Venus, 13}, {Jupiter, 20}, {Mars, 25}, {Saturn, 30}},  // Aquarius
	{{Venus, 12}, {Jupiter, 16}, {Mercury, 19}, {Mars, 28}, {Saturn, 30}}, // Pisces
}

// faces[sign] holds the three Chaldean-order decan rulers (0–10°, 10–20°,
// 20–30°). (+1)
var faces = [12][3]Planet{
	{Mars, Sun, Venus},      // Aries
	{Mercury, Moon, Saturn}, // Taurus
	{Jupiter, Mars, Sun},    // Gemini
	{Venus, Mercury, Moon},  // Cancer
	{Saturn, Jupiter, Mars}, // Leo
	{Sun, Venus, Mercury},   // Virgo
	{Moon, Saturn, Jupiter}, // Libra
	{Mars, Sun, Venus},      // Scorpio
	{Mercury, Moon, Saturn}, // Sagittarius
	{Jupiter, Mars, Sun},    // Capricorn
	{Venus, Mercury, Moon},  // Aquarius
	{Saturn, Jupiter, Mars}, // Pisces
}

// ChaldeanOrder lists the planets slowest→fastest, the sequence used for
// planetary hours (§3.6).
var ChaldeanOrder = [7]Planet{Saturn, Jupiter, Mars, Sun, Venus, Mercury, Moon}

// WeekdayRulers maps a weekday (Sunday=0 … Saturday=6, matching time.Weekday)
// to its planetary day-ruler.
var WeekdayRulers = [7]Planet{
	Sun,     // Sunday
	Moon,    // Monday
	Mars,    // Tuesday
	Mercury, // Wednesday
	Jupiter, // Thursday
	Venus,   // Friday
	Saturn,  // Saturday
}

// MeanDailyMotion is each planet's approximate mean daily motion in degrees,
// used to judge swift vs. slow (§5.2).
var MeanDailyMotion = map[Planet]float64{
	Saturn:  0.034,
	Jupiter: 0.083,
	Mars:    0.524,
	Sun:     0.986,
	Venus:   1.2,
	Mercury: 1.383,
	Moon:    13.176,
}

var planetNames = map[Planet]string{
	Sun: "Sun", Moon: "Moon", Mercury: "Mercury", Venus: "Venus",
	Mars: "Mars", Jupiter: "Jupiter", Saturn: "Saturn", None: "None",
}

// Name returns the planet's human-readable name.
func (p Planet) Name() string {
	if n, ok := planetNames[p]; ok {
		return n
	}
	return "Unknown"
}
