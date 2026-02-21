package output

import (
	"fmt"

	"github.com/dcccxiii/astro/swisseph"
)

// PlanetEntry holds presentation-ready data for a single planet.
type PlanetEntry struct {
	Name       string  `json:"name"`
	Longitude  float64 `json:"longitude"`
	Sign       string  `json:"sign"`
	SignDegree float64 `json:"sign_degree"`
	Speed      float64 `json:"speed"`
}

// AngleEntry holds presentation-ready data for a chart angle (Ascendant, MC).
type AngleEntry struct {
	Longitude  float64 `json:"longitude"`
	Sign       string  `json:"sign"`
	SignDegree float64 `json:"sign_degree"`
}

// CuspEntry holds presentation-ready data for a single house cusp.
type CuspEntry struct {
	House      int     `json:"house"`
	Longitude  float64 `json:"longitude"`
	Sign       string  `json:"sign"`
	SignDegree float64 `json:"sign_degree"`
}

// Result holds all computed, presentation-ready chart data. Both PrintText
// and PrintJSON render from this struct; neither calls swisseph directly.
type Result struct {
	JulianDay float64
	HouseName string
	Lat       float64
	Lon       float64
	Planets   []PlanetEntry
	Ascendant AngleEntry
	MC        AngleEntry
	Cusps     []CuspEntry // one entry per house, 1-12
}

// Build computes a full chart result for the given Julian Day, planets, and
// geographic location. All swisseph calls are concentrated here.
func Build(jd float64, planets []int, lat, lon float64, hsys byte, hsysName string) (Result, error) {
	r := Result{JulianDay: jd, HouseName: hsysName, Lat: lat, Lon: lon}

	for _, p := range planets {
		name := swisseph.PlanetName(p)
		pos, err := swisseph.CalcPlanet(jd, p)
		if err != nil {
			return Result{}, fmt.Errorf("error calculating %s: %w", name, err)
		}
		sign, deg := swisseph.ZodiacSign(pos.Longitude)
		r.Planets = append(r.Planets, PlanetEntry{
			Name:       name,
			Longitude:  pos.Longitude,
			Sign:       sign,
			SignDegree: deg,
			Speed:      pos.SpeedLon,
		})
	}

	houses, err := swisseph.CalcHouses(jd, lat, lon, hsys)
	if err != nil {
		return Result{}, fmt.Errorf("error calculating houses: %w", err)
	}

	ascSign, ascDeg := swisseph.ZodiacSign(houses.Ascendant)
	mcSign, mcDeg := swisseph.ZodiacSign(houses.MC)
	r.Ascendant = AngleEntry{Longitude: houses.Ascendant, Sign: ascSign, SignDegree: ascDeg}
	r.MC = AngleEntry{Longitude: houses.MC, Sign: mcSign, SignDegree: mcDeg}

	for i := 1; i <= 12; i++ {
		sign, deg := swisseph.ZodiacSign(houses.Cusps[i])
		r.Cusps = append(r.Cusps, CuspEntry{
			House:      i,
			Longitude:  houses.Cusps[i],
			Sign:       sign,
			SignDegree: deg,
		})
	}

	return r, nil
}
