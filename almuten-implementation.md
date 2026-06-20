# Almuten Implementation Spec

Implementation reference for two traditional "chart victor" calculations:

1. **Almuten Figuris** (Abraham ibn Ezra) — supreme ruler of the whole figure, scored over five hylegiacal points plus accidental/temporal/synodic bonuses.
2. **Lord of the Geniture** (William Lilly, *Christian Astrology* 1647, p.115) — the planet with the highest net score across a fixed essential + accidental scorecard applied to each planet's own position.

Both reduce to: build a per-planet integer scorecard, return `argmax`. They differ in *what* is scored and *where*.

Source: `./almuten-google-research.md` (Gemini deep research). All point values below are the reconciled numeric values (the research prose rendered them as images; the appended pseudocode carries the explicit numbers).

---

## 1. Scope vs. the current codebase

What the repo already gives us (`swisseph` package):

- `CalcPlanet(tjdUT, planet) (PlanetPos, warn, err)` — longitude, latitude, distance, and **speeds** (so retrograde = `SpeedLon < 0`).
- `CalcHouses(tjdUT, lat, lon, hsys) (HouseResult, err)` — `Cusps[1..12]`, `Ascendant`, `MC`, `ARMC`, `Vertex`.
- `JulDay`, `ZodiacSign`, planet/house-system constants.

What is **missing** and must be added (details in §6):

| Need | Used by | Approach |
|---|---|---|
| Essential dignity tables (domicile, exalt, triplicity, term, face) | both | Static Go tables (§3) |
| Sect (day/night) | both | Sun above/below horizon (§2) |
| Part of Fortune | Ibn Ezra | Formula (§4.1) |
| Prenatal syzygy (prev. new/full Moon) | Ibn Ezra | Backward root-find on Moon−Sun elongation (§6.1) |
| Weekday ruler + planetary hour ruler | Ibn Ezra | Sunrise/sunset via `swe_rise_trans` (§6.2) |
| Solar synodic phase + first station | Ibn Ezra | Elongation + speed sign (§4.4) |
| Combust / cazimi / under-beams | Lilly | Elongation thresholds (§5) |
| Oriental/occidental | Lilly | Sun–planet relative longitude (§5) |
| Fixed-star longitudes (Regulus, Spica, Algol) | Lilly | `swe_fixstar2_ut` (§6.3) |
| Partile aspects, besiegement | Lilly | Pairwise longitude diffs, orb < 1° (§5) |

Recommend a new package `dignities` (pure tables + lookups, no cgo) and a new package `almuten` (orchestration, calls `swisseph` + `dignities`). Keep `swisseph` as the only cgo boundary; add the new C-backed helpers there.

---

## 2. Shared foundations

### Sect

Diurnal if the Sun is above the horizon. Practically: the Sun occupies houses 7–12 (above the Asc→Desc→MC arc).

```
isDiurnal := houseOf(sunLon, cusps) >= 7   // houses 7..12 = above horizon
```

Edge case: when the Sun is within a few degrees of the Asc/Desc, traditional sources differ. Use a simple ≥7 rule by quadrant house, and document it.

### Essential dignity point weights (both methods)

| Dignity | Points |
|---|---|
| Domicile (house/sign ruler) | **5** |
| Exaltation | **4** |
| Triplicity | **3** |
| Term (bound) | **2** |
| Face (decan) | **1** |

Debilities (Lilly only, see §5): Detriment **−5**, Fall **−4**.

> Historical variant (do not implement by default, but note it): some early Latin texts swapped triplicity↔term (term 3, triplicity 2). Alcabitius standardized the 3/2 split above; that is what we use.

---

## 3. Dignity tables (verbatim — implement as static maps)

Sign index 0=Aries … 11=Pisces. Planets: Sun, Moon, Mercury, Venus, Mars, Jupiter, Saturn.

### 3.1 Domicile / Detriment

| Sign | Domicile (+5) | Detriment (−5, Lilly) |
|---|---|---|
| Aries | Mars | Venus |
| Taurus | Venus | Mars |
| Gemini | Mercury | Jupiter |
| Cancer | Moon | Saturn |
| Leo | Sun | Saturn |
| Virgo | Mercury | Jupiter |
| Libra | Venus | Mars |
| Scorpio | Mars | Venus |
| Sagittarius | Jupiter | Mercury |
| Capricorn | Saturn | Moon |
| Aquarius | Saturn | Sun |
| Pisces | Jupiter | Mercury |

### 3.2 Exaltation / Fall (7-planet scheme; nodes excluded)

| Planet | Exalts in (deg) | Falls in |
|---|---|---|
| Sun | Aries (19°) | Libra |
| Moon | Taurus (3°) | Scorpio |
| Mercury | Virgo (15°) | Pisces |
| Venus | Pisces (27°) | Virgo |
| Mars | Capricorn (28°) | Cancer |
| Jupiter | Cancer (15°) | Capricorn |
| Saturn | Libra (21°) | Aries |

Signs with no planetary exaltation (Gemini, Leo, Sagittarius, Aquarius) award 0 exaltation points. The exact degree only matters for accidental "exact exaltation" niceties — for the almuten, presence in the sign suffices for the +4.

### 3.3 Triplicity (Dorothean — 3 lords: day / night / participating)

| Element | Signs | Day | Night | Participating |
|---|---|---|---|---|
| Fire | Aries, Leo, Sagittarius | Sun | Jupiter | Saturn |
| Earth | Taurus, Virgo, Capricorn | Venus | Moon | Mars |
| Air | Gemini, Libra, Aquarius | Saturn | Mercury | Jupiter |
| Water | Cancer, Scorpio, Pisces | Venus | Mars | Moon |

**Ibn Ezra** uses the single *active-sect* lord (day lord if diurnal, else night lord) → one planet gets +3. The participating lord is excluded (following Abu Ma'shar/Bonatti's restriction to active sect rulers).

**Lilly** uses a 2-lord (day/night) table where **Water = Mars day & night** (Lilly's variant). For the Lord-of-Geniture scorecard, award +3 to the active-sect lord only.

> Implementation: keep both — `triplicityDorothean[element] = {day, night, participating}` and a `lillyWater` override. Make the "include participating?" and "Lilly water" choices flags so Bonatti-style extensions (§7) can reuse the table.

### 3.4 Terms / Bounds (Egyptian — the set Lilly uses)

Each cell: planet rules from the previous boundary up to the listed degree.

| Sign | | | | | |
|---|---|---|---|---|---|
| Aries | Jup→6 | Ven→12 | Mer→20 | Mar→25 | Sat→30 |
| Taurus | Ven→8 | Mer→14 | Jup→22 | Sat→27 | Mar→30 |
| Gemini | Mer→6 | Jup→12 | Ven→17 | Mar→24 | Sat→30 |
| Cancer | Mar→7 | Ven→13 | Mer→19 | Jup→26 | Sat→30 |
| Leo | Jup→6 | Ven→11 | Sat→18 | Mer→24 | Mar→30 |
| Virgo | Mer→7 | Ven→17 | Jup→21 | Mar→28 | Sat→30 |
| Libra | Sat→6 | Mer→14 | Jup→21 | Ven→28 | Mar→30 |
| Scorpio | Mar→7 | Ven→11 | Mer→19 | Jup→24 | Sat→30 |
| Sagittarius | Jup→12 | Ven→17 | Mer→21 | Sat→26 | Mar→30 |
| Capricorn | Mer→7 | Jup→14 | Ven→22 | Sat→26 | Mar→30 |
| Aquarius | Mer→7 | Ven→13 | Jup→20 | Mar→25 | Sat→30 |
| Pisces | Ven→12 | Jup→16 | Mer→19 | Mar→28 | Sat→30 |

Lookup: given `degInSign`, the term lord is the first row entry whose upper bound `> degInSign`.

### 3.5 Faces / Decans (Chaldean order, 10° each)

| Sign | 0–10° | 10–20° | 20–30° |
|---|---|---|---|
| Aries | Mars | Sun | Venus |
| Taurus | Mercury | Moon | Saturn |
| Gemini | Jupiter | Mars | Sun |
| Cancer | Venus | Mercury | Moon |
| Leo | Saturn | Jupiter | Mars |
| Virgo | Sun | Venus | Mercury |
| Libra | Moon | Saturn | Jupiter |
| Scorpio | Mars | Sun | Venus |
| Sagittarius | Mercury | Moon | Saturn |
| Capricorn | Jupiter | Mars | Sun |
| Aquarius | Venus | Mercury | Moon |
| Pisces | Saturn | Jupiter | Mars |

(Chaldean sequence Saturn→Jupiter→Mars→Sun→Venus→Mercury→Moon, starting at Aries I = Mars.)

### 3.6 Chaldean order & weekday rulers (for Ibn Ezra temporal points)

- Chaldean (slowest→fastest): **Saturn, Jupiter, Mars, Sun, Venus, Mercury, Moon**.
- Weekday day-rulers: Sun→Sunday, Moon→Monday, Mars→Tuesday, Mercury→Wednesday, Jupiter→Thursday, Venus→Friday, Saturn→Saturday. The astrological day begins at **local sunrise**, not midnight (§6.2).

---

## 4. Algorithm 1 — Ibn Ezra's Almuten Figuris

Score the seven planets across five vital places plus accidental, temporal, and synodic bonuses; the highest total is the Almuten Figuris.

### 4.1 The five vital places (longitudes)

1. **Ascendant** — `houses.Ascendant`.
2. **Sun** — `CalcPlanet(jd, Sun).Longitude`.
3. **Moon** — `CalcPlanet(jd, Moon).Longitude`.
4. **Part of Fortune** (sect-dependent):
   - Diurnal: `PoF = (Asc + Moon − Sun) mod 360`
   - Nocturnal: `PoF = (Asc + Sun − Moon) mod 360`
5. **Prenatal syzygy** — the exact degree of the lunation immediately *before* birth (§6.1):
   - If it was a **New Moon** (conjunctional): use the conjunction longitude (Sun=Moon).
   - If it was a **Full Moon** (preventional): use the longitude of **whichever luminary was above the horizon** at the moment of that opposition. (Requires recomputing houses at the syzygy instant to test which luminary is above horizon — or, acceptably, evaluate the horizon at the *birth* location/time of the syzygy.)

### 4.2 Essential scoring over the five places

For each of the five longitudes, find the five dignity rulers and add weights to those planets:

```
for each place in [Asc, Sun, Moon, PoF, Syzygy]:
    sign      = signOf(place)
    deg       = place mod 30
    add 5 to domicileRuler(sign)
    add 4 to exaltationRuler(sign)         // skip if sign has none
    add 3 to triplicityRuler(sign, isDiurnal)   // active-sect lord only
    add 2 to termRuler(sign, deg)
    add 1 to faceRuler(sign, deg)
```

A planet can collect from multiple places and multiple dignity types; sum freely.

### 4.3 Accidental house-placement points

Each planet gains points for the house it occupies (note: this is a *non-standard, almuten-specific* house valuation, distinct from Lilly's):

| House | 1 | 10 | 7 | 4 | 11 | 5 | 2 | 9 | 8 | 3 | 12 | 6 |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| Points | 12 | 11 | 10 | 9 | 8 | 7 | 6 | 5 | 4 | 3 | 2 | 1 |

> Optional **Hermann of Carinthia variant** (do not default): 7→8, 4→7, 11→6, 5→5, 2→4, 9→10, 8→3, 3→12 (others unchanged). Expose as a flag if desired.

House assignment uses a **5° pre-cusp orb**: a planet within 5° *before* the next cusp (in zodiacal order) counts as already in that next house. Otherwise use the normal containing house.

```
func adjustedHouse(planetLon, cusps):
    for h in 1..12:
        nextCusp = cusps[h % 12 + 1]
        if ((nextCusp - planetLon) mod 360) <= 5.0:
            return h+1 (wrapped)
    return containingHouse(planetLon, cusps)
```

Be careful with the modulo near 0°/360° and with the unequal-house wrap (cusp[12]→cusp[1]).

### 4.4 Temporal points

- Weekday ruler (§3.6, day starts at sunrise): **+7**.
- Planetary-hour ruler at the birth moment (§6.2): **+6**.

### 4.5 Synodic (solar phase) points — superior planets only (Saturn, Jupiter, Mars)

Ibn Ezra's simplified rule, applied when the Sun is **separating** from the planet, by elongation `d`:

| Elongation d | Points |
|---|---|
| 15° ≤ d < 60° | 3 |
| 60° ≤ d < 90° | 2 |
| 90° ≤ d < first station | 1 |

"Separating" = the faster Sun is moving away from the planet (Sun ahead in zodiacal order, increasing distance). "First station" = the elongation at which the planet turns retrograde; approximate per-planet (Saturn ≈ 109°, Jupiter ≈ 115°, Mars ≈ 136°) or detect dynamically by scanning for `SpeedLon` sign change. The research notes a more granular Pseudo-John-of-Seville table; treat that as an optional variant.

### 4.6 Result

`AlmutenFiguris = argmax(scorecard)`. Return the full scorecard too (ties and runners-up are astrologically meaningful). Tie-break policy: report all tied planets rather than picking arbitrarily.

### 4.7 Reference pseudocode (from the research, cleaned)

```python
weights = {"domicile":5,"exaltation":4,"triplicity":3,"term":2,"face":1}
housePts = {1:12,10:11,7:10,4:9,11:8,5:7,2:6,9:5,8:4,3:3,12:2,6:1}

places = [asc, sun, moon, pof, syzygy]
for deg in places:
    s, d = signOf(deg), deg % 30
    rulers = {"domicile":domicile(s),"exaltation":exalt(s),
              "triplicity":activeTriplicity(s,diurnal),
              "term":term(s,d),"face":face(s,d)}
    for k, p in rulers.items():
        if p: score[p] += weights[k]

for p in planets:
    score[p] += housePts[adjustedHouse(lon[p], cusps)]

score[dayRuler]  += 7
score[hourRuler] += 6

for p in ["Saturn","Jupiter","Mars"]:
    if separating(sun, p):
        d = elongation(sun, lon[p])
        if   15 <= d < 60: score[p] += 3
        elif 60 <= d < 90: score[p] += 2
        elif 90 <= d < firstStation(p): score[p] += 1

return argmax(score), score
```

---

## 5. Algorithm 2 — Lilly's Lord of the Geniture

Apply a fixed scorecard to **each planet at its own position**. Highest net total wins. Scale runs roughly −15…+15. This is *Christian Astrology* p.115.

### 5.1 Full fortitude/debility table

| Fortitude | + | Debility | − |
|---|---|---|---|
| In domicile (or mutual reception by sign) | 5 | In detriment | −5 |
| In exaltation (or mutual reception by exaltation) | 4 | In fall | −4 |
| In own triplicity | 3 | Peregrine | −5 |
| In own term | 2 | Retrograde | −5 |
| In own face | 1 | Slow in motion | −2 |
| In 10th or 1st house | 5 | In 12th house | −5 |
| In 7th, 4th, or 11th house | 4 | In 8th or 6th house | −2 |
| In 2nd or 5th house | 3 | Superior (Sa/Ju/Ma) occidental | −2 |
| In 9th house | 2 | Inferior (Me/Ve) oriental | −2 |
| In 3rd house | 1 | Moon waning | −2 |
| Direct in motion | 4 | Combust (within 8°30′ of Sun) | −5 |
| Swift in motion | 2 | Under Sun's beams (within 17°) | −4 |
| Superior (Sa/Ju/Ma) oriental | 2 | Partile conjunction Saturn or Mars | −5 |
| Inferior (Me/Ve) occidental | 2 | Partile conjunction South Node | −4 |
| Moon waxing | 2 | Besieged by Saturn and Mars | −4 |
| Cazimi (within 17′ of Sun) | 5 | Partile opposition Saturn or Mars | −4 |
| Partile conjunction Jupiter or Venus | 5 | Partile square Saturn or Mars | −4 |
| Partile conjunction North Node | 4 | Conjunct Algol (within 5°) | −4 |
| Partile trine Jupiter or Venus | 4 | | |
| Partile sextile Jupiter or Venus | 3 | | |
| Conjunct Regulus | 6 | | |
| Conjunct Spica | 5 | | |

### 5.2 Rules / clarifications for implementation

- **Peregrine**: a planet with *no* essential dignity (not in its own domicile, exalt, triplicity, term, or face) at its position → −5. Mutually exclusive with term/face awards in the research pseudocode (if peregrine, skip the +2/+1).
- **Mutual reception** (by sign or exaltation) counts as if the planet were in that dignity → award the domicile (+5) / exaltation (+4) bonus. Detecting it requires the *other* planets' positions, so compute all longitudes first.
- **Motion** (non-luminaries only): direct +4 / retrograde −5; swift +2 / slow −2. Swift vs. slow = compare `|SpeedLon|` to the planet's mean daily motion (Saturn ~0.034°, Jupiter ~0.083°, Mars ~0.524°, Sun ~0.986°, Venus ~1.2°, Mercury ~1.383°, Moon ~13.176°). Luminaries get no motion score (Sun never retrograde; Moon handled via waxing/waning).
- **Solar relationship** (`d` = elongation from Sun, non-Sun planets):
  - `d ≤ 0.283°` (17′) → cazimi **+5**
  - else `d ≤ 8.5°` (8°30′) → combust **−5**
  - else `d ≤ 17°` → under beams **−4**
  - else → free of the Sun **+5** (the research scores "neither combust nor under beams" as +5; this is the pseudocode's behavior — flag it, as some editions of Lilly do not award a positive here).
- **Oriental/occidental**: oriental = rising before the Sun (lower longitude / east of Sun in the morning sky). Superiors get +2 oriental / −2 occidental; inferiors (Mercury, Venus) get +2 occidental / −2 oriental. Implement via the sign of the normalized `(planetLon − sunLon)` reduced to (−180,180].
- **Moon waxing/waning**: waxing (Sun→Moon elongation increasing toward full, i.e. Moon ahead of Sun by 0–180°) → +2, else −2.
- **Houses** (Lilly's valuation, *different from §4.3*): 1/10 +5; 4/7/11 +4; 2/5 +3; 9 +2; 3 +1; 12 −5; 6/8 −2. Use the plain containing house (no 5° orb here).
- **Partile aspect** = exact to within 1° (same degree number). Aspects to Jupiter/Venus are benefic (+), to Saturn/Mars malefic (−). Conjunction/trine/sextile vs. opposition/square as tabled.
- **Besieged by Saturn and Mars**: planet positioned between Saturn and Mars (enclosed by body or ray) with no intervening aspect → −4.
- **Fixed stars**: conjunct Regulus +6, Spica +5, Algol −4. Use a tight orb (Lilly ~5° for Algol; use ~2–5° configurable). Compute star longitudes dynamically (§6.3), do **not** hardcode — precession moves them ~1°/72yr.

### 5.3 Reference pseudocode

See `./almuten-google-research.md` lines ~256–373 for the full annotated Python; it matches §5.1/§5.2 exactly. Port each branch 1:1, but make the three flagged choices explicit toggles:
1. "+5 when free of Sun" (some editions: 0),
2. peregrine suppresses term/face,
3. fixed-star orb.

`LordOfGeniture = argmax(score)`; return the full scorecard.

---

## 6. Swiss Ephemeris additions (cgo, in `swisseph` package)

### 6.1 Prenatal syzygy finder

No existing helper. Find the last instant before birth where the Moon–Sun elongation hits 0° (new) or 180° (full).

Approach (no extra C entry points needed — reuse `CalcPlanet`):
```
elong(t) = normalize180( moonLon(t) - sunLon(t) )   // (-180,180]
```
- Step backward from birth in coarse steps (e.g. 0.5 day) tracking `elong`; a sign change of `elong` brackets a **new moon**, a crossing of ±180 brackets a **full moon**.
- Take the most recent crossing of *either* type; bisect to ~1-second precision.
- Type = new if it was the 0° crossing, full if the 180° crossing.
- For a full moon, recompute houses at the syzygy instant (need the birth geo-coords) to choose the above-horizon luminary.

Optionally expose `swe_mooncross_ut` / `swe_helio_cross` if the bundled headers include them, but the bisection on `CalcPlanet` is dependency-free and sufficient.

### 6.2 Sunrise/sunset → weekday & planetary hours

Add a binding for `swe_rise_trans` (rise/set) — present in `swephexp.h`:

```c
int swe_rise_trans(double tjd_ut, int ipl, char *starname, int epheflag,
                   int rsmi, double *geopos, double atpress, double attemp,
                   double *tret, char *serr);
```
- `rsmi = SE_CALC_RISE | SE_BIT_DISC_CENTER` for sunrise; `SE_CALC_SET` for sunset.
- `geopos = {lon, lat, alt}`.

Then:
1. Find the sunrise that opens the astrological day containing the birth instant (the most recent sunrise ≤ birth). The **weekday ruler** is the planet of that civil date's weekday (§3.6), evaluated at that sunrise (a birth after midnight but before sunrise still belongs to the previous weekday).
2. Daylight hour length = (sunset − sunrise)/12; night hour = (next sunrise − sunset)/12.
3. The **first hour** after sunrise is ruled by the day-ruler; subsequent hours follow Chaldean order (§3.6), wrapping across day/night. Identify which of the 24 unequal hours contains the birth instant → **hour ruler**.

Edge cases: polar latitudes where the Sun does not rise/set — `swe_rise_trans` returns a flag; fall back to equal 12/12 split of the calendar day or document as unsupported.

### 6.3 Fixed stars

Add a binding for `swe_fixstar2_ut`:
```c
int swe_fixstar2_ut(char *star, double tjd_ut, int iflag, double *xx, char *serr);
```
Call with `star = "Regulus"`, `"Spica"`, `"Algol"` (Swiss Ephemeris resolves traditional names via `sefstars.txt` — confirm that file ships in `ephe/`; if not, add it or hardcode J2000 longitudes with precession applied). Use `xx[0]` (longitude) for the conjunction test.

### 6.4 Stations (optional, for Ibn Ezra §4.5)

Either hardcode approximate first-station elongations per superior planet, or detect by scanning `SpeedLon` for the sign flip around the birth date. Hardcoding is adequate for the coarse 3/2/1 banding.

---

## 7. Suggested repo structure & API

```
dignities/            # pure Go, no cgo, fully unit-testable
  tables.go           # domicile, detriment, exalt, fall, triplicity, terms, faces, Chaldean
  lookup.go           # DomicileRuler(sign), TermRuler(sign,deg), ActiveTriplicity(sign,diurnal), ...
  tables_test.go      # assert every sign sums correctly; spot-check known degrees

almuten/
  figuris.go          # IbnEzraAlmuten(chart) (Scorecard, []Planet, error)
  geniture.go         # LordOfGeniture(chart) (Scorecard, []Planet, error)
  chart.go            # gathers inputs once: planet positions, houses, sect, PoF, syzygy, hour ruler
  *_test.go           # golden-chart tests against published worked examples

swisseph/
  swisseph.go         # + RiseTrans, FixStar, (optional) PrenatalSyzygy helpers
```

Suggested types:

```go
type Scorecard map[Planet]int   // Planet = existing swisseph IDs

type Chart struct {
    JD          float64
    Lat, Lon    float64
    IsDiurnal   bool
    Positions   map[Planet]swisseph.PlanetPos
    Houses      swisseph.HouseResult
    PartFortune float64
    Syzygy      float64        // longitude
    SyzygyType  string         // "new" | "full"
    DayRuler    Planet
    HourRuler   Planet
}
```

CLI surface (mirrors existing flags): add `--almuten=figuris|geniture` (or a subcommand) that prints the scorecard table and the victor, with `--json` and `--verbose` honoring the existing renderers (`output` package). Build the `Chart` once and feed both algorithms.

Per `CLAUDE.md`: build/test exclusively via `make` / `make test`.

---

## 8. Open decisions / variants to surface as flags or config

1. **Triplicity participating lord** — excluded for Ibn Ezra by default (Abu Ma'shar/Bonatti); Lilly uses 2-lord with Water=Mars/Mars. Keep both tables.
2. **Lilly "free of the Sun" = +5** — present in the research pseudocode; not in every edition of *Christian Astrology*. Default on, toggleable.
3. **Peregrine suppresses term/face** — research pseudocode does this; make it a flag.
4. **House-point variant** (Hermann of Carinthia) for Ibn Ezra accidental points — default off.
5. **Sect threshold** near the horizon — simple quadrant-house rule; document.
6. **Fixed-star orbs** — default 5° (Algol per Lilly), configurable.
7. **Ties** — report all tied planets; never silently pick one.

## 9. Validation

- Unit-test every dignity table: each sign's 5 terms must cover 0–30 contiguously; each sign's 3 faces cover 0–30; domicile/detriment and exalt/fall are exact opposites.
- Golden charts: pick 2–3 published worked examples (e.g. the Derek Appleby horary referenced in the research; any chart with a published Almuten Figuris) and assert the full scorecard, not just the victor.
- Cross-check Part of Fortune and prenatal-syzygy degrees against an established astrology program for several dates and both sects.

## 10. Source map (into `./almuten-google-research.md`)

- Essential dignity weights & dispositorship metaphor — lines 14–26.
- Ibn Ezra five places, PoF, syzygy rules — lines 30–41.
- House-point table (standard + Hermann variant) & 5° orb rule — lines 44–65.
- Day/hour ruler points + synodic banding — lines 67–68.
- Lilly p.115 full table — lines 102–128.
- Ibn Ezra pseudocode — lines 162–249.
- Lilly pseudocode — lines 253–373.
- Topical almutens (future extension) — lines 83–96.
- Bonatti Almudebit (future extension) — lines 70–81.
