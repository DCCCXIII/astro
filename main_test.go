package main

import (
	"testing"

	"github.com/dcccxiii/astro/swisseph"
)

func TestParseHouseSystem(t *testing.T) {
	cases := []struct {
		input   string
		want    byte
		wantErr bool
	}{
		// All valid names, canonical casing
		{"placidus", swisseph.HousePlacidus, false},
		{"koch", swisseph.HouseKoch, false},
		{"whole-sign", swisseph.HouseWholeSign, false},
		{"regiomontanus", swisseph.HouseRegiomontanus, false},
		{"equal", swisseph.HouseEqual, false},
		{"campanus", swisseph.HouseCampanus, false},
		// Case-insensitive (function lowercases input)
		{"Placidus", swisseph.HousePlacidus, false},
		{"PLACIDUS", swisseph.HousePlacidus, false},
		{"Koch", swisseph.HouseKoch, false},
		{"Whole-Sign", swisseph.HouseWholeSign, false},
		// Invalid inputs
		{"", 0, true},
		{"unknown", 0, true},
		{"porphyry", 0, true},
	}

	for _, tc := range cases {
		got, err := parseHouseSystem(tc.input)
		if tc.wantErr {
			if err == nil {
				t.Errorf("parseHouseSystem(%q): expected error, got nil", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("parseHouseSystem(%q): unexpected error: %v", tc.input, err)
			}
			if got != tc.want {
				t.Errorf("parseHouseSystem(%q) = %v, want %v", tc.input, got, tc.want)
			}
		}
	}
}
