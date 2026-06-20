package cmd

import (
	"testing"

	"github.com/dcccxiii/astro/swisseph"
)

func TestParseHouseSystem(t *testing.T) {
	cases := []struct {
		input       string
		want        byte
		wantDisplay string
		wantErr     bool
	}{
		// All valid names, canonical casing
		{"placidus", swisseph.HousePlacidus, "Placidus", false},
		{"koch", swisseph.HouseKoch, "Koch", false},
		{"whole-sign", swisseph.HouseWholeSign, "Whole Sign", false},
		{"regiomontanus", swisseph.HouseRegiomontanus, "Regiomontanus", false},
		{"equal", swisseph.HouseEqual, "Equal", false},
		{"campanus", swisseph.HouseCampanus, "Campanus", false},
		// Case-insensitive (function lowercases input)
		{"Placidus", swisseph.HousePlacidus, "Placidus", false},
		{"PLACIDUS", swisseph.HousePlacidus, "Placidus", false},
		{"Koch", swisseph.HouseKoch, "Koch", false},
		{"Whole-Sign", swisseph.HouseWholeSign, "Whole Sign", false},
		// Invalid inputs
		{"", 0, "", true},
		{"unknown", 0, "", true},
		{"porphyry", 0, "", true},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got, display, err := parseHouseSystem(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if got != tc.want {
					t.Errorf("code = %v, want %v", got, tc.want)
				}
				if display != tc.wantDisplay {
					t.Errorf("display = %q, want %q", display, tc.wantDisplay)
				}
			}
		})
	}
}

func TestParseAlmutenMode(t *testing.T) {
	cases := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"", "", false},
		{"geniture", "geniture", false},
		{"figuris", "figuris", false},
		{"bonatti", "bonatti", false},
		{"both", "both", false},
		{"all", "all", false},
		{"Geniture", "geniture", false}, // case-insensitive
		{"BOTH", "both", false},
		{"lord", "", true},
		{"unknown", "", true},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got, err := parseAlmutenMode(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("mode = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseSearchTarget(t *testing.T) {
	cases := []struct {
		input      string
		wantPlanet int
		wantLon    float64
		wantErr    bool
	}{
		{"", noSearch, 0, false},
		{"mars:195", swisseph.Mars, 195, false},
		{"mars:15aries", swisseph.Mars, 15, false},
		{"Venus:0cancer", swisseph.Venus, 90, false},
		{"MARS:195", swisseph.Mars, 195, false}, // case-insensitive planet
		{"saturn:29.5pisces", swisseph.Saturn, 359.5, false},
		{"pluto:10leo", noSearch, 0, true},     // unsupported planet
		{"meannode:10leo", noSearch, 0, true},  // explicitly out of scope
		{"mars", noSearch, 0, true},            // missing colon
		{"mars:", noSearch, 0, true},           // empty longitude
		{":195", noSearch, 0, true},            // empty planet
		{"mars:400", noSearch, 0, true},        // raw degrees out of [0,360)
		{"mars:-10", noSearch, 0, true},        // negative raw degrees
		{"mars:15notasign", noSearch, 0, true}, // unparseable longitude
		{"mars:30aries", noSearch, 0, true},    // sign degree out of [0,30)
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			gotPlanet, gotLon, err := parseSearchTarget(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if gotPlanet != tc.wantPlanet {
				t.Errorf("planet = %v, want %v", gotPlanet, tc.wantPlanet)
			}
			if gotLon != tc.wantLon {
				t.Errorf("longitude = %v, want %v", gotLon, tc.wantLon)
			}
		})
	}
}
