# Implementation Notes

Traditional astrology rarely has a single canonical version of any
technique — different sources disagree on orbs, on which lords to chase, on
how strict a debility should be. This file consolidates every place in this
codebase where a specific choice was made among several traditional or
plausible variants, so the other docs can explain *what the code does*
without constantly hedging about what else it could have done. Each
algorithm file also flags its own entries inline, linking back here.

| Choice | What this codebase does | What varies in tradition | Where | Configurable? |
|---|---|---|---|---|
| Water-triplicity lords | Lilly's override: Mars/Mars/Moon | Dorothean: Venus(day)/Mars(night)/Moon | `dignities/tables.go:105-107` | Yes — `Options.LillyWater` (default `true`), Geniture only |
| Peregrine strictness | Lacking all 3 major dignities ⇒ peregrine, term/face suppressed | Stricter reading: peregrine only if *all five* dignities absent | `almuten/geniture.go:75-94`, `almuten/chart.go:53` | Yes — `Options.PeregrineSuppressesTermFace` (default `true`) |
| Fixed star orb | 5° conjunction orb | Orbs of 1°–10° are all used across different sources | `almuten/geniture.go:204`, `almuten/chart.go:54` | Yes — `Options.FixedStarOrb` |
| Free-of-Sun bonus magnitude | +5 | The other three solar bands (cazimi/combust/under beams) are fixed at +5/−5/−4 with no equivalent toggle — an asymmetry worth noting | `almuten/geniture.go:120`, `almuten/chart.go:52` | Yes — `Options.FreeOfSunBonus`; the other three bands are not |
| Figuris accidental-house pre-cusp orb | 5°, hardcoded | No toggle exists, unlike Geniture's options | `almuten/figuris.go:100` | No |
| Besiegement definition | Minor arc ≤ 30° between Saturn/Mars, no intervening planet | Traditional "besieged by ray" is a fuller, aspect-based condition; the code comment itself flags this as a simplification | `almuten/geniture.go:250-253` | No |
| Synodic banding constants | `firstStationElongation`: Saturn 109°, Jupiter 115°, Mars 136° | Approximate bounding values, not exact ephemeris-derived stations for a specific chart | `almuten/figuris.go:18-24` | No |
| Fixed star catalogue size | Only Regulus, Spica, Algol scored | A much larger traditional/Behenian fixed-star catalogue exists | `almuten/geniture.go:22-30` | The set is not; only the orb is |
| Term (bound) system | Egyptian bounds only | Ptolemy's *Tetrabiblos* revises several boundary degrees from the Egyptian table | `dignities/tables.go:109-131` | No — single table, no toggle |
| Planet set | Only the 7 traditional planets (Sun–Saturn) | No outer planets (Uranus/Neptune/Pluto) or asteroids/lots beyond Part of Fortune | `almuten/chart.go:17-20` | No — architectural, not a runtime option |
| Bonatti triplicity chase | Active-sect *and* participating lord, for angles only | Figuris chases neither; vital points in Bonatti's own method chase only the domicile lord, not triplicity | `almuten/bonatti.go:47-52` vs. `almuten/bonatti.go:66-70` | No |

If you're auditing this codebase against a specific historical source and
find a mismatch not listed here, it's worth checking whether it's a genuine
divergence (worth adding to this table) or a difference between two
historical sources that simply disagree with each other (worth noting in
the relevant algorithm file instead).

## Historical claims flagged for review

A pass of outside web research against the prose (not the scoring tables —
those checked out: Egyptian terms, Lilly's house-point table against
*Christian Astrology* p.115, the Dorothean/Lilly water-triplicity divergence,
Ibn Ezra's five places and house table, the cazimi/combust/under-the-beams
orbs, and the Chaldean-order transmission chain all matched independent
sources) turned up two claims in
[houses-and-angles.md](houses-and-angles.md) that outside sources
contradict or complicate. Implementation is unaffected either way — both
items are about the historical prose, not the `swisseph.CalcHouses` call
itself.

- **Koch described as "a Placidus variant... adjusted for high-latitude
  reliability."** Multiple independent sources (Wikipedia's Walter Koch
  entry, Astrodienst-adjacent dictionaries, and astrology-history writeups)
  describe Koch instead as a modification of the **Alcabitius** system, and
  several state Koch shares Placidus's *same* inability to calculate at
  polar latitudes rather than improving on it. The "Placidus variant" framing
  isn't baseless — Koch is time-based like Placidus and some sources do call
  it "a more complicated version of Placidus" — but "adjusted for
  high-latitude reliability" appears to overstate things; worth softening or
  re-sourcing.
- **Regiomontanus as "a refinement of Campanus, not an independent
  invention"** stated as settled fact. The historical record is actually
  contested: Pico della Mirandola accused Regiomontanus of plagiarizing
  Campanus, but historian John North has argued this reflects patriotic bias
  (German vs. Italian) rather than established derivation. Worth rephrasing
  as a documented accusation rather than a settled lineage.

One lower-confidence item, not corroborated well enough to act on: a couple
of SEO-grade astrology sites attribute the term "Almudebit" directly to
Bonatti rather than to a later translator, which would resolve
[almuten-bonatti.md](almuten-bonatti.md)'s existing hedge — but the sourcing
wasn't strong enough to treat as confirmed. Worth checking directly against
Dykes's translation if anyone revisits that file.

Back to [docs/README.md](README.md).
