# Bonatti's Almudebit (Guido Bonatti)

**Source:** Guido Bonatti, *Liber Astronomiae* ("Book of Astronomy" /
"Book of Astrology"), Latin, c. 1277. Bonatti was a 13th-century Italian
astrologer who compiled and synthesized earlier Arabic-transmitted material
into a vast Latin treatise (translated into English in full for the first
time by Benjamin Dykes). This codebase's name for the algorithm,
"Almudebit," is the form used in the secondary literature this
implementation draws on; whether it is Bonatti's own term or a later
translator's transliteration is not confirmed — treat it as a label for
"Bonatti's lord-finding procedure" rather than a precisely sourced term.

**Code:** `almuten/bonatti.go`. **Call:** `AlmutenBonatti(c)`.

## The core idea: a two-round "chase"

Unlike Lilly (score planets directly) or Ibn Ezra (score five fixed places),
Bonatti's method scores a point, then re-scores at the dispositor(s) of that
point — chasing dignity one step along a chain of rulership before stopping.
The chase happens twice, with a different chase rule each time:

1. **Four radix angles** → chase **two** triplicity lords (active-sect and
   participating).
2. **Four vital points** → chase **one** lord (domicile only).

This asymmetry — two lords chased for angles, one for vital points — is easy
to misread if you only skim the code; it's deliberate, not a bug.

## Round 1: the four angles

`scoreAnglesWithDispositors` (`almuten/bonatti.go:41-53`):

- Ascendant and MC come directly from `Houses.Ascendant` / `Houses.MC`.
- Descendant and IC aren't returned by `swisseph.CalcHouses` at all
  (`swisseph.HouseResult` only has `Ascendant`/`MC`/`ARMC`/`Vertex`,
  `swisseph/swisseph.go:219-225`) — they're computed as the simple
  oppositions `norm360(Ascendant+180)` and `norm360(MC+180)`
  (`almuten/bonatti.go:42-43`).

Each angle is scored for essential dignity via `addEssentialDignities`
(shared with Figuris — see [almuten-figuris.md](almuten-figuris.md)), then
its sign's **active-sect triplicity lord** *and* **participating triplicity
lord** are each looked up and scored at their own chart position
(`almuten/bonatti.go:47-52`):

```go
for _, lord := range []int{
    int(dignities.ActiveTriplicityRuler(sign, c.IsDiurnal, false)),
    int(dignities.ParticipatingTriplicityRuler(sign)),
} {
    addEssentialDignities(score, c.Positions[lord].Longitude, c.IsDiurnal)
}
```

This is the key structural difference from Figuris, which chases no
triplicity lord at all for its five places. Bonatti's tradition treats both
triplicity lords as worth following; Ibn Ezra's treats triplicity as a
dignity to score in place, not a chain to follow.

## Round 2: the four vital points

`scoreVitalPointsWithDispositor` (`almuten/bonatti.go:59-71`): Sun, Moon,
Part of Fortune, and prenatal syzygy (see
[vital-points-and-sect.md](vital-points-and-sect.md) for how the latter two
are derived). Each is scored for essential dignity, then **only its domicile
lord** is chased to its own position and scored again:

```go
lord := int(dignities.DomicileRuler(sign))
addEssentialDignities(score, c.Positions[lord].Longitude, c.IsDiurnal)
```

No triplicity chase here — domicile only.

## What Bonatti's method omits

No accidental house, no temporal points (weekday/hour rulers), no synodic
points, no besiegement, no fixed stars, no motion or partile-aspect
component. It is purely essential dignity, scored at eight starting points
(four angles + four vital points) and then again at each one's dispositor(s)
— nine to twelve scoring passes total per chart, depending on how many
triplicity lords coincide.

## Comparing all three algorithms

| | Lord of the Geniture (Lilly) | Almuten Figuris (Ibn Ezra) | Almudebit (Bonatti) |
|---|---|---|---|
| Scores | the 7 planets directly | 5 fixed places | 4 angles + 4 vital points, then their dispositors |
| Essential dignity | own position | yes | yes (twice per point) |
| Detriment/fall | yes | no | no |
| Accidental house | Lilly's table | Ibn Ezra's table (different scale/orb) | no |
| Temporal (day/hour ruler) | no | yes | no |
| Synodic (superiors vs. Sun) | no | yes | no |
| Motion (retrograde/swift) | yes | no | no |
| Solar relationship (cazimi/combust) | yes | no | no |
| Partile aspects | yes | no | no |
| Besiegement | yes | no | no |
| Fixed stars | yes | no | no |
| Triplicity lords chased | — | none | active-sect + participating (angles only) |

Back to [almuten-lord-of-geniture.md](almuten-lord-of-geniture.md) or
[almuten-figuris.md](almuten-figuris.md), or continue to
[houses-and-angles.md](houses-and-angles.md) for how house-system choice
feeds into all three algorithms' angle and house calculations.
