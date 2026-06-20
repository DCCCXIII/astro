# Astrological Concepts in `astro`

This directory explains *why* the calculations in this codebase matter
astrologically — the historical sources and scoring logic behind each
feature — complementing the root [README.md](../README.md), which covers
the API/CLI reference. Written for readers who already know basic astrology
terms (sign, house, dignity); each file cross-references the actual Go code
(file:line) that implements what it describes.

## Reading order

The almuten files depend on the two foundational files before them — read in
this order if you're new to the codebase:

1. [dignities.md](dignities.md) — the five essential-dignity tables every
   algorithm scores against.
2. [vital-points-and-sect.md](vital-points-and-sect.md) — sect, Part of
   Fortune, and prenatal syzygy, shared by all three almuten algorithms.
3. [almuten-lord-of-geniture.md](almuten-lord-of-geniture.md) — Lilly's
   algorithm.
4. [almuten-figuris.md](almuten-figuris.md) — Ibn Ezra's algorithm.
5. [almuten-bonatti.md](almuten-bonatti.md) — Bonatti's algorithm.
6. [houses-and-angles.md](houses-and-angles.md) — the six house systems and
   four angles, and how house choice feeds into the almuten scores above.
7. [glossary.md](glossary.md) — alphabetical reference for named conditions
   (cazimi, peregrine, besiegement, etc.).
8. [implementation-notes.md](implementation-notes.md) — every place this
   codebase makes a specific implementation choice among several
   traditional variants.

## The three almuten algorithms at a glance

| | Lord of the Geniture | Almuten Figuris | Almudebit |
|---|---|---|---|
| Source | William Lilly, *Christian Astrology* (1647) p.115 | Abraham Ibn Ezra, 12th c. (*Reshit Hokhmah* tradition) | Guido Bonatti, *Liber Astronomiae* (c. 1277) |
| Philosophy | score the 7 planets directly | score 5 fixed hylegiacal places | score 4 angles + 4 vital points, then chase their dispositors |
| Scores | essential dignity, debility, motion, solar relationship, partile aspect, besiegement, fixed stars, accidental house | essential dignity at 5 places, accidental house, temporal rulers, synodic bands | essential dignity at 8 points, twice (own position + dispositor) |
| Omits | temporal/synodic points | debility, motion, solar relationship, partile aspect, besiegement, fixed stars | accidental house, temporal, synodic, motion, solar relationship, partile aspect, besiegement, fixed stars |

Full breakdown: [almuten-bonatti.md](almuten-bonatti.md#comparing-all-three-algorithms).

## Sources at a glance

- **William Lilly** — *Christian Astrology* (1647). Source for Lord of the
  Geniture's full scorecard, including its accidental-house table and most
  named conditions in [glossary.md](glossary.md).
- **Abraham Ibn Ezra** — 12th c., *Reshit Hokhmah* tradition. Source for
  Almuten Figuris's five-place + temporal + synodic scoring.
- **Guido Bonatti** — *Liber Astronomiae*, c. 1277. Source for the
  Almudebit's two-round dispositor-chasing method.
- **Dorotheus of Sidon** — *Carmen Astrologicum*, 1st c. CE. Source for the
  day/night/participating triplicity lords used by all three algorithms.
- **Vettius Valens** — 2nd c. CE *Anthologies*. Contemporary source for the
  same Egyptian-bounds and triplicity tradition; also the earliest surviving
  detailed source on the Lot of Fortune.
- **Egyptian bounds** — attributed by Valens to "the Egyptian sages,"
  likely the lost work ascribed to Nechepso/Petosiris; this codebase
  implements this table rather than Ptolemy's later revision.
- **Chaldean order** — the planets sequenced by apparent speed
  (Saturn→...→Moon); used here both for decan/face rulers and for planetary
  hours and weekday rulers in Almuten Figuris.

See [dignities.md](dignities.md), [almuten-figuris.md](almuten-figuris.md),
and [almuten-bonatti.md](almuten-bonatti.md) for where each source is
discussed in depth.
