package output

import (
	"fmt"
	"sort"

	"github.com/dcccxiii/astro/almuten"
	"github.com/dcccxiii/astro/swisseph"
)

// PlanetEntry holds presentation-ready data for a single planet.
// It is a pure internal data carrier; json.go owns all wire-format decisions.
type PlanetEntry struct {
	Name          string
	Longitude     float64
	Sign          string
	SignDegree    float64
	Speed         float64
	Latitude      float64
	Distance      float64
	SpeedLat      float64
	SpeedDistance float64
}

// AngleEntry holds presentation-ready data for a chart angle (Ascendant, MC).
type AngleEntry struct {
	Longitude  float64 `json:"longitude"`
	Sign       string  `json:"sign"`
	SignDegree float64 `json:"sign_degree"`
}

// CuspEntry holds presentation-ready data for a single house cusp.
type CuspEntry struct {
	House      int     `json:"house"`
	Longitude  float64 `json:"longitude"`
	Sign       string  `json:"sign"`
	SignDegree float64 `json:"sign_degree"`
}

// Result holds all computed, presentation-ready chart data. Both PrintText
// and PrintJSON render from this struct; neither calls swisseph directly.
// All fields — including ARMC, Vertex, and the verbose planet fields
// (Latitude, Distance, SpeedLat, SpeedDistance) — are always populated by
// Build(). Renderers decide which fields to surface based on the verbose flag.
type Result struct {
	JulianDay        float64
	HouseName        string
	Lat              float64
	Lon              float64
	Planets          []PlanetEntry
	Ascendant        AngleEntry
	MC               AngleEntry
	ARMC             float64        // sidereal time in degrees
	Vertex           AngleEntry     // ecliptic longitude of the Vertex
	Cusps            []CuspEntry    // one entry per house, 1-12
	EphemerisWarning string         // non-empty when the library fell back to Moshier
	Almuten          []AlmutenEntry // chart-victor results, one per algorithm requested
	Search           *SearchResult  // longitude search result; nil unless --search was passed
}

// SearchResult holds the outcome of a --search longitude lookup: the
// nearest moment — before or after the reference date, per Direction — at
// which a planet sat at a target ecliptic longitude.
type SearchResult struct {
	Planet        string
	TargetLon     float64
	TargetSign    string
	TargetSignDeg float64
	Direction     string // "forward" or "backward"
	JulianDay     float64
	Year          int
	Month         int
	Day           int
	Hour          float64 // decimal UT
	SpeedLon      float64 // degrees/day at the crossing; negative = retrograde
	Retrograde    bool
	House         int // 1-12, occupied by the planet at the crossing instant
}

// AlmutenScore is a single planet's net score in an almuten scorecard.
type AlmutenScore struct {
	Planet string `json:"planet"`
	Score  int    `json:"score"`
}

// AlmutenEntry holds the result of one chart-victor algorithm: the method name,
// every planet tied for the highest score, and the full scorecard (descending).
type AlmutenEntry struct {
	Method     string         `json:"method"`
	Winners    []string       `json:"winners"`
	Scoreboard []AlmutenScore `json:"scoreboard"`
}

// BuildAlmuten computes the requested chart-victor algorithm(s) for the given
// moment and location. mode is "geniture", "figuris", "bonatti", "both"
// (geniture+figuris), or "all". The chart is built once and reused. All
// swisseph access happens inside the almuten package.
func BuildAlmuten(jd, lat, lon float64, hsys byte, mode string) ([]AlmutenEntry, error) {
	chart, err := almuten.BuildChart(jd, lat, lon, hsys)
	if err != nil {
		return nil, fmt.Errorf("building chart: %w", err)
	}

	var entries []AlmutenEntry
	if mode == "geniture" || mode == "both" || mode == "all" {
		score, winners := almuten.LordOfGeniture(chart, almuten.DefaultOptions())
		entries = append(entries, almutenEntry("Lord of the Geniture", score, winners))
	}
	if mode == "figuris" || mode == "both" || mode == "all" {
		score, winners, err := almuten.AlmutenFiguris(chart)
		if err != nil {
			return nil, fmt.Errorf("computing almuten figuris: %w", err)
		}
		entries = append(entries, almutenEntry("Almuten Figuris", score, winners))
	}
	if mode == "bonatti" || mode == "all" {
		score, winners, err := almuten.AlmutenBonatti(chart)
		if err != nil {
			return nil, fmt.Errorf("computing almuten bonatti: %w", err)
		}
		entries = append(entries, almutenEntry("Bonatti's Almudebit", score, winners))
	}
	return entries, nil
}

// almutenEntry converts an almuten scorecard into a presentation-ready entry,
// with the scoreboard sorted by score (descending) then traditional planet order.
func almutenEntry(method string, score almuten.Scorecard, winners []int) AlmutenEntry {
	board := make([]AlmutenScore, 0, len(almuten.Planets))
	for _, p := range almuten.Planets {
		board = append(board, AlmutenScore{Planet: swisseph.PlanetName(p), Score: score[p]})
	}
	sort.SliceStable(board, func(i, j int) bool {
		if board[i].Score != board[j].Score {
			return board[i].Score > board[j].Score
		}
		return false
	})

	names := make([]string, len(winners))
	for i, p := range winners {
		names[i] = swisseph.PlanetName(p)
	}

	return AlmutenEntry{Method: method, Winners: names, Scoreboard: board}
}

// BuildSearch finds the nearest time, in the given direction ("forward" or
// "backward") relative to jd, that planet was/will be at targetLon, and
// returns a presentation-ready SearchResult. lat, lon, and hsys are used
// only to report the house the planet occupied at the moment found (the
// longitude search itself is location-independent).
func BuildSearch(jd float64, planet int, targetLon, lat, lon float64, hsys byte, direction string) (*SearchResult, error) {
	dir, err := searchDirectionFromString(direction)
	if err != nil {
		return nil, err
	}

	crossingJD, speedLon, retrograde, err := almuten.FindLongitudeDefault(jd, planet, targetLon, dir)
	if err != nil {
		return nil, fmt.Errorf("searching for %s at %.4f°: %w", swisseph.PlanetName(planet), targetLon, err)
	}

	year, month, day, hour := swisseph.RevJul(crossingJD)

	houses, err := swisseph.CalcHouses(crossingJD, lat, lon, hsys)
	if err != nil {
		return nil, fmt.Errorf("computing houses at crossing: %w", err)
	}
	pos, _, err := swisseph.CalcPlanet(crossingJD, planet)
	if err != nil {
		return nil, fmt.Errorf("recomputing %s position at crossing: %w", swisseph.PlanetName(planet), err)
	}

	sign, deg := swisseph.ZodiacSign(targetLon)
	return &SearchResult{
		Planet:        swisseph.PlanetName(planet),
		TargetLon:     targetLon,
		TargetSign:    sign,
		TargetSignDeg: deg,
		Direction:     direction,
		JulianDay:     crossingJD,
		Year:          year,
		Month:         month,
		Day:           day,
		Hour:          hour,
		SpeedLon:      speedLon,
		Retrograde:    retrograde,
		House:         almuten.HouseOf(pos.Longitude, houses.Cusps),
	}, nil
}

// searchDirectionFromString validates and converts the --search direction
// token into an almuten.Direction.
func searchDirectionFromString(direction string) (almuten.Direction, error) {
	switch direction {
	case "backward":
		return almuten.Backward, nil
	case "forward":
		return almuten.Forward, nil
	default:
		return 0, fmt.Errorf("unknown search direction %q: valid values are forward, backward", direction)
	}
}

// Build computes a full chart result for the given Julian Day, planets, and
// geographic location. All swisseph calls are concentrated here.
func Build(jd float64, planets []int, lat, lon float64, hsys byte, hsysName string) (Result, error) {
	r := Result{JulianDay: jd, HouseName: hsysName, Lat: lat, Lon: lon}

	for _, p := range planets {
		name := swisseph.PlanetName(p)
		pos, warning, err := swisseph.CalcPlanet(jd, p)
		if err != nil {
			return Result{}, fmt.Errorf("error calculating %s: %w", name, err)
		}
		// The Moshier fallback produces the same warning text for every planet,
		// so capturing the first non-empty warning is sufficient to represent
		// the fallback condition for the entire calculation.
		if warning != "" && r.EphemerisWarning == "" {
			r.EphemerisWarning = warning
		}
		sign, deg := swisseph.ZodiacSign(pos.Longitude)
		r.Planets = append(r.Planets, PlanetEntry{
			Name:          name,
			Longitude:     pos.Longitude,
			Sign:          sign,
			SignDegree:    deg,
			Speed:         pos.SpeedLon,
			Latitude:      pos.Latitude,
			Distance:      pos.Distance,
			SpeedLat:      pos.SpeedLat,
			SpeedDistance: pos.SpeedDistance,
		})
	}

	houses, err := swisseph.CalcHouses(jd, lat, lon, hsys)
	if err != nil {
		return Result{}, fmt.Errorf("error calculating houses: %w", err)
	}

	ascSign, ascDeg := swisseph.ZodiacSign(houses.Ascendant)
	mcSign, mcDeg := swisseph.ZodiacSign(houses.MC)
	vtxSign, vtxDeg := swisseph.ZodiacSign(houses.Vertex)
	r.Ascendant = AngleEntry{Longitude: houses.Ascendant, Sign: ascSign, SignDegree: ascDeg}
	r.MC = AngleEntry{Longitude: houses.MC, Sign: mcSign, SignDegree: mcDeg}
	r.ARMC = houses.ARMC
	r.Vertex = AngleEntry{Longitude: houses.Vertex, Sign: vtxSign, SignDegree: vtxDeg}

	for i := 1; i <= 12; i++ {
		sign, deg := swisseph.ZodiacSign(houses.Cusps[i])
		r.Cusps = append(r.Cusps, CuspEntry{
			House:      i,
			Longitude:  houses.Cusps[i],
			Sign:       sign,
			SignDegree: deg,
		})
	}

	return r, nil
}
