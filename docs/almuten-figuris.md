# Almuten Figuris (Abraham Ibn Ezra)

**Source:** Abraham Ibn Ezra (12th c.), most directly associated with his
treatise *Reshit Hokhmah* ("The Beginning of Wisdom"), which codifies the
five hylegiacal places this algorithm scores.
**Code:** `almuten/figuris.go`. **Call:** `AlmutenFiguris(c)`.

Where Lilly scores planets directly, Ibn Ezra's almuten figuris asks "which
planet has the most dignity *across the chart's key points*" — a census
taken at five fixed places rather than at each planet's own position. Ibn
Ezra's own words (as preserved in the tradition): "there are five places of
life, namely, the two places of the luminaries... the third is the place of
the conjunction or opposition of the luminaries... the fourth is the
ascending degree, and the fifth is the lot of fortune." These are the exact
same five places Ptolemy's *Tetrabiblos* uses as hyleg candidates (see
[vital-points-and-sect.md](vital-points-and-sect.md)) — Figuris repurposes
that older five-place framework to find a *ruler* rather than a *hyleg*.

## The five vital places

`almuten/figuris.go:39-45`:

1. Ascendant
2. Sun
3. Moon
4. Part of Fortune
5. Prenatal syzygy

Each is scored for essential dignity via the shared `addEssentialDignities`
helper (`almuten/figuris.go:80-93`) — domicile +5, exaltation +4, triplicity
+3, term +2, face +1, added directly to the ruling planet's score for that
place. This same helper is also used by Bonatti
(`almuten/bonatti.go:45,51,67,70`) — it's the common dignity-scoring
primitive behind two of the three algorithms, not duplicated logic.

**Active-sect triplicity only:** Figuris calls
`dignities.ActiveTriplicityRuler(sign, isDiurnal, false)`
(`almuten/figuris.go:90`) — note the hardcoded `false`. Figuris never
chases the participating triplicity lord (unlike Bonatti, see
[almuten-bonatti.md](almuten-bonatti.md)) and never applies Lilly's
water-triplicity override even though the underlying table supports it.

## Accidental house placement

A second, independent house-valuation table from Lilly's
(`figurisHousePoints`, `almuten/figuris.go:13-16`):

| House | 1 | 10 | 7 | 4 | 11 | 5 | 2 | 9 | 8 | 3 | 12 | 6 |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| Points | 12 | 11 | 10 | 9 | 8 | 7 | 6 | 5 | 4 | 3 | 2 | 1 |

Unlike Lilly's table (which has no penalties below zero outside the 6th/8th/
12th houses), Figuris's table is a strict 12-down-to-1 ranking with no
negative values at all.

This algorithm also uses a different rule for *which* house a planet counts
as occupying: `adjustedHouse` (`almuten/figuris.go:97-104`) treats a planet
within **5° before** the next cusp (in zodiacal order) as already belonging
to that next house, rather than the plain containing-house test Lilly uses
(`HouseOf`, `almuten/chart.go:161-176`). The 5° pre-cusp orb is hardcoded —
see [implementation-notes.md](implementation-notes.md).

## Temporal rulers

Two more bonuses, computed by `computeTemporalRulers`
(`almuten/figuris.go:122-149`):

- **Weekday ruler** (+7): the planet ruling the day on which the astrological
  day began — the sunrise opening the relevant 24-hour window, not the civil
  midnight (`dignities.WeekdayRulers`).
- **Planetary-hour ruler** (+6): the Chaldean-order ruler of the unequal hour
  (1/12 of daylight or 1/12 of nighttime) containing the birth instant.

At polar latitudes where sunrise/sunset don't occur, `temporalFallback`
(`almuten/figuris.go:173-179`) substitutes an equal 12-hour day/night split
starting at 06:00 UT — a practical necessity with no historical equivalent,
since the unequal-hour system assumes a normal sunrise/sunset cycle.

## Synodic points (superior planets)

Saturn, Jupiter, and Mars each get a bonus while **separating** from the Sun
after conjunction — i.e. while the Sun has moved ahead of them — banded by
elongation (`almuten/figuris.go:60-74`):

| Elongation from Sun | Points |
|---|---|
| 15°–60° | +3 |
| 60°–90° | +2 |
| 90° up to first station | +1 |

The upper bound of the last band is each planet's approximate first-station
elongation — the point at which it turns retrograde —
(`firstStationElongation`, `almuten/figuris.go:20-24`): Saturn 109°, Jupiter
115°, Mars 136°. These are approximate bounding constants, not exact
ephemeris-derived values for the specific chart. No synodic bonus exists for
Mercury or Venus (inferior planets don't separate from the Sun the same way)
or for the Moon.

## What Figuris omits

Compared to Lilly's algorithm, Figuris has no besiegement, no fixed stars,
and no motion or partile-aspect penalties at all — its entire model is
dignity-at-a-place plus accidental house plus temporal plus synodic.

Next: [almuten-bonatti.md](almuten-bonatti.md) for the third algorithm's
two-round "chase" structure, or back to
[vital-points-and-sect.md](vital-points-and-sect.md) for how Part of Fortune
and prenatal syzygy (two of the five places here) are derived.
