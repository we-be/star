-- ============================================================
-- XANDARIS — Star Service Queries
-- ============================================================
-- Convention: geometry column (pos) is NEVER selected.
-- All reads use pos_x, pos_y, pos_z integer columns.
-- All spatial WHERE clauses reference pos for index acceleration.
-- All writes set pos_x, pos_y, pos_z; the generated column handles the rest.
-- ============================================================

-- ===================
-- STARS
-- ===================

-- name: GetStar :one
SELECT s.id, s.dt_created, s.dt_updated,
       s.name,
       s.id_spectral_class, s.id_luminosity_class, s.id_lifecycle_stage,
       s.mass, s.radius, s.temp, s.luminosity, s.metallicity, s.age,
       s.flare_frequency, s.solar_wind, s.variability,
       s.habitable_inner, s.habitable_outer, s.frost_line,
       s.pos_x, s.pos_y, s.pos_z,
       sc.letter, sc.subtype, sc.name AS spectral_name,
       lc.numeral AS luminosity_numeral, lc.name AS luminosity_name,
       ls.name AS lifecycle_name, ls.is_active
FROM star s
JOIN spectral_class sc ON sc.id = s.id_spectral_class
JOIN luminosity_class lc ON lc.id = s.id_luminosity_class
JOIN lifecycle_stage ls ON ls.id = s.id_lifecycle_stage
WHERE s.id = $1;

-- name: InsertStar :one
INSERT INTO star (
    id, name,
    id_spectral_class, id_luminosity_class, id_lifecycle_stage,
    mass, radius, temp, luminosity, metallicity, age,
    flare_frequency, solar_wind, variability,
    habitable_inner, habitable_outer, frost_line,
    pos_x, pos_y, pos_z
) VALUES (
    $1, $2,
    $3, $4, $5,
    $6, $7, $8, $9, $10, $11,
    $12, $13, $14,
    $15, $16, $17,
    $18, $19, $20
)
RETURNING id, name, mass, radius, temp, luminosity, pos_x, pos_y, pos_z;

-- name: UpdateStarLifecycle :exec
UPDATE star SET
    id_lifecycle_stage = $2,
    luminosity = $3,
    temp = $4,
    habitable_inner = $5,
    habitable_outer = $6,
    frost_line = $7,
    dt_updated = now()
WHERE id = $1;

-- ===================
-- SPATIAL QUERIES
-- ===================

-- name: FindStarsWithinRadius :many
-- Uses the generated pos column + GiST index for spatial filtering,
-- but returns integer coordinates only.
SELECT s.id, s.name, s.temp, s.luminosity, s.mass,
       s.pos_x, s.pos_y, s.pos_z,
       ST_3DDistance(s.pos, ST_MakePoint($1::float8, $2::float8, $3::float8))::integer AS dist
FROM star s
WHERE ST_3DDWithin(s.pos, ST_MakePoint($1::float8, $2::float8, $3::float8), $4::float8)
ORDER BY dist;

-- name: FindNearestStars :many
SELECT s.id, s.name, s.temp, s.luminosity, s.mass,
       s.pos_x, s.pos_y, s.pos_z,
       (s.pos <-> ST_MakePoint($1::float8, $2::float8, $3::float8))::integer AS dist
FROM star s
ORDER BY s.pos <-> ST_MakePoint($1::float8, $2::float8, $3::float8)
LIMIT $4;

-- name: FindSystemsWithinRadius :many
SELECT sys.id, sys.name,
       sys.pos_x, sys.pos_y, sys.pos_z,
       ST_3DDistance(sys.pos, ST_MakePoint($1::float8, $2::float8, $3::float8))::integer AS dist
FROM system sys
WHERE ST_3DDWithin(sys.pos, ST_MakePoint($1::float8, $2::float8, $3::float8), $4::float8)
ORDER BY dist;

-- name: FindNearestSystems :many
SELECT sys.id, sys.name,
       sys.pos_x, sys.pos_y, sys.pos_z,
       (sys.pos <-> ST_MakePoint($1::float8, $2::float8, $3::float8))::integer AS dist
FROM system sys
ORDER BY sys.pos <-> ST_MakePoint($1::float8, $2::float8, $3::float8)
LIMIT $4;

-- ===================
-- SYSTEMS
-- ===================

-- name: GetSystem :one
SELECT id, dt_created, dt_updated, name, pos_x, pos_y, pos_z
FROM system WHERE id = $1;

-- name: InsertSystem :one
INSERT INTO system (id, name, pos_x, pos_y, pos_z)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, name, pos_x, pos_y, pos_z;

-- name: GetSystemStars :many
SELECT ss.id_system, ss.id_star, ss.semi_major_axis, ss.eccentricity,
       ss.inclination, ss.orbital_period, ss.is_primary,
       s.name AS star_name, s.mass, s.radius, s.temp, s.luminosity,
       s.habitable_inner, s.habitable_outer, s.frost_line,
       s.flare_frequency, s.solar_wind,
       s.pos_x, s.pos_y, s.pos_z,
       sc.name AS spectral_name,
       lc.numeral AS luminosity_numeral,
       ls.name AS lifecycle_name
FROM system_star ss
JOIN star s ON s.id = ss.id_star
JOIN spectral_class sc ON sc.id = s.id_spectral_class
JOIN luminosity_class lc ON lc.id = s.id_luminosity_class
JOIN lifecycle_stage ls ON ls.id = s.id_lifecycle_stage
WHERE ss.id_system = $1
ORDER BY ss.is_primary DESC, s.mass DESC;

-- name: InsertSystemStar :exec
INSERT INTO system_star (
    id_system, id_star, semi_major_axis, eccentricity,
    inclination, orbital_period, is_primary
) VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: CountSystemStars :one
SELECT COUNT(*) FROM system_star WHERE id_system = $1;

-- ===================
-- STELLAR ENVIRONMENT
-- (what the planet service calls)
-- ===================

-- name: GetStellarEnvironment :one
SELECT
    sys.id AS system_id,
    sys.name AS system_name,
    sys.pos_x, sys.pos_y, sys.pos_z,
    s.id AS primary_star_id,
    s.temp, s.luminosity, s.mass AS star_mass,
    s.metallicity,
    s.habitable_inner, s.habitable_outer, s.frost_line,
    s.flare_frequency, s.solar_wind,
    sc.name AS spectral_name,
    ls.name AS lifecycle_name, ls.is_active
FROM system sys
JOIN system_star ss ON ss.id_system = sys.id AND ss.is_primary = TRUE
JOIN star s ON s.id = ss.id_star
JOIN spectral_class sc ON sc.id = s.id_spectral_class
JOIN lifecycle_stage ls ON ls.id = s.id_lifecycle_stage
WHERE sys.id = $1;

-- ===================
-- LOOKUP TABLES
-- ===================

-- name: ListSpectralClasses :many
SELECT * FROM spectral_class ORDER BY letter, subtype;

-- name: ListLuminosityClasses :many
SELECT * FROM luminosity_class ORDER BY id;

-- name: ListLifecycleStages :many
SELECT * FROM lifecycle_stage ORDER BY id;

-- name: GetSpectralClass :one
SELECT * FROM spectral_class WHERE id = $1;

-- name: GetSpectralClassByName :one
SELECT * FROM spectral_class WHERE letter = $1 AND subtype = $2;
