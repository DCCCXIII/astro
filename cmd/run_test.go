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
