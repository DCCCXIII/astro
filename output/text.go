package output

import "fmt"

// planetNameWidth is the column width used for planet names in the primary
// planet line ("%-10s"). The verbose continuation line uses the same width
// as blank padding to keep the two lines visually aligned.
const planetNameWidth = 10

// PrintText writes a human-readable report of planetary positions and house
// cusps to stdout. When verbose is true, additional raw data is included
// (ecliptic latitude, distance, latitude/distance speeds, ARMC, Vertex).
func PrintText(r Result, verbose bool) error {
	fmt.Printf("Julian Day: %.6f\n\n", r.JulianDay)

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
	return nil
}
