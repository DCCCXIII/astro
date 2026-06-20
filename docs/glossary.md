# Glossary

Named conditions referenced across the almuten files, alphabetically. Nearly
every entry below is specific to Lilly's Lord of the Geniture — Lilly's
algorithm is the only one of the three that models these accidental/
incidental conditions; Ibn Ezra's Almuten Figuris and Bonatti's Almudebit
score essential dignity (plus, for Figuris, accidental house/temporal/
synodic factors) but none of the conditions below. If you're looking for
cazimi or besiegement logic in Figuris or Bonatti, it isn't there — see the
comparison table in [almuten-bonatti.md](almuten-bonatti.md#comparing-all-three-algorithms).

**Besiegement** — a planet trapped on the minor arc (≤ 30°) between Saturn
and Mars, with no other classical planet between them. A pragmatic
simplification of the traditional "besieged by ray" condition (the fuller
ray-based form isn't modeled). `besiegedBySaturnMars`,
`almuten/geniture.go:254-285`. −4. Geniture only.

**Cazimi** — exact conjunction with the Sun, within 17′ (0.283°) of orb. The
tightest and most favorable solar condition. `almuten/geniture.go:113`. +5.
Geniture only.

**Combust** — within 8.5° of the Sun (beyond cazimi range). The planet is
considered overpowered by solar light. `almuten/geniture.go:115`. −5.
Geniture only.

**Detriment** — a planet positioned in the sign opposite the one it rules.
See [dignities.md](dignities.md#detriment-5-lilly-only). `−5`. Geniture
only.

**Fall** — a planet positioned in the sign opposite its exaltation. See
[dignities.md](dignities.md#exaltation-and-fall-4---4). −4. Geniture only.

**Free of the Sun** — beyond 17° of solar elongation; the opposite of
combust/under beams. The only one of the four solar-relationship bands that
is configurable (`Options.FreeOfSunBonus`, default +5).
`almuten/geniture.go:119-121`. Geniture only.

**Mutual reception** — two planets each occupying a sign the other rules
(domicile reception) or is exalted in (exaltation reception). Substitutes
for direct dignity in scoring. `mutualReceptionDomicile`,
`mutualReceptionExalt`, `almuten/geniture.go:226-248`. Geniture only.

**Oriental / Occidental** — a planet is oriental when it rises before the
Sun (lower ecliptic longitude); occidental otherwise. The bonus polarity
flips between superior planets (Saturn/Jupiter/Mars favor oriental) and
inferior/personal planets (Mercury/Venus favor occidental).
`almuten/geniture.go:124-143`. ±2. Geniture only.

**Partile aspect** — an aspect (conjunction, sextile, square, trine,
opposition) within 1° of exact. This codebase only scores partile aspects —
no wider applying/separating model. `partileAspect`,
`almuten/geniture.go:214-222`. Geniture only.

**Peregrine** — a planet lacking essential dignity at its own position. Two
historical readings are exposed as `Options.PeregrineSuppressesTermFace` —
see [almuten-lord-of-geniture.md](almuten-lord-of-geniture.md#essential-dignities-at-the-planets-own-position).
`almuten/geniture.go:75-94`. −5. Geniture only.

**Sect** — whether a chart is diurnal (Sun above horizon) or nocturnal. The
one concept in this glossary shared by all three algorithms — see
[vital-points-and-sect.md](vital-points-and-sect.md). `almuten/chart.go:93-95`.

**Under the Sun's Beams** — within 17° of the Sun but beyond the 8.5°
combust threshold; a milder solar affliction than combustion.
`almuten/geniture.go:117`. −4. Geniture only.

**Fixed stars (Regulus, Spica, Algol)** — conjunction (default 5° orb) with
one of three traditionally significant fixed stars. See
[almuten-lord-of-geniture.md](almuten-lord-of-geniture.md#fixed-stars).
`almuten/geniture.go:22-30,198-207`. Regulus +6, Spica +5, Algol −4.
Configurable orb (`Options.FixedStarOrb`). Geniture only.

Back to [docs/README.md](README.md) for the full reading order, or
[implementation-notes.md](implementation-notes.md) for which of these are
this codebase's specific choice rather than universal tradition.
