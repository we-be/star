# star

Star microservice for [Xandaris](https://github.com/hunterjsb/xandaris) — procedurally generates and serves star systems with PostGIS spatial indexing.

## Quick start

```bash
brew install postgresql@17 postgis sqlc
brew services start postgresql@17
psql postgres -c "CREATE DATABASE star;"
psql star -f db/migrations/001_init.sql
psql star -f db/migrations/002_seed.sql
sqlc generate
go build -o star-bin .
./star-bin -addr :8081
```

Open http://localhost:8081 for the 3D galaxy explorer.

### Generate a galaxy

```bash
curl -X POST http://localhost:8081/generate/universe \
  -H 'Content-Type: application/json' \
  -d '{"center_x":0,"center_y":0,"center_z":0,"radius":10000,"num_systems":500,"seed":1}'
```

## Architecture

```
star/
├── main.go                  Entry point, DB pool, graceful shutdown
├── sqlc.yaml                sqlc config (pgx/v5, PostGIS geometry → []byte)
├── db/
│   ├── migrations/
│   │   ├── 001_init.sql     Schema: spectral_class, luminosity_class, lifecycle_stage,
│   │   │                    star, system, system_star (PostGIS POINTZ generated columns)
│   │   └── 002_seed.sql     70 spectral subtypes, 8 luminosity classes, 9 lifecycle stages
│   └── queries/
│       └── stars.sql        20 queries including spatial (KNN, radius) + GetStellarEnvironment
├── internal/
│   ├── gen/
│   │   ├── physics.go       Mass-luminosity, habitable zone, frost line derivations
│   │   ├── star.go          Star generation (weighted IMF, property derivation)
│   │   ├── system.go        System generation (binary probability, Keplerian orbits)
│   │   ├── universe.go      Galaxy seeder (disk density, spiral arms)
│   │   ├── names.go         Procedural Latin/Greek-inspired name generator
│   │   └── id.go            Snowflake-ish ID generator (JS-safe, < 2^53)
│   ├── service/             Orchestrates gen + DB persistence
│   ├── api/                 net/http handlers with CORS
│   └── db/                  sqlc generated (models, queries)
└── static/
    └── index.html           Three.js 3D galaxy explorer with planet service integration
```

## API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/stars/{id}` | Star with full classification |
| POST | `/stars` | Create star |
| PUT | `/stars/{id}/lifecycle` | Update lifecycle + derived zones |
| GET | `/stars/nearby?x=&y=&z=&radius=` | Stars within 3D radius (GiST) |
| GET | `/stars/nearest?x=&y=&z=&limit=` | K-nearest stars (KNN) |
| GET | `/systems/{id}` | System by ID |
| POST | `/systems` | Create system |
| GET | `/systems/{id}/stars` | Stars in system with orbital params |
| GET | `/systems/{id}/environment` | Stellar environment (planet service contract) |
| GET | `/systems/nearby` | Systems within radius |
| GET | `/systems/nearest` | K-nearest systems |
| POST | `/generate/system` | Generate single system |
| POST | `/generate/universe` | Generate galaxy region |
| GET | `/spectral-classes` | List all spectral types |
| GET | `/luminosity-classes` | List luminosity classes |
| GET | `/lifecycle-stages` | List lifecycle stages |

## Web client controls

| Key | Action |
|-----|--------|
| WASD | Hop between neighboring stars |
| Tab | Cycle planets in system view |
| / | Search systems by name |
| R | Jump to random unexplored system |
| L | Toggle discovery log |
| G | Galaxy overview stats |
| 1-7 | Filter by spectral type (OBAFGKM) |
| 0 | Clear filter |
| Esc | Back one level (planet → system → galaxy) |

## Design decisions

- **Integer coordinates + PostGIS generated column**: `pos_x/y/z` (INTEGER) are the source of truth. A `GENERATED ALWAYS AS` geometry column provides GiST spatial indexing. Go code never touches geometry types.
- **Barycenter-based systems**: `system` is the gravitational center. Stars orbit it via `system_star`. Single-star systems have zeroed orbital params. Binaries have real Keplerian elements.
- **Gameplay-weighted IMF**: Real stellar distribution is 76% M-dwarfs. Game distribution is 35% M, 20% G, 5% A, 2% O — more interesting systems to explore.

## Related

- [we-be/planets](https://github.com/we-be/planets) — Planet microservice (queries this service's `/systems/{id}/environment`)
