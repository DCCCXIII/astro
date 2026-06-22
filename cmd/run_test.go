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
		wantDir    string
		wantErr    bool
	}{
		{"", noSearch, 0, "", false},
		{"mars:195", swisseph.Mars, 195, "backward", false},
		{"mars:15aries", swisseph.Mars, 15, "backward", false},
		{"Venus:0cancer", swisseph.Venus, 90, "backward", false},
		{"MARS:195", swisseph.Mars, 195, "backward", false}, // case-insensitive planet
		{"saturn:29.5pisces", swisseph.Saturn, 359.5, "backward", false},
		{"mars:15aries:forward", swisseph.Mars, 15, "forward", false},
		{"mars:15aries:backward", swisseph.Mars, 15, "backward", false},
		{"mars:15aries:FORWARD", swisseph.Mars, 15, "forward", false}, // case-insensitive direction
		{"mars:15aries:", swisseph.Mars, 15, "backward", false},       // trailing empty segment defaults
		{"pluto:10leo", noSearch, 0, "", true},                        // unsupported planet
		{"meannode:10leo", noSearch, 0, "", true},                     // explicitly out of scope
		{"mars", noSearch, 0, "", true},                               // missing colon
		{"mars:", noSearch, 0, "", true},                              // empty longitude
		{":195", noSearch, 0, "", true},                               // empty planet
		{"mars:400", noSearch, 0, "", true},                           // raw degrees out of [0,360)
		{"mars:-10", noSearch, 0, "", true},                           // negative raw degrees
		{"mars:15notasign", noSearch, 0, "", true},                    // unparseable longitude
		{"mars:30aries", noSearch, 0, "", true},                       // sign degree out of [0,30)
		{"mars:15aries:sideways", noSearch, 0, "", true},              // invalid direction
		{"mars:15aries:forward:extra", noSearch, 0, "", true},         // extra trailing segment rejected
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			gotPlanet, gotLon, gotDir, err := parseSearchTarget(tc.input)
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
			if gotDir != tc.wantDir {
				t.Errorf("direction = %q, want %q", gotDir, tc.wantDir)
			}
		})
	}
}

func TestParseSearchDirection(t *testing.T) {
	cases := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"", "backward", false},
		{"backward", "backward", false},
		{"forward", "forward", false},
		{"Forward", "forward", false},
		{"BACKWARD", "backward", false},
		{"sideways", "", true},
		{"forwards", "", true}, // close-but-wrong is still an error, no fuzzy match
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got, err := parseSearchDirection(tc.input)
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
				t.Errorf("direction = %q, want %q", got, tc.want)
			}
		})
	}
}
