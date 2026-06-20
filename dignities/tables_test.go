package dignities

import "testing"

// TestTermsCoverSignContiguously asserts each sign's five Egyptian terms run in
// ascending order and collectively cover 0–30° with no gaps or overlaps.
func TestTermsCoverSignContiguously(t *testing.T) {
	for sign := 0; sign < 12; sign++ {
		prev := 0.0
		for i, term := range terms[sign] {
			if term.upTo <= prev {
				t.Errorf("sign %d term %d: upper bound %.0f not greater than previous %.0f", sign, i, term.upTo, prev)
			}
			prev = term.upTo
		}
		if prev != 30 {
			t.Errorf("sign %d terms end at %.0f, want 30", sign, prev)
		}
	}
}

// TestFacesCoverSign asserts each sign has exactly three faces (covering the
// three decans 0–10, 10–20, 20–30).
func TestFacesCoverSign(t *testing.T) {
	for sign := 0; sign < 12; sign++ {
		if len(faces[sign]) != 3 {
			t.Errorf("sign %d has %d faces, want 3", sign, len(faces[sign]))
		}
	}
}

// TestDomicileDetrimentOpposite asserts the detriment ruler of a sign is the
// domicile ruler of the opposite sign.
func TestDomicileDetrimentOpposite(t *testing.T) {
	for sign := 0; sign < 12; sign++ {
		opposite := (sign + 6) % 12
		if detriment[sign] != domicile[opposite] {
			t.Errorf("sign %d: detriment %s != domicile of opposite sign %s",
				sign, detriment[sign].Name(), domicile[opposite].Name())
		}
	}
}

// TestExaltFallOpposite asserts the fall ruler of a sign is the planet exalted
// in the opposite sign.
func TestExaltFallOpposite(t *testing.T) {
	for sign := 0; sign < 12; sign++ {
		exaltOpp, _, exaltOK := ExaltationRuler((sign + 6) % 12)
		fall, fallOK := FallRuler(sign)
		if fallOK != exaltOK {
			t.Errorf("sign %d: FallRuler ok=%v, ExaltationRuler(opposite) ok=%v", sign, fallOK, exaltOK)
		}
		if fallOK && fall != exaltOpp {
			t.Errorf("sign %d: fall %s != exalt-of-opposite %s", sign, fall.Name(), exaltOpp.Name())
		}
	}
}

func TestTermRulerKnownDegrees(t *testing.T) {
	cases := []struct {
		sign int
		deg  float64
		want Planet
	}{
		{0, 5, Jupiter},  // Aries 0–6 Jupiter
		{0, 6, Venus},    // boundary: 6 belongs to next term
		{0, 24.9, Mars},  // Aries 20–25 Mars
		{0, 25, Saturn},  // Aries 25–30 Saturn
		{0, 30, Saturn},  // exact 30 -> last term
		{11, 0, Venus},   // Pisces 0–12 Venus
		{11, 28, Saturn}, // Pisces 28–30 Saturn
	}
	for _, c := range cases {
		if got := TermRuler(c.sign, c.deg); got != c.want {
			t.Errorf("TermRuler(%d, %.1f) = %s, want %s", c.sign, c.deg, got.Name(), c.want.Name())
		}
	}
}

func TestFaceRulerKnownDegrees(t *testing.T) {
	cases := []struct {
		sign int
		deg  float64
		want Planet
	}{
		{0, 0, Mars},   // Aries I
		{0, 9.9, Mars}, // still Aries I
		{0, 10, Sun},   // Aries II
		{0, 20, Venus}, // Aries III
		{0, 29.9, Venus},
	}
	for _, c := range cases {
		if got := FaceRuler(c.sign, c.deg); got != c.want {
			t.Errorf("FaceRuler(%d, %.1f) = %s, want %s", c.sign, c.deg, got.Name(), c.want.Name())
		}
	}
}

func TestActiveTriplicityRuler(t *testing.T) {
	// Fire diurnal = Sun, nocturnal = Jupiter.
	if got := ActiveTriplicityRuler(0, true, false); got != Sun {
		t.Errorf("fire day lord = %s, want Sun", got.Name())
	}
	if got := ActiveTriplicityRuler(0, false, false); got != Jupiter {
		t.Errorf("fire night lord = %s, want Jupiter", got.Name())
	}
	// Water with Lilly override is Mars day and night.
	if got := ActiveTriplicityRuler(3, true, true); got != Mars {
		t.Errorf("lilly water day lord = %s, want Mars", got.Name())
	}
	if got := ActiveTriplicityRuler(3, false, true); got != Mars {
		t.Errorf("lilly water night lord = %s, want Mars", got.Name())
	}
}
