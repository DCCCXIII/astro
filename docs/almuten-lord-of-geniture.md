# Lord of the Geniture (William Lilly)

**Source:** William Lilly, *Christian Astrology* (1647), p.115.
**Code:** `almuten/geniture.go`. **Call:** `LordOfGeniture(c, opts)`.

Lilly's Lord of the Geniture asks a different question than Ibn Ezra's or
Bonatti's almuten: instead of asking "which planet rules a specific *place* in
the chart," it scores each of the seven traditional planets directly, on its
own merits — essential dignity plus a long list of accidental and incidental
conditions — and the highest total wins. It's the most expansive of the three
algorithms in this codebase: motion, solar relationship, partile aspect,
besiegement, and fixed stars all contribute, none of which the other two
methods model at all (see the comparison table in
[docs/README.md](README.md)).

## Essential dignities at the planet's own position

The five dignities from [dignities.md](dignities.md) are scored at each
planet's own longitude (`almuten/geniture.go:53-94`), plus the two debilities
unique to this algorithm:

| Condition | Points |
|---|---|
| Domicile (incl. mutual reception) | +5 |
| Exaltation (incl. mutual reception) | +4 |
| Triplicity (active sect) | +3 |
| Term | +2 |
| Face | +1 |
| Detriment | −5 |
| Fall | −4 |
| Peregrine | −5 |

**Peregrine** means lacking essential dignity — but "lacking" has two
historical readings, and this codebase exposes the choice as
`Options.PeregrineSuppressesTermFace` (`almuten/chart.go:53`):

- Strict reading (`false`): peregrine only if the planet holds *none* of the
  five dignities; if it holds a term or face, those points still count even
  without domicile/exaltation/triplicity.
- Lenient-toward-debility reading (`true`, the default): a planet lacking
  all three *major* dignities is peregrine regardless of term/face, and the
  term/face points are suppressed entirely.

See `almuten/geniture.go:75-94` for the exact logic, and
[implementation-notes.md](implementation-notes.md) for why the default
favors the stricter reading.

## Accidental house placement

Lilly's own valuation table, attributed directly to *Christian Astrology*
p.115 (`lillyHousePoints`, `almuten/geniture.go:10-20`), using the plain
containing house (no orb):

| House | 1 | 10 | 4, 7, 11 | 2, 5 | 9 | 3 | 6, 8 | 12 |
|---|---|---|---|---|---|---|---|---|
| Points | +5 | +5 | +4 | +3 | +2 | +1 | −2 | −5 |

The angles (1st, 10th) score highest; the 12th — traditionally the house of
self-undoing — is the only double-digit penalty.

## Motion

Applies to the five non-luminaries only — the Sun and Moon never go
retrograde, so they're exempt (`almuten/geniture.go:97`):

| Condition | Points |
|---|---|
| Direct | +4 |
| Retrograde | −5 |
| Swift (faster than `dignities.MeanDailyMotion`) | +2 |
| Slow | −2 |

## Solar relationship (non-Sun planets)

Banded by elongation from the Sun (`almuten/geniture.go:111-122`):

| Condition | Orb | Points |
|---|---|---|
| Cazimi | ≤ 0.283° (17′) | +5 |
| Combust | ≤ 8.5° | −5 |
| Under the Sun's beams | ≤ 17° | −4 |
| Free of the Sun | > 17° | `Options.FreeOfSunBonus` (default +5) |

Of these four bands, only "free of the Sun" is configurable — the other
three are fixed point values. See
[implementation-notes.md](implementation-notes.md).

## Oriental / occidental

A planet is **oriental** when it rises before the Sun — i.e. its longitude is
behind the Sun's (`almuten/geniture.go:124-143`). The bonus polarity flips
between the outer ("superior") and inner ("inferior") planets, per
traditional rulership doctrine:

| Planets | Oriental | Occidental |
|---|---|---|
| Saturn, Jupiter, Mars | +2 | −2 |
| Mercury, Venus | −2 | +2 |

## Moon waxing / waning

The Moon alone gets ±2 depending on whether it's ahead of or behind the Sun
by less than 180° (`almuten/geniture.go:146-152`): waxing (ahead, +2) or
waning (behind, −2).

## Partile aspects

Aspects to the benefics (Jupiter, Venus) and malefics (Saturn, Mars), and to
the lunar nodes, count only when **partile** — within a 1° orb of exact —
using the five Ptolemaic aspects (0/60/90/120/180°)
(`almuten/geniture.go:157-191`):

| Aspect to | Conjunction | Trine | Sextile | Square | Opposition |
|---|---|---|---|---|---|
| Benefic | +5 | +4 | +3 | — | — |
| Malefic | −5 | — | — | −4 | −4 |
| North Node | +4 | — | — | — | — |
| South Node | −4 | — | — | — | — |

This is a 1° partile-only model — no wider applying/separating aspect logic
is implemented.

## Besiegement by Saturn and Mars

A planet is besieged when it sits on the **minor arc (≤ 30°) between Saturn
and Mars**, with no other classical planet between them
(`besiegedBySaturnMars`, `almuten/geniture.go:254-285`): −4. The code's own
comment is explicit that this is "a pragmatic body-besiegement test; the
fuller 'by ray' form is not modelled" (`almuten/geniture.go:252-253`) — see
[implementation-notes.md](implementation-notes.md).

## Fixed stars

Three stars only — Regulus, Spica, and Algol (`fixedStars`,
`almuten/geniture.go:23-30`) — the most commonly cited fixed stars in natal
work, drawn from the Behenian/Royal-star tradition, not a full catalogue.
Conjunction orb defaults to 5° (`Options.FixedStarOrb`):

| Star | Points |
|---|---|
| Regulus | +6 |
| Spica | +5 |
| Algol | −4 |

If the star catalogue (`ephe/sefstars.txt`) can't resolve a name, the bonus
is silently skipped rather than failing the whole calculation
(`almuten/geniture.go:200-202`).

## Mutual reception

Two planets in mutual reception each occupy a sign the other rules or is
exalted in — checked separately for domicile and exaltation
(`mutualReceptionDomicile`, `mutualReceptionExalt`,
`almuten/geniture.go:226-248`). Reception substitutes for direct dignity: a
planet in reception scores the domicile/exaltation points even without
holding the dignity outright (`almuten/geniture.go:55-56`).

Next: [almuten-figuris.md](almuten-figuris.md) for Ibn Ezra's very different
five-place approach, or [implementation-notes.md](implementation-notes.md)
for a consolidated list of where this file's rules are this codebase's
specific choice rather than settled tradition.
