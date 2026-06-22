package output

import (
	"encoding/json"
	"fmt"
	"math"
	"time"
)

type searchJSON struct {
	Planet     string   `json:"planet"`
	TargetLon  float64  `json:"target_longitude"`
	TargetSign string   `json:"target_sign"`
	TargetDeg  float64  `json:"target_sign_degree"`
	Direction  string   `json:"direction"`
	Timestamp  string   `json:"timestamp"`
	JulianDay  float64  `json:"julian_day"`
	SpeedLon   *float64 `json:"speed,omitempty"`
	Retrograde *bool    `json:"retrograde,omitempty"`
	House      *int     `json:"house,omitempty"`
}

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

type almutenJSON struct {
	Method     string         `json:"method"`
	Winners    []string       `json:"winners"`
	Scoreboard []AlmutenScore `json:"scoreboard,omitempty"`
}

type resultJSON struct {
	JulianDay        float64       `json:"julian_day"`
	Planets          []planetJSON  `json:"planets"`
	Houses           housesJSON    `json:"houses"`
	Almuten          []almutenJSON `json:"almuten,omitempty"`
	Search           *searchJSON   `json:"search,omitempty"`
	EphemerisWarning *string       `json:"ephemeris_warning,omitempty"`
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
	for _, a := range r.Almuten {
		entry := almutenJSON{Method: a.Method, Winners: a.Winners}
		if verbose {
			entry.Scoreboard = a.Scoreboard
		}
		out.Almuten = append(out.Almuten, entry)
	}
	if r.Search != nil {
		s := r.Search
		hh := int(s.Hour)
		mm := int((s.Hour - float64(hh)) * 60)
		ss := int(math.Round((s.Hour - float64(hh) - float64(mm)/60) * 3600))
		ts := time.Date(s.Year, time.Month(s.Month), s.Day, hh, mm, ss, 0, time.UTC)
		entry := searchJSON{
			Planet:     s.Planet,
			TargetLon:  s.TargetLon,
			TargetSign: s.TargetSign,
			TargetDeg:  s.TargetSignDeg,
			Direction:  s.Direction,
			Timestamp:  ts.Format(time.RFC3339),
			JulianDay:  s.JulianDay,
		}
		if verbose {
			entry.SpeedLon = &s.SpeedLon
			entry.Retrograde = &s.Retrograde
			entry.House = &s.House
		}
		out.Search = &entry
	}

	if verbose && r.EphemerisWarning != "" {
		out.EphemerisWarning = &r.EphemerisWarning
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}
	fmt.Printf("%s\n", data)
	return nil
}
