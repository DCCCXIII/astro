package output

import (
	"encoding/json"
	"fmt"
)

type planetJSON struct {
	Name          string   `json:"name"`
	Longitude     float64  `json:"longitude"`
	Sign          string   `json:"sign"`
	SignDegree    float64  `json:"sign_degree"`
	Speed         float64  `json:"speed"`
	Latitude      *float64 `json:"latitude,omitempty"`
	Distance      *float64 `json:"distance,omitempty"`
	SpeedLat      *float64 `json:"speed_lat,omitempty"`
	SpeedDistance *float64 `json:"speed_distance,omitempty"`
}

type housesJSON struct {
	System    string      `json:"system"`
	Ascendant AngleEntry  `json:"ascendant"`
	MC        AngleEntry  `json:"mc"`
	ARMC      *float64    `json:"armc,omitempty"`
	Vertex    *AngleEntry `json:"vertex,omitempty"`
	Cusps     []CuspEntry `json:"cusps"`
}

type resultJSON struct {
	JulianDay float64      `json:"julian_day"`
	Planets   []planetJSON `json:"planets"`
	Houses    housesJSON   `json:"houses"`
}

// PrintJSON writes planetary positions and house cusps as indented JSON to
// stdout. When verbose is true, additional raw fields are included (ecliptic
// latitude, distance, latitude/distance speeds, ARMC, Vertex).
func PrintJSON(r Result, verbose bool) error {
	planets := make([]planetJSON, len(r.Planets))
	for i, p := range r.Planets {
		entry := planetJSON{
			Name:       p.Name,
			Longitude:  p.Longitude,
			Sign:       p.Sign,
			SignDegree: p.SignDegree,
			Speed:      p.Speed,
		}
		if verbose {
			entry.Latitude = &p.Latitude
			entry.Distance = &p.Distance
			entry.SpeedLat = &p.SpeedLat
			entry.SpeedDistance = &p.SpeedDistance
		}
		planets[i] = entry
	}

	houses := housesJSON{
		System:    r.HouseName,
		Ascendant: r.Ascendant,
		MC:        r.MC,
		Cusps:     r.Cusps,
	}
	if verbose {
		houses.ARMC = &r.ARMC
		vtx := r.Vertex
		houses.Vertex = &vtx
	}

	out := resultJSON{
		JulianDay: r.JulianDay,
		Planets:   planets,
		Houses:    houses,
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}
	fmt.Printf("%s\n", data)
	return nil
}
