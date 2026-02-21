package output

import "fmt"

// PrintText writes a human-readable report of planetary positions and house
// cusps to stdout.
func PrintText(r Result) error {
	fmt.Printf("Julian Day: %.6f\n\n", r.JulianDay)

	fmt.Println("=== Planetary Positions ===")
	for _, p := range r.Planets {
		fmt.Printf("%-10s  %9.4f°  (%s %5.2f°)  speed: %+.4f°/day\n",
			p.Name, p.Longitude, p.Sign, p.SignDegree, p.Speed)
	}

	fmt.Printf("\n=== Houses (%s) for (%.4f°, %.4f°) ===\n", r.HouseName, r.Lat, r.Lon)
	fmt.Printf("Ascendant:  %9.4f°  (%s %.2f°)\n", r.Ascendant.Longitude, r.Ascendant.Sign, r.Ascendant.SignDegree)
	fmt.Printf("MC:         %9.4f°  (%s %.2f°)\n", r.MC.Longitude, r.MC.Sign, r.MC.SignDegree)

	fmt.Println("\nHouse cusps:")
	for _, c := range r.Cusps {
		fmt.Printf("  House %2d: %9.4f°  (%s %.2f°)\n", c.House, c.Longitude, c.Sign, c.SignDegree)
	}
	return nil
}
