# House Systems and Angles

A house system divides the ecliptic into twelve houses relative to a time and
place. Astrologers have never settled on one method — the choice affects
which house a planet falls into, which in turn changes Lilly's and Ibn
Ezra's accidental-house scores (see the callout at the end of this file). This
codebase supports six systems via `swisseph.CalcHouses`
(`swisseph/swisseph.go:231-...`), selected with the `--house-system` flag
(`cmd/run.go:118-135`).

## The six supported systems

| Flag value | Constant | Code | Era / character |
|---|---|---|---|
| `placidus` (default) | `HousePlacidus` | `'P'` | 17th c. time-based quadrant system; became the de facto modern Western default |
| `koch` | `HouseKoch` | `'K'` | 1960s, Walter Koch — the "birthplace system," a Placidus variant adjusted for high-latitude reliability |
| `whole-sign` | `HouseWholeSign` | `'W'` | Hellenistic-era — the oldest attested system; one house per sign starting at the Ascendant's sign |
| `regiomontanus` | `HouseRegiomontanus` | `'R'` | Described 1467 by Johannes Müller (Regiomontanus); dominant from the 15th c. until Placidus overtook it around 1900 |
| `equal` | `HouseEqual` | `'A'` | Simplest mathematically: 30° houses measured from the Ascendant |
| `campanus` | `HouseCampanus` | `'C'` | 13th c., Giovanni Campano — a prime-vertical-based quadrant system |

Constants: `swisseph/swisseph.go:40-47`. Flag parsing:
`parseHouseSystem` (`cmd/run.go:118-135`).

A few notes worth being precise about:

- **Whole Sign predates Placidus by well over a millennium** — it's the
  house system used in surviving Hellenistic horoscopes from the 1st century
  CE onward. Don't read "Placidus is the default" as "Placidus is the
  historically original" system; it's simply the most common one in modern
  Western practice.
- **Regiomontanus is a refinement of Campanus**, not an independent
  invention — both divide a great circle (the celestial equator for
  Regiomontanus, the prime vertical for Campanus) into 12 equal arcs and
  project the divisions onto the ecliptic. Interestingly, a version of the
  same equal-celestial-equator-division idea is also attributed to Ibn Ezra,
  predating Regiomontanus by roughly three centuries — the naming reflects
  who popularized the method in Latin Europe, not strict priority.
- **Koch is explicitly a Placidus variant**, not an unrelated system: it
  keeps Placidus's time-based logic for houses 10/11/12/1/2/3 but recomputes
  11 and 12 differently to behave better at high latitudes where Placidus
  can fail outright.

## Angles

`HouseResult` (`swisseph/swisseph.go:219-225`) returns four angles alongside
the twelve cusps:

| Angle | Meaning |
|---|---|
| Ascendant | the ecliptic point rising on the eastern horizon at the moment in question |
| MC (Midheaven) | the ecliptic point culminating on the meridian |
| ARMC | sidereal time expressed in degrees — the astronomical input house math is built from, not itself a point astrologers interpret the way they interpret the Ascendant or MC |
| Vertex | the ecliptic point on the western horizon, of more limited and varied use across traditions — included here for completeness rather than because every algorithm in this codebase consumes it |

Note that the Descendant and IC (Imum Coeli) are *not* separate fields on
`HouseResult` — they're simply the oppositions of the Ascendant and MC.
Bonatti's algorithm computes them this way explicitly
(`almuten/bonatti.go:42-43`); see
[almuten-bonatti.md](almuten-bonatti.md).

## House-system choice affects almuten scoring

Both Lilly's Lord of the Geniture and Ibn Ezra's Almuten Figuris assign
points based on which house a planet occupies — Lilly via the plain
containing house (`HouseOf`, `almuten/chart.go:161-176`), Figuris via a
5°-pre-cusp-adjusted house (`adjustedHouse`, `almuten/figuris.go:97-104`).
Changing `--house-system` changes the cusp positions, which can shift a
planet from one house to an adjacent one — and therefore change its
accidental-house score and, in close cases, the algorithm's winner. This is
an easy lever to overlook: the same birth data run with `--house-system
whole-sign` versus the default Placidus can produce a different Lord of the
Geniture or Almuten Figuris victor purely from the house reassignment, with
no change to any planet's actual position.

Bonatti's algorithm is unaffected by house-system choice except through the
Ascendant/MC angles themselves (which all six systems compute identically,
since they depend only on time and location, not the house-division method)
— see [almuten-bonatti.md](almuten-bonatti.md).

Next: [glossary.md](glossary.md) for named conditions referenced across the
three almuten files, or [implementation-notes.md](implementation-notes.md)
for a consolidated view of every place this codebase makes a specific
implementation choice.
