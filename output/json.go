package output

import (
	"encoding/json"
	"fmt"
)

type housesJSON struct {
	System    string      `json:"system"`
	Ascendant AngleEntry  `json:"ascendant"`
	MC        AngleEntry  `json:"mc"`
	Cusps     []CuspEntry `json:"cusps"`
}

type resultJSON struct {
	JulianDay float64       `json:"julian_day"`
	Planets   []PlanetEntry `json:"planets"`
	Houses    housesJSON    `json:"houses"`
}

// PrintJSON writes planetary positions and house cusps as indented JSON to stdout.
func PrintJSON(r Result) error {
	out := resultJSON{
		JulianDay: r.JulianDay,
		Planets:   r.Planets,
		Houses: housesJSON{
			System:    r.HouseName,
			Ascendant: r.Ascendant,
			MC:        r.MC,
			Cusps:     r.Cusps,
		},
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}
