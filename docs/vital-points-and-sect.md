# Sect and the Vital Points

Two mechanics are computed once in `almuten/chart.go` and `almuten/syzygy.go`
and then reused across all three almuten algorithms (with one exception
noted below): **sect**, which determines which triplicity lord is "active,"
and the **vital points** — Part of Fortune and prenatal syzygy — which stand
in for the Ascendant and luminaries as additional places to score essential
dignity.

## Sect

A chart is diurnal when the Sun is above the horizon (houses 7–12) at the
moment in question, nocturnal otherwise:

```go
c.IsDiurnal = houseOf(c.Positions[swisseph.Sun].Longitude, houses.Cusps) >= 7
```

(`almuten/chart.go:94`)

This single boolean gates `dignities.ActiveTriplicityRuler(sign, isDiurnal,
lillyWater)` everywhere it's called — Geniture (`almuten/geniture.go:54`),
Figuris (`almuten/figuris.go:90`), and both rounds of Bonatti
(`almuten/bonatti.go:48`). It is the single most load-bearing shared concept
in this codebase: get sect wrong and the triplicity dignity (+3) is wrong for
every planet in every algorithm. See [dignities.md](dignities.md#triplicity-3)
for the day/night lord table itself.

## Part of Fortune

The Part (or Lot) of Fortune is a chart-derived point, not a planet's actual
position — classically read as the place of material fortune and bodily
well-being. Its formula reverses by sect, in the Hellenistic tradition:

```go
if c.IsDiurnal {
    c.PartFortune = norm360(asc + moon - sun)
} else {
    c.PartFortune = norm360(asc + sun - moon)
}
```

(`almuten/chart.go:106-110`)

By day: Ascendant + Moon − Sun. By night: Ascendant + Sun − Moon. The
reversal exists so that the Part keeps the same *relationship* to the
Ascendant as the Moon has to the Sun, regardless of which luminary is
"in charge" for the sect of the chart.

## Prenatal Syzygy

The prenatal syzygy is the most recent new or full moon (lunation) before the
moment being charted — Hellenistic astrologers called it simply "syzygy" and
treated it as a "place of life" alongside the Ascendant and luminaries.
Ptolemy's *Tetrabiblos* groups exactly these same five points — Sun, Moon,
Ascendant, Part of Fortune, and prenatal syzygy — as the candidate places for
the *hyleg* (the chart's life-indicating point); Ibn Ezra's Almuten Figuris
(see [almuten-figuris.md](almuten-figuris.md)) reuses the identical five
places to find a *ruler* instead, applying full essential-dignity scoring at
each one rather than just checking which planet governs it.

`PrenatalSyzygy(jd, lat, lon)` (`almuten/syzygy.go:30-67`) walks backward from
the birth instant in half-day steps looking for a sign change in the
Moon-Sun elongation:

- A **new moon** is a sign change crossing 0° — the longitude returned is the
  exact conjunction degree (the Sun's longitude at that instant).
- A **full moon** is a wrap crossing ±180°.

Once a bracket containing the crossing is found, `refineCrossing`
(`almuten/syzygy.go:72-104`) bisects it down to about one second of time —
far tighter precision than the coarse half-day search step, since a wrong
sign assignment for the syzygy point would propagate into every algorithm
that scores it.

For a full moon, the longitude used isn't simply "the Moon's position" — it's
whichever luminary was **above the horizon at the syzygy instant, evaluated
at the birth location** (not the birth time):

```go
// The luminary in houses 7–12 is above the horizon.
if houseOf(sun.Longitude, houses.Cusps) >= 7 {
    return norm360(sun.Longitude), "full", nil
}
return norm360(moon.Longitude), "full", nil
```

(`almuten/syzygy.go:121-125`)

This mirrors Ptolemy's own rule for which luminary to read at a full-moon
syzygy, and is easy to misread as "just take the Moon's longitude" if you
don't look at the code closely.

## Where these are consumed

`Chart.computeVitalPoints()` (`almuten/chart.go:101-118`) is the single
shared entry point that fills `PartFortune`, `Syzygy`, and `SyzygyType` on a
`Chart`. It's called by Figuris (via `prepareFiguris`,
`almuten/figuris.go:110-114`) and directly by Bonatti
(`almuten/bonatti.go:19`) — **not** by Geniture, which doesn't use Part of
Fortune or prenatal syzygy at all; its only vital-point-adjacent factor is the
Sun itself, used for cazimi/combustion (see
[almuten-lord-of-geniture.md](almuten-lord-of-geniture.md)).

| Vital point | Geniture | Figuris | Bonatti |
|---|---|---|---|
| Ascendant | — | scored (1 of 5 places) | scored as an angle (4 of 4) |
| Sun | solar-relationship only | scored (1 of 5 places) | scored (1 of 4 vital points) |
| Moon | waxing/waning only | scored (1 of 5 places) | scored (1 of 4 vital points) |
| Part of Fortune | not used | scored (1 of 5 places) | scored (1 of 4 vital points) |
| Prenatal syzygy | not used | scored (1 of 5 places) | scored (1 of 4 vital points) |

Next: pick an algorithm —
[almuten-lord-of-geniture.md](almuten-lord-of-geniture.md),
[almuten-figuris.md](almuten-figuris.md), or
[almuten-bonatti.md](almuten-bonatti.md).
