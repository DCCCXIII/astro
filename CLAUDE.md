# CLAUDE.md

## Project Overview

Go wrapper around the Swiss Ephemeris C library for astrological calculations, paired with a unix-like CLI to compute planetary positions, house cusps, and traditional "chart victor" rulers (almutens) for a given time and geographic location.

## Repository Structure

```
astro/
├── main.go              # Minimal entry point — delegates to cmd.Run
├── cmd/
│   ├── run.go           # CLI flag parsing, validation, orchestration
│   └── run_test.go      # Tests for flag parsing (house system, almuten mode)
├── output/
│   ├── result.go        # Result type + Build()/BuildAlmuten() — all swisseph calls live here
│   ├── text.go          # PrintText() — human-readable renderer
│   └── json.go          # PrintJSON() — JSON renderer
├── dignities/           # Pure-Go essential-dignity tables + lookups (no cgo)
│   ├── tables.go        # domicile, exalt, triplicity, terms, faces, Chaldean order
│   └── lookup.go        # DomicileRuler, TermRuler, ActiveTriplicityRuler, ...
├── almuten/             # Chart-victor algorithms (orchestrates swisseph + dignities)
│   ├── chart.go         # BuildChart, Chart, Scorecard, Options, geometric helpers
│   ├── geniture.go      # LordOfGeniture — Lilly's Christian Astrology p.115 scorecard
│   ├── figuris.go       # AlmutenFiguris — Ibn Ezra's five-place scorecard
│   ├── bonatti.go       # AlmutenBonatti — Bonatti's Almudebit (angles + vital points + dispositors)
│   ├── syzygy.go        # Prenatal new/full-moon finder (pure Go on CalcPlanet)
│   └── search.go        # FindLongitude — bidirectional search for a planet's next/last occurrence at a longitude
├── swisseph/
│   ├── swisseph.go      # Go cgo bindings to Swiss Ephemeris
│   ├── swisseph_test.go # Tests for the swisseph package
│   ├── *.c / *.h        # Bundled Swiss Ephemeris C source (no external install needed)
├── ephe/                # Binary ephemeris data files (~105 MB, .se1 + sefstars.txt)
├── go.mod               # module github.com/dcccxiii/astro, go 1.25
└── README.md
```

Dependency direction: `cmd` → `output` → `almuten` → `dignities` + `swisseph`.
`swisseph` is the only cgo boundary; `dignities` is pure Go (fully unit-testable
without the C library).

## Commands

ALWAYS use these make targets instead of raw Go commands:

- **Build:** `make` (runs `go fmt`, `go vet`, `go test`, then `go build`)
- **Test:** `make test` (runs `go fmt`, `go vet`, then `go test -v ./...`)
- **Format only:** `make fmt`
- **Vet only:** `make vet`

Never run `go build`, `go test`, `go fmt`, or `go vet` directly.

Requires Go 1.25+ and a C compiler (GCC or Clang). No external C library installation needed — Swiss Ephemeris C sources are bundled. cgo compiles them automatically via directives in `swisseph/swisseph.go`.

## CLI Usage

```bash
astro [--house-system <system>] [--almuten <mode>] [--search <planet>:<longitude>[:<direction>]] [--json] [--verbose] <datetime> <lat> <lon>
```

- `<datetime>`: UTC time in ISO 8601 (e.g. `2024-03-20T12:00:00Z`)
- `<lat>`: Decimal degrees, north positive
- `<lon>`: Decimal degrees, east positive
- `--house-system`: `placidus` (default), `koch`, `whole-sign`, `regiomontanus`, `equal`, `campanus`
- `--almuten`: `geniture` (Lilly's Lord of the Geniture), `figuris` (Ibn Ezra's Almuten Figuris), `bonatti` (Bonatti's Almudebit), `both` (geniture+figuris), or `all`. Omitted = no almuten section.
- `--search`: `<planet>:<longitude>[:<direction>]` (e.g. `mars:15aries`, `mars:195`, or `mars:15aries:forward`) — finds the nearest time `<planet>` was/will be at that ecliptic longitude relative to `<datetime>`. `<direction>` is `backward` (default, most recent occurrence before `<datetime>`) or `forward` (next occurrence after `<datetime>`). Longitude accepts raw degrees `[0,360)` or sign+degree `[0,30)`. Omitted = no search section.
- `--json`: Output JSON instead of human-readable text
- `--verbose`: Include ecliptic latitude, distance, speed components, ARMC, Vertex, the almuten scoreboard table, search speed/retrograde/house, and ephemeris source warning (if Moshier fallback is active). Without `--verbose`, almuten output shows only the victor(s) and search output omits speed/house.

## Package Overview

### `cmd`

`Run(args []string) error` is the real entry point. It parses flags with `flag.NewFlagSet`, validates arguments, resolves the ephemeris path relative to the executable, and delegates to the `output` package.

### `output`

Three files with a clean separation of concerns:

- **`result.go`** — `Build()` assembles the planetary/house `Result`; `BuildAlmuten()` builds an `almuten.Chart` once and runs the requested algorithm(s), returning `[]AlmutenEntry`. Neither renderer touches the C library.
- **`text.go`** — `PrintText(r Result, verbose bool) error` writes human-readable output to stdout.
- **`json.go`** — `PrintJSON(r Result, verbose bool) error` marshals to indented JSON and writes to stdout.

### `dignities`

Pure-Go static tables (domicile/detriment, exaltation/fall, Dorothean triplicity + Lilly's Mars/Mars water override, Egyptian terms, Chaldean faces, Chaldean order, weekday rulers, mean daily motion) with lookup functions. No cgo — keyed by `dignities.Planet` whose int values match the `swisseph` planet IDs.

### `almuten`

Orchestration. `BuildChart` gathers positions, lunar node, houses, and sect once into a `Chart`. `LordOfGeniture`, `AlmutenFiguris`, and `AlmutenBonatti` each return a `Scorecard` (`map[int]int`) plus the tied winners. Part of Fortune and prenatal syzygy are shared via `Chart.computeVitalPoints()`; Figuris additionally derives weekday/planetary-hour rulers. Bonatti scores the four angles and four vital points, then chases each angle's triplicity lords and each vital point's domicile lord to their own chart position for a second scoring pass — no accidental house, temporal, or synodic points. `Options` exposes the spec's variant toggles (`DefaultOptions()` for the recommended defaults). `FindLongitude` generalizes `PrenatalSyzygy`'s scan-then-bisect pattern to any of the seven planets, an arbitrary target longitude, and either scan direction (`Backward`/`Forward`); per-planet search windows account for retrograde loops and orbital period (Moon: 40 days, Saturn: 11500 days).

### `swisseph`

Low-level cgo bindings. All C calls are mutex-protected for thread safety. Callers never interact with C types directly.

## Key Go API

### `swisseph` package

| Function | Description |
|---|---|
| `SetEphePath(path)` | Set path to `ephe/` directory |
| `Close()` | Free C library resources |
| `JulDay(year, month, day, hour)` | Calendar date → Julian Day |
| `RevJul(jd)` | Julian Day → calendar date (exact inverse of `JulDay`) |
| `CalcPlanet(tjdUT, planet)` | Planet position at Julian Day; second return value is a warning string (non-empty when Moshier fallback is active) |
| `CalcHouses(tjdUT, lat, lon, hsys)` | House cusps for location/time |
| `FixStar(name, tjdUT)` | Fixed-star position by catalogue name (e.g. `"Regulus"`); resolves via `ephe/sefstars.txt` |
| `RiseTrans(tjdUT, planet, lat, lon, alt, rising)` | Next rise/set time as a Julian Day; returns `ErrNoRiseSet` at polar latitudes |
| `ZodiacSign(longitude)` | Ecliptic longitude → sign name + degree (normalises to [0, 360) automatically) |

**Planet IDs:** `swisseph.Sun`, `Moon`, `Mercury`, `Venus`, `Mars`, `Jupiter`, `Saturn`, `MeanNode`

**House system bytes:** `HousePlacidus='P'`, `HouseKoch='K'`, `HouseWholeSign='W'`, `HouseRegiomontanus='R'`, `HouseEqual='A'`, `HouseCampanus='C'`

### `output` package

| Function | Description |
|---|---|
| `Build(jd, planets, lat, lon, hsys, hsysName)` | Compute full chart; returns `Result` or error |
| `BuildAlmuten(jd, lat, lon, hsys, mode)` | Compute `[]AlmutenEntry` for `mode` in `geniture`/`figuris`/`bonatti`/`both`/`all` |
| `BuildSearch(jd, planet, targetLon, lat, lon, hsys, direction)` | Longitude search ("forward" or "backward"); returns `*SearchResult` |
| `PrintText(r Result, verbose bool) error` | Render human-readable output to stdout |
| `PrintJSON(r Result, verbose bool) error` | Render JSON output to stdout |

### `almuten` package

| Symbol | Description |
|---|---|
| `BuildChart(jd, lat, lon, hsys) (Chart, error)` | Gather positions, node, houses, sect once |
| `LordOfGeniture(c, opts) (Scorecard, []int)` | Lilly's scorecard + tied winners |
| `AlmutenFiguris(c) (Scorecard, []int, error)` | Ibn Ezra's scorecard + tied winners |
| `AlmutenBonatti(c) (Scorecard, []int, error)` | Bonatti's Almudebit scorecard + tied winners |
| `PrenatalSyzygy(jd, lat, lon) (lon, kind, error)` | Most recent new/full moon before `jd` |
| `FindLongitudeDefault(jd, planet, targetLon, dir)` | Nearest time (per `Direction`) `planet` was/will be at `targetLon`, using recommended window/step |
| `HouseOf(lon, cusps)` | Quadrant house (1-12) containing a longitude |
| `DefaultOptions()` | Recommended variant defaults |

## Key Data Structures

### `swisseph` package

- `PlanetPos` — Longitude, Latitude, Distance, SpeedLon, SpeedLat, SpeedDistance
- `HouseResult` — Cusps[13], Ascendant, MC, ARMC, Vertex

### `output` package

- `Result` — JulianDay, HouseName, Lat, Lon, Planets, Ascendant, MC, ARMC, Vertex, Cusps, EphemerisWarning, Almuten, Search
- `PlanetEntry` — Name, Longitude, Sign, SignDegree, Speed, Latitude, Distance, SpeedLat, SpeedDistance
- `AngleEntry` — Longitude, Sign, SignDegree
- `CuspEntry` — House, Longitude, Sign, SignDegree
- `AlmutenEntry` — Method, Winners, Scoreboard (`[]AlmutenScore{Planet, Score}`, descending)
- `SearchResult` — Planet, TargetLon, TargetSign, TargetSignDeg, Direction, JulianDay, Year/Month/Day/Hour, SpeedLon, Retrograde, House

### `almuten` package

- `Chart` — JD, Lat, Lon, IsDiurnal, Positions, NorthNode, Houses, plus Figuris fields (PartFortune, Syzygy, SyzygyType, DayRuler, HourRuler)
- `Scorecard` — `map[int]int` keyed by planet ID
- `Options` — FreeOfSunBonus, PeregrineSuppressesTermFace, FixedStarOrb, LillyWater

## Output JSON Shape

```json
{
  "julian_day": 2460389.0,
  "planets": [{ "name": "Sun", "longitude": 0.0, "sign": "Aries", "sign_degree": 0.0, "speed": 1.0 }],
  "houses": {
    "system": "Placidus",
    "ascendant": { "longitude": 0.0, "sign": "Aries", "sign_degree": 0.0 },
    "mc": { "longitude": 0.0, "sign": "Aries", "sign_degree": 0.0 },
    "cusps": [{ "house": 1, "longitude": 0.0, "sign": "Aries", "sign_degree": 0.0 }]
  },
  "almuten": [
    { "method": "Lord of the Geniture", "winners": ["Mars"],
      "scoreboard": [{ "planet": "Mars", "score": 18 }] }
  ],
  "search": {
    "planet": "Mars", "target_longitude": 15.0, "target_sign": "Aries", "target_sign_degree": 15.0, "direction": "backward",
    "timestamp": "2021-09-14T03:22:00Z", "julian_day": 2459471.6, "speed": 0.45, "retrograde": false, "house": 4
  },
  "ephemeris_warning": "SwissEph file '...' not found; using Moshier eph."
}
```

`almuten` is present only when `--almuten` is passed (one entry per algorithm); its `scoreboard` is included only under `--verbose`. `search` is present only when `--search` is passed; its `speed`/`retrograde`/`house` fields are included only under `--verbose`. `ephemeris_warning` is only present under `--verbose` and only when the Swiss Ephemeris `.se1` files were not found (Moshier fallback is active). All are omitted entirely from normal output.

## Ephemeris Data

Binary `.se1` files in `ephe/` cover multiple epochs:
- `sepl_*.se1` — planets
- `semo_*.se1` — moon
- `seas_*.se1` — asteroids
- `sefstars.txt` — fixed-star catalogue (text; required for `FixStar` to resolve names like Regulus/Spica/Algol)

The path is resolved relative to the executable at runtime via `SetEphePath`. If `os.Executable()` fails, `cmd.Run` returns a hard error. If the `ephe/` directory is simply absent at the resolved path, the Swiss Ephemeris library silently falls back to the built-in Moshier approximation (lower precision, no external files required); `FixStar` instead returns an error, and the almuten code skips the fixed-star bonus rather than failing.

## License

Dual-licensed: AGPL v3 (open source) or commercial (Astrodienst AG). Choose commercial if distributing non-AGPL software.
