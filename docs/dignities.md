# Essential Dignities

The `dignities` package (pure Go, no cgo) holds the static tables that every
almuten algorithm in this codebase spends as its scoring currency. Before
reading any of the three almuten files, it's worth understanding these five
tables and what each one claims about a planet's "strength" in a sign.

Essential dignity is positional: it depends only on which sign (and degree)
a point falls in, never on aspects, houses, or motion. The same five-dignity
scheme — domicile, exaltation, triplicity, term, face, each worth a
descending number of points — is shared by all three almuten methods via
`dignities.EssentialDignities` (`dignities/lookup.go:85-94`); only Lilly's
Lord of the Geniture also subtracts for the corresponding debilities
(detriment, fall).

## Domicile (+5)

Each sign is "ruled" by one planet — the planet most at home there. This is
the oldest and least contested dignity: `domicile[sign]`
(`dignities/tables.go:36-49`), looked up via `DomicileRuler(sign)`
(`dignities/lookup.go:7`).

| Sign | Ruler | Sign | Ruler |
|---|---|---|---|
| Aries | Mars | Libra | Venus |
| Taurus | Venus | Scorpio | Mars |
| Gemini | Mercury | Sagittarius | Jupiter |
| Cancer | Moon | Capricorn | Saturn |
| Leo | Sun | Aquarius | Saturn |
| Virgo | Mercury | Pisces | Jupiter |

This is the classical seven-planet rulership scheme, predating the discovery
of Uranus, Neptune, and Pluto — modern domicile schemes that assign those
three planets a sign are a later addition this codebase does not implement.

## Exaltation and Fall (+4 / −4)

A planet is exalted in one sign at one specific degree — a position of
honor short of rulership. Four signs (Gemini, Leo, Scorpio, Sagittarius) have
no traditional exaltation at all. `exaltations[sign]`
(`dignities/tables.go:68-90`) stores both the planet and its exact degree
(used only for niceties beyond the +4 itself); `ExaltationRuler(sign)`
(`dignities/lookup.go:14-20`) returns `ok=false` for the four signless signs.

| Planet | Exalted in | Degree |
|---|---|---|
| Sun | Aries | 19° |
| Moon | Taurus | 3° |
| Jupiter | Cancer | 15° |
| Mercury | Virgo | 15° |
| Saturn | Libra | 21° |
| Mars | Capricorn | 28° |
| Venus | Pisces | 27° |

Fall is not a separately tabulated dignity — it's derived as "exalted in the
opposite sign," via `FallRuler(sign)` (`dignities/lookup.go:24-31`), which
looks up `exaltations[(sign+6)%12]`.

## Detriment (−5, Lilly only)

Detriment is the mirror of domicile: a planet sits in detriment in the sign
opposite its own rulership. `detriment[sign]` (`dignities/tables.go:53-66`),
via `DetrimentRuler(sign)` (`dignities/lookup.go:9-10`).

> **Implementation note:** detriment is scored only by `LordOfGeniture`
> (`almuten/geniture.go:68-70`) — neither Figuris nor Bonatti calls
> `DetrimentRuler` at all, because `addEssentialDignities`
> (`almuten/figuris.go:80-93`, shared by both) only ever adds points and never
> subtracts. Detriment/fall debilities are a Lilly-specific feature of this
> codebase's scoring, not a universal dignity-table component. See
> [implementation-notes.md](implementation-notes.md).

## Triplicity (+3)

Each element (Fire, Earth, Air, Water) has three rulers in the Dorothean
scheme: a day lord, a night lord, and a participating lord. `triplicities`
(`dignities/tables.go:92-103`) stores all three; which one is "active" for
scoring depends on [sect](vital-points-and-sect.md) — `ActiveTriplicityRuler(sign,
isDiurnal, lillyWater)` (`dignities/lookup.go:36-45`) returns the day lord by
day, the night lord by night.

| Element | Signs | Day lord | Night lord | Participating |
|---|---|---|---|---|
| Fire | Aries, Leo, Sagittarius | Sun | Jupiter | Saturn |
| Earth | Taurus, Virgo, Capricorn | Venus | Moon | Mars |
| Air | Gemini, Libra, Aquarius | Saturn | Mercury | Jupiter |
| Water | Cancer, Scorpio, Pisces | Venus | Mars | Moon |

> **Implementation note:** Lilly's water-triplicity override
> (`lillyWaterTriplicity`, `dignities/tables.go:105-107`) replaces the day/night
> water lords with Mars/Mars, diverging from the Dorothean Venus(day)/Mars(night)
> shown above. It's gated by the `lillyWater` parameter and applied by
> `LordOfGeniture` via `Options.LillyWater` (default `true`); Figuris and
> Bonatti both call `ActiveTriplicityRuler` with `lillyWater=false`
> (`almuten/figuris.go:90`, `almuten/bonatti.go:48`), so they always use the
> plain Dorothean table even for water signs. See
> [implementation-notes.md](implementation-notes.md).

The participating lord (the third name in each row) is excluded from Figuris
entirely but chased by Bonatti — see
[almuten-bonatti.md](almuten-bonatti.md) — via
`ParticipatingTriplicityRuler(sign)` (`dignities/lookup.go:47-51`).

This table follows the system used by Dorotheus of Sidon (*Carmen
Astrologicum*, 1st c. CE) and Vettius Valens; it is not the only historical
triplicity scheme — Ptolemy's *Tetrabiblos* assigns a single ruler per
element rather than day/night/participating lords, and this codebase does
not implement the Ptolemaic variant.

## Terms / Bounds (+2)

Each sign is divided into five unequal segments (bounds), each ruled by a
different planet, together covering the full 0–30°. `terms[sign]`
(`dignities/tables.go:118-131`) stores each bound's upper edge and ruler, in
ascending order; `TermRuler(sign, degInSign)` (`dignities/lookup.go:55-62`)
does a linear scan to find the bound containing a given degree.

This codebase implements the **Egyptian** bound system — the one used by
Dorotheus, Vettius Valens, and (with minor adjustments) Ptolemy's
*Tetrabiblos* I.20. Valens attributes the Egyptian bounds to "the Egyptian
sages," most likely referring to the now-lost work ascribed to
Nechepso/Petosiris. Ptolemy revised a handful of degree boundaries in his own
table for reasons that aren't entirely clear (possibly astronomical
recalculation); this codebase implements the Egyptian table, not Ptolemy's
revision, and has no toggle between them.

## Faces / Decans (+1)

Each sign is divided into three equal 10° decans, each ruled by a planet in
the **Chaldean order** — the planets sequenced by apparent geocentric speed,
slowest to fastest: Saturn → Jupiter → Mars → Sun → Venus → Mercury → Moon
(`dignities.ChaldeanOrder`, `dignities/tables.go:150-152`). `faces[sign]`
(`dignities/tables.go:135-148`) stores the three rulers per sign;
`FaceRuler(sign, degInSign)` (`dignities/lookup.go:64-71`) divides the degree
by 10 to pick one.

The Chaldean order itself predates its use for decans: it was taught by the
Greco-Egyptian astrologer Teucer of Babylon (3rd c. CE), Firmicus Maternus
(4th c.), and Abu Ma'shar (9th c.), and the same sequence drives planetary
hours and weekday rulers in [almuten-figuris.md](almuten-figuris.md)
(`dignities.WeekdayRulers`, `dignities/tables.go:156-164`) — one historical
ordering reused for two unrelated purposes in this codebase.

## Quick reference

| Dignity | Points | Table | Lookup | Used by |
|---|---|---|---|---|
| Domicile | +5 | `tables.go:36-49` | `DomicileRuler` | all three |
| Exaltation | +4 | `tables.go:68-90` | `ExaltationRuler` | all three |
| Triplicity | +3 | `tables.go:92-107` | `ActiveTriplicityRuler` | all three |
| Term | +2 | `tables.go:109-131` | `TermRuler` | all three |
| Face | +1 | `tables.go:133-152` | `FaceRuler` | all three |
| Detriment | −5 | `tables.go:51-66` | `DetrimentRuler` | Geniture only |
| Fall | −4 | (derived) | `FallRuler` | Geniture only |

Next: [vital-points-and-sect.md](vital-points-and-sect.md) — the chart-derived
points (Part of Fortune, prenatal syzygy) that Figuris and Bonatti score
against these same tables, and the sect rule that selects each sign's active
triplicity lord.
