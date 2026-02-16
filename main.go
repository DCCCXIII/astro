package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rubenrm/astro/swisseph"
)

func main() {
	// Point the library to the ephemeris data files.
	// The path is relative to where you run the binary from.
	exe, _ := os.Executable()
	ephePath := filepath.Join(filepath.Dir(exe), "ephe")
	swisseph.SetEphePath(ephePath)
	defer swisseph.Close()

	// Example: calculate positions for 2024-Mar-20 at 12:00 UTC
	// (approximately the spring equinox)
	jd := swisseph.JulDay(2024, 3, 20, 12.0)
	fmt.Printf("Julian Day: %.6f\n\n", jd)

	// Calculate positions for each traditional planet.
	planets := []int{
		swisseph.Sun, swisseph.Moon, swisseph.Mercury,
		swisseph.Venus, swisseph.Mars, swisseph.Jupiter,
		swisseph.Saturn,
	}

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

	// Calculate the Ascendant and house cusps for New York City.
	// Latitude 40.7128° N, Longitude 74.0060° W (west is negative).
	lat := 40.7128
	lon := -74.0060
	fmt.Printf("\n=== Houses (Placidus) for New York City (%.4f°N, %.4f°W) ===\n", lat, -lon)

	houses, err := swisseph.CalcHouses(jd, lat, lon, swisseph.HousePlacidus)
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
