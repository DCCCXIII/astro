package output

import (
	"fmt"
	"strings"
)

// planetNameWidth is the column width used for planet names in the primary
// planet line ("%-10s"). The verbose continuation line uses the same width
// as blank padding to keep the two lines visually aligned.
const planetNameWidth = 10

// PrintText writes a human-readable report of planetary positions and house
// cusps to stdout. When verbose is true, additional raw data is included
// (ecliptic latitude, distance, latitude/distance speeds, ARMC, Vertex).
func PrintText(r Result, verbose bool) error {
	fmt.Printf("Julian Day: %.6f\n", r.JulianDay)
	if verbose && r.EphemerisWarning != "" {
		fmt.Printf("Ephemeris:  %s\n", r.EphemerisWarning)
	}
	fmt.Println()

	fmt.Println("=== Planetary Positions ===")
	for _, p := range r.Planets {
		fmt.Printf("%-*s  %9.4f°  (%s %5.2f°)  speed: %+.4f°/day\n",
			planetNameWidth, p.Name, p.Longitude, p.Sign, p.SignDegree, p.Speed)
		if verbose {
			fmt.Printf("%-*s  lat: %+.6f°  dist: %.8f AU  speed_lat: %+.6f°/day  speed_dist: %+.8f AU/day\n",
				planetNameWidth, "", p.Latitude, p.Distance, p.SpeedLat, p.SpeedDistance)
		}
	}

	fmt.Printf("\n=== Houses (%s) for (%.4f°, %.4f°) ===\n", r.HouseName, r.Lat, r.Lon)
	fmt.Printf("Ascendant:  %9.4f°  (%s %.2f°)\n", r.Ascendant.Longitude, r.Ascendant.Sign, r.Ascendant.SignDegree)
	fmt.Printf("MC:         %9.4f°  (%s %.2f°)\n", r.MC.Longitude, r.MC.Sign, r.MC.SignDegree)
	if verbose {
		fmt.Printf("ARMC:       %9.4f°\n", r.ARMC)
		fmt.Printf("Vertex:     %9.4f°  (%s %.2f°)\n", r.Vertex.Longitude, r.Vertex.Sign, r.Vertex.SignDegree)
	}

	fmt.Println("\nHouse cusps:")
	for _, c := range r.Cusps {
		fmt.Printf("  House %2d: %9.4f°  (%s %.2f°)\n", c.House, c.Longitude, c.Sign, c.SignDegree)
	}

	for _, a := range r.Almuten {
		fmt.Printf("\n=== Almuten: %s ===\n", a.Method)
		victor := "none"
		if len(a.Winners) > 0 {
			victor = strings.Join(a.Winners, ", ")
			if len(a.Winners) > 1 {
				victor += " (tie)"
			}
		}
		fmt.Printf("Victor: %s\n", victor)
		if verbose {
			for _, sc := range a.Scoreboard {
				fmt.Printf("  %-*s  %+d\n", planetNameWidth, sc.Planet, sc.Score)
			}
		}
	}

	if r.Search != nil {
		s := r.Search
		fmt.Printf("\n=== Search: %s at %s %.2f° ===\n", s.Planet, s.TargetSign, s.TargetSignDeg)
		label := "Last occurrence"
		if s.Direction == "forward" {
			label = "Next occurrence"
		}
		fmt.Printf("%s: %04d-%02d-%02d %02d:%02d UTC  (JD %.6f)\n",
			label, s.Year, s.Month, s.Day, int(s.Hour), int((s.Hour-float64(int(s.Hour)))*60), s.JulianDay)
		if verbose {
			dir := "direct"
			if s.Retrograde {
				dir = "retrograde"
			}
			fmt.Printf("Speed: %+.4f°/day (%s)\n", s.SpeedLon, dir)
			fmt.Printf("House: %d\n", s.House)
		}
	}
	return nil
}
