# CLAUDE.md

## Project Overview

Go wrapper around the Swiss Ephemeris C library for astrological calculations, paired with a unix-like CLI to compute planetary positions and house cusps for a given time and geographic location.

## Repository Structure

```
astro/
├── main.go              # Minimal entry point — delegates to cmd.Run
├── cmd/
│   ├── run.go           # CLI flag parsing, validation, orchestration
│   └── run_test.go      # Tests for flag parsing and house system lookup
├── output/
│   ├── result.go        # Result type + Build() — all swisseph calls live here
│   ├── text.go          # PrintText() — human-readable renderer
│   └── json.go          # PrintJSON() — JSON renderer
├── swisseph/
│   ├── swisseph.go      # Go cgo bindings to Swiss Ephemeris
│   ├── swisseph_test.go # Tests for the swisseph package
│   ├── *.c / *.h        # Bundled Swiss Ephemeris C source (no external install needed)
├── ephe/                # Binary ephemeris data files (~105 MB, .se1 format)
├── go.mod               # module github.com/dcccxiii/astro, go 1.25
└── README.md
```

## Commands

ALWAYS use these make targets instead of raw Go commands:

- **Build:** `make` (runs `go fmt`, `go vet`, `go test`, then `go build`)
- **Test:** `make test` (runs `go fmt`, `go vet`, then `go test -v ./...`)
- **Format only:** `make fmt`
- **Vet only:** `make vet`

Never run `go build`, `go test`, `go fmt`, or `go vet` directly.

## Build

```bash
make
```

Requires Go 1.25+ and a C compiler (GCC or Clang). No external C library installation needed — Swiss Ephemeris C sources are bundled. cgo compiles them automatically via directives in `swisseph/swisseph.go`.

## CLI Usage

```bash
astro [--house-system <system>] [--json] [--verbose] <datetime> <lat> <lon>
```

- `<datetime>`: UTC time in ISO 8601 (e.g. `2024-03-20T12:00:00Z`)
- `<lat>`: Decimal degrees, north positive
- `<lon>`: Decimal degrees, east positive
- `--house-system`: `placidus` (default), `koch`, `whole-sign`, `regiomontanus`, `equal`, `campanus`
- `--json`: Output JSON instead of human-readable text
- `--verbose`: Include ecliptic latitude, distance, speed components, ARMC, and Vertex

## Package Overview

### `cmd`

`Run(args []string) error` is the real entry point. It parses flags with `flag.NewFlagSet`, validates arguments, resolves the ephemeris path relative to the executable, and delegates to the `output` package.

### `output`

Three files with a clean separation of concerns:

- **`result.go`** — `Build()` calls `swisseph.CalcPlanet` and `swisseph.CalcHouses`, assembles a `Result` struct. Neither renderer touches the C library.
- **`text.go`** — `PrintText(r Result, verbose bool) error` writes human-readable output to stdout.
- **`json.go`** — `PrintJSON(r Result, verbose bool) error` marshals to indented JSON and writes to stdout.

### `swisseph`

Low-level cgo bindings. All C calls are mutex-protected for thread safety. Callers never interact with C types directly.

## Key Go API

### `swisseph` package

| Function | Description |
|---|---|
| `SetEphePath(path)` | Set path to `ephe/` directory |
| `Close()` | Free C library resources |
| `JulDay(year, month, day, hour)` | Calendar date → Julian Day |
| `CalcPlanet(tjdUT, planet)` | Planet position at Julian Day |
| `CalcHouses(tjdUT, lat, lon, hsys)` | House cusps for location/time |
| `ZodiacSign(longitude)` | Ecliptic longitude → sign name + degree (normalises to [0, 360) automatically) |

**Planet IDs:** `swisseph.Sun`, `Moon`, `Mercury`, `Venus`, `Mars`, `Jupiter`, `Saturn`

**House system bytes:** `HousePlacidus='P'`, `HouseKoch='K'`, `HouseWholeSign='W'`, `HouseRegiomontanus='R'`, `HouseEqual='A'`, `HouseCampanus='C'`

### `output` package

| Function | Description |
|---|---|
| `Build(jd, planets, lat, lon, hsys, hsysName)` | Compute full chart; returns `Result` or error |
| `PrintText(r Result, verbose bool) error` | Render human-readable output to stdout |
| `PrintJSON(r Result, verbose bool) error` | Render JSON output to stdout |

## Key Data Structures

### `swisseph` package

- `PlanetPos` — Longitude, Latitude, Distance, SpeedLon, SpeedLat, SpeedDistance
- `HouseResult` — Cusps[13], Ascendant, MC, ARMC, Vertex

### `output` package

- `Result` — JulianDay, HouseName, Lat, Lon, Planets, Ascendant, MC, ARMC, Vertex, Cusps
- `PlanetEntry` — Name, Longitude, Sign, SignDegree, Speed, Latitude, Distance, SpeedLat, SpeedDistance
- `AngleEntry` — Longitude, Sign, SignDegree
- `CuspEntry` — House, Longitude, Sign, SignDegree

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
  }
}
```

## Ephemeris Data

Binary `.se1` files in `ephe/` cover multiple epochs:
- `sepl_*.se1` — planets
- `semo_*.se1` — moon
- `seas_*.se1` — stars

The path is resolved relative to the executable at runtime via `SetEphePath`. If `os.Executable()` fails, `cmd.Run` returns a hard error. If the `ephe/` directory is simply absent at the resolved path, the Swiss Ephemeris library silently falls back to the built-in Moshier approximation (lower precision, no external files required).

## License

Dual-licensed: AGPL v3 (open source) or commercial (Astrodienst AG). Choose commercial if distributing non-AGPL software.
