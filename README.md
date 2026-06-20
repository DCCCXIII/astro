# astro

Go bindings for the [Swiss Ephemeris](https://www.astro.com/swisseph/) C library, providing high-precision astronomical calculations for astrological software.

## Features

- Planetary position calculations (ecliptic longitude, latitude, distance, and daily speeds) for the seven traditional planets: Sun, Moon, Mercury, Venus, Mars, Jupiter, and Saturn
- House cusp calculations with support for multiple house systems (Placidus, Koch, Whole Sign, Regiomontanus, Equal, Campanus)
- Ascendant, Midheaven (MC), ARMC, and Vertex angles
- Traditional "chart victor" (almuten) calculations: Lilly's Lord of the Geniture (*Christian Astrology* p.115), Ibn Ezra's Almuten Figuris, and Bonatti's Almudebit (*Liber Astronomiae*)
- Fixed-star and rise/set helpers, and a pure-Go `dignities` package of essential-dignity tables (domicile, exaltation, triplicity, terms, faces)
- Zodiac sign conversion utility
- Thread-safe: all calls to the underlying C library are protected by a mutex

## Background

This tool's calculations implement specific historical astrological methods — Lilly's *Christian Astrology*, Ibn Ezra's Almuten Figuris, Bonatti's Almudebit, and the underlying essential-dignity tables (domicile, exaltation, triplicity, terms, faces) — rather than a single modern synthesis. See [docs/](docs/README.md) for the historical sources, scoring logic, and explicit notes on where this implementation makes a specific choice among several traditional variants.

## Prerequisites

- Go 1.25+
- A C compiler (GCC or Clang) -- required by cgo to compile the bundled Swiss Ephemeris sources

## Building

```bash
go build -o astro .
```

The Swiss Ephemeris C sources are bundled in the `swisseph/` directory and compiled automatically by cgo during `go build`. No external library installation is needed.

## Running

```
astro [--house-system <system>] [--almuten <mode>] [--json] [--verbose] <datetime> <lat> <lon>
```

**Arguments:**

| Argument | Description |
|---|---|
| `<datetime>` | UTC date/time in ISO 8601 format, e.g. `2024-03-20T12:00:00Z` |
| `<lat>` | Geographic latitude in decimal degrees (north = positive) |
| `<lon>` | Geographic longitude in decimal degrees (east = positive, west = negative) |

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--house-system` | `placidus` | House system: `placidus`, `koch`, `whole-sign`, `regiomontanus`, `equal`, `campanus` |
| `--almuten` | — | Chart victor: `geniture` (Lilly's Lord of the Geniture), `figuris` (Ibn Ezra's Almuten Figuris), `bonatti` (Bonatti's Almudebit), `both` (geniture+figuris), or `all`. Omit for no almuten output |
| `--json` | — | Output results as JSON instead of human-readable text |
| `--verbose` | — | Include ecliptic latitude, distance, speed components, ARMC, Vertex, the almuten scoreboard, and ephemeris source warning (if Moshier fallback is active) |

The binary looks for ephemeris data files (`.se1`) in an `ephe/` directory next to the executable. These files are included in the repository and provide high-precision planetary data.

### Examples

```bash
# Human-readable output
./astro 2024-03-20T12:00:00Z 40.7128 -74.0060

# JSON output
./astro --json 2024-03-20T12:00:00Z 40.7128 -74.0060

# Verbose output (adds latitude, distance, speed components, ARMC, Vertex)
./astro --verbose 2024-03-20T12:00:00Z 40.7128 -74.0060

# Verbose JSON
./astro --json --verbose 2024-03-20T12:00:00Z 40.7128 -74.0060

# Different house system
./astro --house-system koch 2024-03-20T12:00:00Z 40.7128 -74.0060

# Chart victor (almuten): victor only
./astro --almuten geniture 2024-03-20T12:00:00Z 40.7128 -74.0060

# Both Lilly/Ibn Ezra algorithms with the full scoreboard
./astro --almuten both --verbose 2024-03-20T12:00:00Z 40.7128 -74.0060

# Bonatti's Almudebit
./astro --almuten bonatti 2024-03-20T12:00:00Z 40.7128 -74.0060

# All three algorithms
./astro --almuten all --verbose 2024-03-20T12:00:00Z 40.7128 -74.0060
```

### Example output (human-readable)

```
Julian Day: 2460390.000000

=== Planetary Positions ===
Sun            0.3681°  (Aries  0.37°)  speed: +0.9933°/day
Moon         128.2881°  (Leo  8.29°)  speed: +12.0327°/day
Mercury       17.9740°  (Aries 17.97°)  speed: +1.4214°/day
Venus        340.6300°  (Pisces 10.63°)  speed: +1.2370°/day
Mars         328.0619°  (Aquarius 28.06°)  speed: +0.7779°/day
Jupiter       44.9602°  (Taurus 14.96°)  speed: +0.2013°/day
Saturn       342.2693°  (Pisces 12.27°)  speed: +0.1182°/day

=== Houses (placidus) for (40.7128°, -74.0060°) ===
Ascendant:    24.6432°  (Aries 24.64°)
MC:          283.3523°  (Capricorn 13.35°)

House cusps:
  House  1:   24.6432°  (Aries 24.64°)
  House  2:   58.8302°  (Taurus 28.83°)
  ...
```

### Example output (almuten, with `--almuten both --verbose`)

Without `--verbose`, only the `Victor:` line is shown per algorithm. Tied planets are all reported.

```
=== Almuten: Lord of the Geniture ===
Victor: Jupiter
  Jupiter     +21
  Mars        +12
  Venus       +11
  ...

=== Almuten: Almuten Figuris ===
Victor: Mars
  Mars        +27
  Mercury     +26
  Sun         +25
  ...
```

### Example output (JSON)

```json
{
  "julian_day": 2460390,
  "planets": [
    {"name": "Sun", "longitude": 0.368, "sign": "Aries", "sign_degree": 0.368, "speed": 0.993},
    ...
  ],
  "houses": {
    "system": "placidus",
    "ascendant": {"longitude": 24.643, "sign": "Aries", "sign_degree": 24.643},
    "mc": {"longitude": 283.352, "sign": "Capricorn", "sign_degree": 13.352},
    "cusps": [
      {"house": 1, "longitude": 24.643, "sign": "Aries", "sign_degree": 24.643},
      ...
    ]
  }
}
```

With `--almuten`, an `almuten` array is added (one entry per algorithm; `scoreboard` only under `--verbose`):

```json
{
  "almuten": [
    {
      "method": "Lord of the Geniture",
      "winners": ["Jupiter"],
      "scoreboard": [{"planet": "Jupiter", "score": 21}, ...]
    }
  ]
}
```

## Package API

The `swisseph` package exposes the low-level bindings:

### Functions

| Function | Description |
|---|---|
| `SetEphePath(path string)` | Set the path to `.se1` ephemeris data files |
| `Close()` | Free all library resources (call via `defer`) |
| `JulDay(year, month, day int, hour float64) float64` | Convert a calendar date (UTC) to a Julian Day number |
| `CalcPlanet(tjdUT float64, planet int) (PlanetPos, string, error)` | Calculate a planet's position at a given time; the string is a warning when Moshier fallback is active (empty otherwise) |
| `CalcHouses(tjdUT float64, geoLat, geoLon float64, hsys byte) (HouseResult, error)` | Calculate house cusps and angles for a time and location |
| `FixStar(name string, tjdUT float64) (PlanetPos, error)` | Position of a fixed star by catalogue name (resolved via `ephe/sefstars.txt`) |
| `RiseTrans(tjdUT float64, planet int, geoLat, geoLon, alt float64, rising bool) (float64, error)` | Next rise/set time as a Julian Day; returns `ErrNoRiseSet` at polar latitudes |
| `PlanetName(planet int) string` | Get the human-readable name for a planet ID |
| `ZodiacSign(longitude float64) (string, float64)` | Convert ecliptic longitude to zodiac sign and degree |

### Constants

**Planets:** `Sun`, `Moon`, `Mercury`, `Venus`, `Mars`, `Jupiter`, `Saturn`, `MeanNode`

**House systems:** `HousePlacidus`, `HouseKoch`, `HouseWholeSign`, `HouseRegiomontanus`, `HouseEqual`, `HouseCampanus`

### Higher-level packages

- **`dignities`** — pure-Go essential-dignity tables and lookups (`DomicileRuler`, `ExaltationRuler`, `ActiveTriplicityRuler`, `TermRuler`, `FaceRuler`, …). No cgo.
- **`almuten`** — chart-victor algorithms: `BuildChart` assembles a `Chart` once, then `LordOfGeniture`, `AlmutenFiguris`, and `AlmutenBonatti` each return a `Scorecard` plus the tied winners.

### Types

**`PlanetPos`** -- returned by `CalcPlanet`:
- `Longitude` -- ecliptic longitude in degrees (0-360)
- `Latitude` -- ecliptic latitude in degrees
- `Distance` -- distance from Earth in AU
- `SpeedLon`, `SpeedLat`, `SpeedDistance` -- daily speeds

**`HouseResult`** -- returned by `CalcHouses`:
- `Cusps[1..12]` -- house cusp longitudes in degrees
- `Ascendant`, `MC`, `ARMC`, `Vertex` -- key angles in degrees

## Usage example

```go
package main

import (
    "fmt"
    "github.com/dcccxiii/astro/swisseph"
)

func main() {
    swisseph.SetEphePath("./ephe")
    defer swisseph.Close()

    jd := swisseph.JulDay(2024, 3, 20, 12.0)

    pos, _, _ := swisseph.CalcPlanet(jd, swisseph.Sun)
    sign, deg := swisseph.ZodiacSign(pos.Longitude)
    fmt.Printf("Sun: %s %.2f°\n", sign, deg)

    houses, _ := swisseph.CalcHouses(jd, 40.7128, -74.0060, swisseph.HousePlacidus)
    ascSign, ascDeg := swisseph.ZodiacSign(houses.Ascendant)
    fmt.Printf("Ascendant: %s %.2f°\n", ascSign, ascDeg)
}
```

## License

See [LICENSE](LICENSE) for the Swiss Ephemeris licensing terms (AGPL or commercial).

## Development

This repository was developed with the assistance of [Claude Code](https://claude.ai/claude-code) (Anthropic's AI coding assistant).
