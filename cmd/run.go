package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dcccxiii/astro/output"
	"github.com/dcccxiii/astro/swisseph"
)

// noSearch is the "no --search flag was given" sentinel for the planet
// return of parseSearchTarget. 0 can't serve as the sentinel because it
// collides with swisseph.Sun.
const noSearch = -1

var searchPlanets = map[string]int{
	"sun":     swisseph.Sun,
	"moon":    swisseph.Moon,
	"mercury": swisseph.Mercury,
	"venus":   swisseph.Venus,
	"mars":    swisseph.Mars,
	"jupiter": swisseph.Jupiter,
	"saturn":  swisseph.Saturn,
}

var zodiacSigns = []string{
	"aries", "taurus", "gemini", "cancer", "leo", "virgo",
	"libra", "scorpio", "sagittarius", "capricorn", "aquarius", "pisces",
}

var signDegreeRe = regexp.MustCompile(`^(\d+(?:\.\d+)?)([a-zA-Z]+)$`)

// Run is the CLI entry point. It parses args, sets up the ephemeris, and
// delegates rendering to the output package.
func Run(args []string) error {
	fs := flag.NewFlagSet("astro", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: astro [--house-system <system>] [--almuten <mode>] [--search <planet>:<longitude>[:forward|backward]] [--json] [--verbose] <datetime> <lat> <lon>\n")
		fmt.Fprintf(fs.Output(), "  <datetime>  ISO 8601 date/time in UTC, e.g. 2024-03-20T12:00:00Z\n")
		fmt.Fprintf(fs.Output(), "  <lat>       geographic latitude in decimal degrees (north = positive)\n")
		fmt.Fprintf(fs.Output(), "  <lon>       geographic longitude in decimal degrees (east = positive)\n\n")
		fs.PrintDefaults()
	}

	houseSystemFlag := fs.String("house-system", "placidus", "House system: placidus, koch, whole-sign, regiomontanus, equal, campanus")
	almutenFlag := fs.String("almuten", "", "Chart victor: geniture (Lilly's Lord of the Geniture), figuris (Ibn Ezra), bonatti (Bonatti's Almudebit), both (geniture+figuris), or all")
	searchFlag := fs.String("search", "", "Find when a planet was/will be at a given ecliptic longitude, relative to <datetime>: <planet>:<longitude>[:forward|backward] (default backward), e.g. mars:15aries, mars:195, or mars:15aries:forward")
	jsonFlag := fs.Bool("json", false, "Output results as JSON")
	verboseFlag := fs.Bool("verbose", false, "Verbose output: include ecliptic latitude, distance, speed components, ARMC, and Vertex")

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}

	if fs.NArg() != 3 {
		fs.Usage()
		return fmt.Errorf("expected 3 arguments, got %d", fs.NArg())
	}

	t, err := time.Parse(time.RFC3339, fs.Arg(0))
	if err != nil {
		return fmt.Errorf("invalid datetime %q: %w", fs.Arg(0), err)
	}
	t = t.UTC()

	lat, err := strconv.ParseFloat(fs.Arg(1), 64)
	if err != nil {
		return fmt.Errorf("invalid latitude %q: %w", fs.Arg(1), err)
	}

	lon, err := strconv.ParseFloat(fs.Arg(2), 64)
	if err != nil {
		return fmt.Errorf("invalid longitude %q: %w", fs.Arg(2), err)
	}

	hsys, hsysName, err := parseHouseSystem(*houseSystemFlag)
	if err != nil {
		return err
	}

	almutenMode, err := parseAlmutenMode(*almutenFlag)
	if err != nil {
		return err
	}

	searchPlanet, searchLon, searchDir, err := parseSearchTarget(*searchFlag)
	if err != nil {
		return err
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not resolve executable path: %w", err)
	}
	swisseph.SetEphePath(filepath.Join(filepath.Dir(exe), "ephe"))
	defer swisseph.Close()

	decimalHour := float64(t.Hour()) + float64(t.Minute())/60 + float64(t.Second())/3600
	jd := swisseph.JulDay(t.Year(), int(t.Month()), t.Day(), decimalHour)

	if searchPlanet != noSearch {
		jd, err = output.FindSearchJD(jd, searchPlanet, searchLon, searchDir)
		if err != nil {
			return err
		}
	}

	planets := []int{
		swisseph.Sun, swisseph.Moon, swisseph.Mercury,
		swisseph.Venus, swisseph.Mars, swisseph.Jupiter,
		swisseph.Saturn,
	}

	r, err := output.Build(jd, planets, lat, lon, hsys, hsysName)
	if err != nil {
		return err
	}

	if almutenMode != "" {
		r.Almuten, err = output.BuildAlmuten(jd, lat, lon, hsys, almutenMode)
		if err != nil {
			return err
		}
	}

	if *jsonFlag {
		return output.PrintJSON(r, *verboseFlag)
	}
	return output.PrintText(r, *verboseFlag)
}

// parseAlmutenMode validates the --almuten flag value. An empty value means no
// almuten calculation was requested.
func parseAlmutenMode(mode string) (string, error) {
	switch strings.ToLower(mode) {
	case "":
		return "", nil
	case "geniture", "figuris", "bonatti", "both", "all":
		return strings.ToLower(mode), nil
	default:
		return "", fmt.Errorf("unknown almuten mode %q: valid values are geniture, figuris, bonatti, both, all", mode)
	}
}

// parseSearchTarget validates the --search flag value
// "<planet>:<longitude>[:<direction>]". An empty value means no search was
// requested, signalled by returning noSearch for planet. longitude may be
// given as raw ecliptic degrees in [0, 360) (e.g. "195") or as sign+degree
// (e.g. "15aries"). direction is "forward" or "backward" (case-insensitive);
// omitting the third segment defaults to "backward".
func parseSearchTarget(spec string) (planet int, longitude float64, direction string, err error) {
	if spec == "" {
		return noSearch, 0, "", nil
	}

	parts := strings.SplitN(spec, ":", 3)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return noSearch, 0, "", fmt.Errorf("invalid --search value %q: expected <planet>:<longitude>[:<direction>], e.g. mars:15aries, mars:195, or mars:15aries:forward", spec)
	}

	p, ok := searchPlanets[strings.ToLower(parts[0])]
	if !ok {
		return noSearch, 0, "", fmt.Errorf("unknown --search planet %q: valid values are sun, moon, mercury, venus, mars, jupiter, saturn", parts[0])
	}

	lon, err := parseSearchLongitude(parts[1])
	if err != nil {
		return noSearch, 0, "", fmt.Errorf("invalid --search longitude %q: %w", parts[1], err)
	}

	dirSpec := ""
	if len(parts) == 3 {
		dirSpec = parts[2]
	}
	dir, err := parseSearchDirection(dirSpec)
	if err != nil {
		return noSearch, 0, "", err
	}

	return p, lon, dir, nil
}

// parseSearchDirection validates the optional third --search segment. An
// empty value defaults to "backward", preserving the two-segment spec's
// pre-existing documented behavior unchanged.
func parseSearchDirection(direction string) (string, error) {
	switch strings.ToLower(direction) {
	case "":
		return "backward", nil
	case "forward", "backward":
		return strings.ToLower(direction), nil
	default:
		return "", fmt.Errorf("unknown --search direction %q: valid values are forward, backward", direction)
	}
}

// parseSearchLongitude parses either raw ecliptic degrees in [0, 360) (e.g.
// "195") or sign+degree notation (e.g. "15aries", degrees in [0, 30)).
func parseSearchLongitude(s string) (float64, error) {
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		if v < 0 || v >= 360 {
			return 0, fmt.Errorf("raw degrees must be in [0, 360), got %v", v)
		}
		return v, nil
	}

	m := signDegreeRe.FindStringSubmatch(s)
	if m == nil {
		return 0, fmt.Errorf("expected raw degrees (e.g. 195) or <degrees><sign> (e.g. 15aries)")
	}
	deg, err := strconv.ParseFloat(m[1], 64)
	if err != nil || deg < 0 || deg >= 30 {
		return 0, fmt.Errorf("sign degree must be in [0, 30), got %s", m[1])
	}
	signName := strings.ToLower(m[2])
	for i, name := range zodiacSigns {
		if name == signName {
			return float64(i)*30 + deg, nil
		}
	}
	return 0, fmt.Errorf("unknown zodiac sign %q", m[2])
}

func parseHouseSystem(name string) (code byte, displayName string, err error) {
	switch strings.ToLower(name) {
	case "placidus":
		return swisseph.HousePlacidus, "Placidus", nil
	case "koch":
		return swisseph.HouseKoch, "Koch", nil
	case "whole-sign":
		return swisseph.HouseWholeSign, "Whole Sign", nil
	case "regiomontanus":
		return swisseph.HouseRegiomontanus, "Regiomontanus", nil
	case "equal":
		return swisseph.HouseEqual, "Equal", nil
	case "campanus":
		return swisseph.HouseCampanus, "Campanus", nil
	default:
		return 0, "", fmt.Errorf("unknown house system %q: valid values are placidus, koch, whole-sign, regiomontanus, equal, campanus", name)
	}
}
