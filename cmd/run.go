package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dcccxiii/astro/output"
	"github.com/dcccxiii/astro/swisseph"
)

// Run is the CLI entry point. It parses args, sets up the ephemeris, and
// delegates rendering to the output package.
func Run(args []string) error {
	fs := flag.NewFlagSet("astro", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: astro [--house-system <system>] [--json] <datetime> <lat> <lon>\n")
		fmt.Fprintf(fs.Output(), "  <datetime>  ISO 8601 date/time in UTC, e.g. 2024-03-20T12:00:00Z\n")
		fmt.Fprintf(fs.Output(), "  <lat>       geographic latitude in decimal degrees (north = positive)\n")
		fmt.Fprintf(fs.Output(), "  <lon>       geographic longitude in decimal degrees (east = positive)\n\n")
		fs.PrintDefaults()
	}

	houseSystemFlag := fs.String("house-system", "placidus", "House system: placidus, koch, whole-sign, regiomontanus, equal, campanus")
	jsonFlag := fs.Bool("json", false, "Output results as JSON")

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

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not resolve executable path: %w", err)
	}
	swisseph.SetEphePath(filepath.Join(filepath.Dir(exe), "ephe"))
	defer swisseph.Close()

	decimalHour := float64(t.Hour()) + float64(t.Minute())/60 + float64(t.Second())/3600
	jd := swisseph.JulDay(t.Year(), int(t.Month()), t.Day(), decimalHour)

	planets := []int{
		swisseph.Sun, swisseph.Moon, swisseph.Mercury,
		swisseph.Venus, swisseph.Mars, swisseph.Jupiter,
		swisseph.Saturn,
	}

	r, err := output.Build(jd, planets, lat, lon, hsys, hsysName)
	if err != nil {
		return err
	}

	if *jsonFlag {
		return output.PrintJSON(r)
	}
	return output.PrintText(r)
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
