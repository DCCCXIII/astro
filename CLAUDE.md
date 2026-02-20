# CLAUDE.md

## Project Overview

Go wrapper around the Swiss Ephemeris C library for astrological calculations, paired with a unix-like CLI to compute planetary positions and house cusps for a given time and geographic location.

## Repository Structure

```
astro/
├── main.go              # CLI entry point
├── swisseph/
│   ├── swisseph.go      # Go cgo bindings to Swiss Ephemeris
│   ├── *.c / *.h        # Bundled Swiss Ephemeris C source (no external install needed)
├── ephe/                # Binary ephemeris data files (~105 MB, .se1 format)
├── go.mod               # module github.com/dcccxiii/astro, go 1.25
└── README.md
```

## Build

```bash
go build -o astro .
```

Requires Go 1.25+ and a C compiler (GCC or Clang). No external C library installation needed — Swiss Ephemeris C sources are bundled. cgo compiles them automatically via directives in `swisseph/swisseph.go`.

## CLI Usage

```bash
astro [--house-system <system>] [--json] <datetime> <lat> <lon>
```

- `<datetime>`: UTC time in ISO 8601 (e.g. `2024-03-20T12:00:00Z`)
- `<lat>`: Decimal degrees, north positive
- `<lon>`: Decimal degrees, east positive
- `--house-system`: `placidus` (default), `koch`, `whole-sign`, `regiomontanus`, `equal`, `campanus`
- `--json`: Output JSON instead of human-readable text

## Key Go API (`swisseph` package)

| Function | Description |
|---|---|
| `SetEphePath(path)` | Set path to `ephe/` directory |
| `Close()` | Free C library resources |
| `JulDay(year, month, day, hour)` | Calendar date → Julian Day |
| `CalcPlanet(tjdUT, planet)` | Planet position at Julian Day |
| `CalcHouses(tjdUT, lat, lon, hsys)` | House cusps for location/time |
| `ZodiacSign(longitude)` | Ecliptic longitude → sign name + degree |

All C calls are mutex-protected for thread safety.

**Planet IDs:** `swisseph.Sun`, `Moon`, `Mercury`, `Venus`, `Mars`, `Jupiter`, `Saturn`

**House system bytes:** `HousePlacidus='P'`, `HouseKoch='K'`, `HouseWholeSign='W'`, `HouseRegiomontanus='R'`, `HouseEqual='A'`, `HouseCampanus='C'`

## Key Data Structures

- `PlanetPos` — Longitude, Latitude, Distance, SpeedLon, SpeedLat, SpeedDistance
- `HouseResult` — Cusps[13], Ascendant, MC, ARMC, Vertex

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

The path is resolved relative to the executable at runtime via `SetEphePath`.

## License

Dual-licensed: AGPL v3 (open source) or commercial (Astrodienst AG). Choose commercial if distributing non-AGPL software.
