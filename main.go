package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dcccxiii/astro/swisseph"
)

func main() {
	houseSystemFlag := flag.String("house-system", "placidus", "House system: placidus, koch, whole-sign, regiomontanus, equal, campanus")
	jsonFlag := flag.Bool("json", false, "Output results as JSON")
	flag.Parse()

	if flag.NArg() != 3 {
		fmt.Fprintf(os.Stderr, "Usage: astro [--house-system <system>] [--json] <datetime> <lat> <lon>\n")
		fmt.Fprintf(os.Stderr, "  <datetime>  ISO 8601 date/time in UTC, e.g. 2024-03-20T12:00:00Z\n")
		fmt.Fprintf(os.Stderr, "  <lat>       geographic latitude in decimal degrees (north = positive)\n")
		fmt.Fprintf(os.Stderr, "  <lon>       geographic longitude in decimal degrees (east = positive)\n")
		os.Exit(1)
	}

	t, err := time.Parse(time.RFC3339, flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing datetime %q: %v\n", flag.Arg(0), err)
		os.Exit(1)
	}
	t = t.UTC()

	lat, err := strconv.ParseFloat(flag.Arg(1), 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing latitude %q: %v\n", flag.Arg(1), err)
		os.Exit(1)
	}

	lon, err := strconv.ParseFloat(flag.Arg(2), 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing longitude %q: %v\n", flag.Arg(2), err)
		os.Exit(1)
	}

	hsys, err := parseHouseSystem(*houseSystemFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	exe, _ := os.Executable()
	ephePath := filepath.Join(filepath.Dir(exe), "ephe")
	swisseph.SetEphePath(ephePath)
	defer swisseph.Close()

	decimalHour := float64(t.Hour()) + float64(t.Minute())/60 + float64(t.Second())/3600
	jd := swisseph.JulDay(t.Year(), int(t.Month()), t.Day(), decimalHour)

	planets := []int{
		swisseph.Sun, swisseph.Moon, swisseph.Mercury,
		swisseph.Venus, swisseph.Mars, swisseph.Jupiter,
		swisseph.Saturn,
	}

	if *jsonFlag {
		printJSON(jd, planets, lat, lon, hsys, *houseSystemFlag)
	} else {
		printText(jd, planets, lat, lon, hsys, *houseSystemFlag)
	}
}

func parseHouseSystem(name string) (byte, error) {
	switch strings.ToLower(name) {
	case "placidus":
		return swisseph.HousePlacidus, nil
	case "koch":
		return swisseph.HouseKoch, nil
	case "whole-sign":
		return swisseph.HouseWholeSign, nil
	case "regiomontanus":
		return swisseph.HouseRegiomontanus, nil
	case "equal":
		return swisseph.HouseEqual, nil
	case "campanus":
		return swisseph.HouseCampanus, nil
	default:
		return 0, fmt.Errorf("unknown house system %q: valid values are placidus, koch, whole-sign, regiomontanus, equal, campanus", name)
	}
}

func printText(jd float64, planets []int, lat, lon float64, hsys byte, hsysName string) {
	fmt.Printf("Julian Day: %.6f\n\n", jd)

	fmt.Println("=== Planetary Positions ===")
	for _, p := range planets {
		pos, err := swisseph.CalcPlanet(jd, p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error calculating %s: %v\n", swisseph.PlanetName(p), err)
			continue
		}
		sign, deg := swisseph.ZodiacSign(pos.Longitude)
		fmt.Printf("%-10s  %9.4f°  (%s %5.2f°)  speed: %+.4f°/day\n",
			swisseph.PlanetName(p), pos.Longitude, sign, deg, pos.SpeedLon)
	}

	fmt.Printf("\n=== Houses (%s) for (%.4f°, %.4f°) ===\n", hsysName, lat, lon)

	houses, err := swisseph.CalcHouses(jd, lat, lon, hsys)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error calculating houses: %v\n", err)
		os.Exit(1)
	}

	ascSign, ascDeg := swisseph.ZodiacSign(houses.Ascendant)
	mcSign, mcDeg := swisseph.ZodiacSign(houses.MC)
	fmt.Printf("Ascendant:  %9.4f°  (%s %.2f°)\n", houses.Ascendant, ascSign, ascDeg)
	fmt.Printf("MC:         %9.4f°  (%s %.2f°)\n", houses.MC, mcSign, mcDeg)

	fmt.Println("\nHouse cusps:")
	for i := 1; i <= 12; i++ {
		sign, deg := swisseph.ZodiacSign(houses.Cusps[i])
		fmt.Printf("  House %2d: %9.4f°  (%s %.2f°)\n", i, houses.Cusps[i], sign, deg)
	}
}

type planetJSON struct {
	Name       string  `json:"name"`
	Longitude  float64 `json:"longitude"`
	Sign       string  `json:"sign"`
	SignDegree float64 `json:"sign_degree"`
	Speed      float64 `json:"speed"`
}

type angleJSON struct {
	Longitude  float64 `json:"longitude"`
	Sign       string  `json:"sign"`
	SignDegree float64 `json:"sign_degree"`
}

type cuspJSON struct {
	House      int     `json:"house"`
	Longitude  float64 `json:"longitude"`
	Sign       string  `json:"sign"`
	SignDegree float64 `json:"sign_degree"`
}

type housesJSON struct {
	System    string     `json:"system"`
	Ascendant angleJSON  `json:"ascendant"`
	MC        angleJSON  `json:"mc"`
	Cusps     []cuspJSON `json:"cusps"`
}

type resultJSON struct {
	JulianDay float64      `json:"julian_day"`
	Planets   []planetJSON `json:"planets"`
	Houses    housesJSON   `json:"houses"`
}

func printJSON(jd float64, planets []int, lat, lon float64, hsys byte, hsysName string) {
	out := resultJSON{JulianDay: jd}

	for _, p := range planets {
		pos, err := swisseph.CalcPlanet(jd, p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error calculating %s: %v\n", swisseph.PlanetName(p), err)
			os.Exit(1)
		}
		sign, deg := swisseph.ZodiacSign(pos.Longitude)
		out.Planets = append(out.Planets, planetJSON{
			Name:       swisseph.PlanetName(p),
			Longitude:  pos.Longitude,
			Sign:       sign,
			SignDegree: deg,
			Speed:      pos.SpeedLon,
		})
	}

	houses, err := swisseph.CalcHouses(jd, lat, lon, hsys)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error calculating houses: %v\n", err)
		os.Exit(1)
	}

	ascSign, ascDeg := swisseph.ZodiacSign(houses.Ascendant)
	mcSign, mcDeg := swisseph.ZodiacSign(houses.MC)

	out.Houses = housesJSON{
		System:    hsysName,
		Ascendant: angleJSON{Longitude: houses.Ascendant, Sign: ascSign, SignDegree: ascDeg},
		MC:        angleJSON{Longitude: houses.MC, Sign: mcSign, SignDegree: mcDeg},
	}

	for i := 1; i <= 12; i++ {
		sign, deg := swisseph.ZodiacSign(houses.Cusps[i])
		out.Houses.Cusps = append(out.Houses.Cusps, cuspJSON{
			House:      i,
			Longitude:  houses.Cusps[i],
			Sign:       sign,
			SignDegree: deg,
		})
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshalling JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}
