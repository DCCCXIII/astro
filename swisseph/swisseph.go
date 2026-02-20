// Package swisseph provides Go bindings for the Swiss Ephemeris C library.
// It exposes functions for calculating planetary positions and house cusps/angles
// (including the Ascendant) for a given point in time and geographic location.
package swisseph

/*
#cgo CFLAGS: -w
#cgo LDFLAGS: -lm
#include "swephexp.h"
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"sync"
	"unsafe"
)

// Planet identifiers for the traditional planets.
const (
	Sun     = C.SE_SUN
	Moon    = C.SE_MOON
	Mercury = C.SE_MERCURY
	Venus   = C.SE_VENUS
	Mars    = C.SE_MARS
	Jupiter = C.SE_JUPITER
	Saturn  = C.SE_SATURN
)

// House system codes (passed as a single character).
const (
	HousePlacidus      = 'P'
	HouseKoch          = 'K'
	HouseWholeSign     = 'W'
	HouseRegiomontanus = 'R'
	HouseEqual         = 'A'
	HouseCampanus      = 'C'
)

// mu protects the Swiss Ephemeris global state from concurrent access.
var mu sync.Mutex

// SetEphePath tells the library where to find the .se1 ephemeris data files.
// If path is empty, the library falls back to the Moshier ephemeris (lower
// precision but needs no external files).
func SetEphePath(path string) {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	mu.Lock()
	defer mu.Unlock()
	C.swe_set_ephe_path(cpath)
}

// Close frees all resources allocated by the library. Call this when done.
func Close() {
	mu.Lock()
	defer mu.Unlock()
	C.swe_close()
}

// PlanetName returns the human-readable name for a planet ID.
func PlanetName(planet int) string {
	var buf [256]C.char
	C.swe_get_planet_name(C.int(planet), &buf[0])
	return C.GoString(&buf[0])
}

// JulDay converts a calendar date and time (UTC) to a Julian Day number,
// which is the time format used internally by the library.
// hour is in decimal form (e.g. 14.5 means 2:30 PM).
func JulDay(year, month, day int, hour float64) float64 {
	return float64(C.swe_julday(
		C.int(year), C.int(month), C.int(day),
		C.double(hour), C.SE_GREG_CAL,
	))
}

// PlanetPos holds the result of a planetary position calculation.
type PlanetPos struct {
	Longitude     float64 // ecliptic longitude in degrees (0-360)
	Latitude      float64 // ecliptic latitude in degrees
	Distance      float64 // distance from Earth in AU
	SpeedLon      float64 // daily speed in longitude (degrees/day)
	SpeedLat      float64 // daily speed in latitude (degrees/day)
	SpeedDistance float64 // daily speed in distance (AU/day)
}

// CalcPlanet calculates the position of a planet at the given Julian Day (UT).
// Use the planet constants (Sun, Moon, Mercury, etc.) for the planet argument.
func CalcPlanet(tjdUT float64, planet int) (PlanetPos, error) {
	var xx [6]C.double
	var serr [256]C.char

	mu.Lock()
	ret := C.swe_calc_ut(
		C.double(tjdUT),
		C.int(planet),
		C.SEFLG_SWIEPH|C.SEFLG_SPEED,
		&xx[0],
		&serr[0],
	)
	mu.Unlock()

	if int(ret) < 0 {
		return PlanetPos{}, fmt.Errorf("swe_calc_ut: %s", C.GoString(&serr[0]))
	}

	return PlanetPos{
		Longitude:     float64(xx[0]),
		Latitude:      float64(xx[1]),
		Distance:      float64(xx[2]),
		SpeedLon:      float64(xx[3]),
		SpeedLat:      float64(xx[4]),
		SpeedDistance: float64(xx[5]),
	}, nil
}

// HouseResult holds the result of a house calculation.
type HouseResult struct {
	Cusps     [13]float64 // house cusps in degrees; index 1-12 are houses 1-12 (index 0 is unused)
	Ascendant float64     // Ascendant in degrees
	MC        float64     // Midheaven (Medium Coeli) in degrees
	ARMC      float64     // sidereal time in degrees
	Vertex    float64     // Vertex in degrees
}

// CalcHouses calculates house cusps and angles for a given time and location.
// geoLat and geoLon are geographic latitude and longitude in degrees
// (north and east are positive). hsys is a house system code (use the
// House* constants).
func CalcHouses(tjdUT float64, geoLat, geoLon float64, hsys byte) (HouseResult, error) {
	var cusps [13]C.double
	var ascmc [10]C.double

	mu.Lock()
	ret := C.swe_houses(
		C.double(tjdUT),
		C.double(geoLat),
		C.double(geoLon),
		C.int(hsys),
		&cusps[0],
		&ascmc[0],
	)
	mu.Unlock()

	if int(ret) < 0 {
		return HouseResult{}, fmt.Errorf("swe_houses failed (return code %d)", int(ret))
	}

	var result HouseResult
	for i := 0; i < 13; i++ {
		result.Cusps[i] = float64(cusps[i])
	}
	result.Ascendant = float64(ascmc[0])
	result.MC = float64(ascmc[1])
	result.ARMC = float64(ascmc[2])
	result.Vertex = float64(ascmc[3])
	return result, nil
}

// ZodiacSign returns the zodiac sign name and degree within that sign
// for a given ecliptic longitude (0-360).
func ZodiacSign(longitude float64) (sign string, degrees float64) {
	signs := [12]string{
		"Aries", "Taurus", "Gemini", "Cancer",
		"Leo", "Virgo", "Libra", "Scorpio",
		"Sagittarius", "Capricorn", "Aquarius", "Pisces",
	}
	idx := int(longitude / 30.0)
	if idx >= 12 {
		idx = 11
	}
	return signs[idx], longitude - float64(idx)*30.0
}
