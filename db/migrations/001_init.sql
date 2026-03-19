-- ============================================================
-- XANDARIS — Star Microservice Schema
-- ============================================================
-- Coordinate convention:
--   - All positions are absolute 3D coordinates stored as INTEGER.
--   - 1 coordinate unit = game-defined scale (e.g., 1 unit = 0.01 ly).
--   - The geometry(POINTZ, 0) column exists ONLY for PostGIS spatial
--     indexing. Go code never touches it directly.
--   - Integer columns (pos_x, pos_y, pos_z) are the source of truth.
--   - A generated column keeps the geometry in sync automatically.
--
-- Other conventions:
--   - dt_created / dt_updated on every table
--   - Integer math: temperatures in Kelvin, masses in solar-masses × 1000,
--     distances in AU × 1000 (milliau), luminosity in solar × 1000
-- ============================================================

CREATE EXTENSION IF NOT EXISTS postgis;

-- ----------------------------------------------------------
-- LOOKUP / REFERENCE TABLES
-- ----------------------------------------------------------

CREATE TABLE spectral_class (
    id              BIGINT PRIMARY KEY,
    dt_created      TIMESTAMPTZ NOT NULL DEFAULT now(),
    dt_updated      TIMESTAMPTZ NOT NULL DEFAULT now(),

    letter          CHAR(1)      NOT NULL,
    subtype         SMALLINT     NOT NULL,
    name            VARCHAR(8)   NOT NULL,

    min_temp        INTEGER      NOT NULL,
    max_temp        INTEGER      NOT NULL,
    min_mass        INTEGER      NOT NULL,
    max_mass        INTEGER      NOT NULL,
    min_luminosity  INTEGER      NOT NULL,
    max_luminosity  INTEGER      NOT NULL,
    min_radius      INTEGER      NOT NULL,
    max_radius      INTEGER      NOT NULL,

    UNIQUE (letter, subtype)
);

CREATE TABLE luminosity_class (
    id              BIGINT PRIMARY KEY,
    dt_created      TIMESTAMPTZ NOT NULL DEFAULT now(),
    dt_updated      TIMESTAMPTZ NOT NULL DEFAULT now(),

    numeral         VARCHAR(4)   NOT NULL UNIQUE,
    name            VARCHAR(32)  NOT NULL,
    mass_modifier   INTEGER      NOT NULL DEFAULT 1000,
    radius_modifier INTEGER      NOT NULL DEFAULT 1000
);

CREATE TABLE lifecycle_stage (
    id              BIGINT PRIMARY KEY,
    dt_created      TIMESTAMPTZ NOT NULL DEFAULT now(),
    dt_updated      TIMESTAMPTZ NOT NULL DEFAULT now(),

    name            VARCHAR(32)  NOT NULL UNIQUE,
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    mod_luminosity  INTEGER      NOT NULL DEFAULT 0,
    mod_temp        INTEGER      NOT NULL DEFAULT 0
);

-- ----------------------------------------------------------
-- CORE: SYSTEM (barycenter)
-- ----------------------------------------------------------

CREATE TABLE system (
    id              BIGINT PRIMARY KEY,
    dt_created      TIMESTAMPTZ NOT NULL DEFAULT now(),
    dt_updated      TIMESTAMPTZ NOT NULL DEFAULT now(),

    name            VARCHAR(128) NOT NULL,

    -- integer coordinates (source of truth)
    pos_x           INTEGER      NOT NULL,
    pos_y           INTEGER      NOT NULL,
    pos_z           INTEGER      NOT NULL,

    -- PostGIS geometry (generated, for spatial index only)
    pos             GEOMETRY(POINTZ, 0) GENERATED ALWAYS AS (
                        ST_MakePoint(pos_x::float8, pos_y::float8, pos_z::float8)
                    ) STORED
);

-- ----------------------------------------------------------
-- CORE: STAR
-- ----------------------------------------------------------

CREATE TABLE star (
    id              BIGINT PRIMARY KEY,
    dt_created      TIMESTAMPTZ NOT NULL DEFAULT now(),
    dt_updated      TIMESTAMPTZ NOT NULL DEFAULT now(),

    name            VARCHAR(128) NOT NULL,

    -- classification
    id_spectral_class   BIGINT   NOT NULL REFERENCES spectral_class(id),
    id_luminosity_class BIGINT   NOT NULL REFERENCES luminosity_class(id),
    id_lifecycle_stage  BIGINT   NOT NULL REFERENCES lifecycle_stage(id),

    -- physical properties
    mass            INTEGER      NOT NULL,
    radius          INTEGER      NOT NULL,
    temp            INTEGER      NOT NULL,
    luminosity      INTEGER      NOT NULL,
    metallicity     INTEGER      NOT NULL DEFAULT 0,
    age             INTEGER      NOT NULL,

    -- tier 2: activity
    flare_frequency SMALLINT     NOT NULL DEFAULT 0,
    solar_wind      SMALLINT     NOT NULL DEFAULT 100,
    variability     SMALLINT     NOT NULL DEFAULT 0,

    -- derived zones (milliau)
    habitable_inner INTEGER      NOT NULL,
    habitable_outer INTEGER      NOT NULL,
    frost_line      INTEGER      NOT NULL,

    -- integer coordinates (source of truth)
    pos_x           INTEGER      NOT NULL,
    pos_y           INTEGER      NOT NULL,
    pos_z           INTEGER      NOT NULL,

    -- PostGIS geometry (generated, for spatial index only)
    pos             GEOMETRY(POINTZ, 0) GENERATED ALWAYS AS (
                        ST_MakePoint(pos_x::float8, pos_y::float8, pos_z::float8)
                    ) STORED
);

-- ----------------------------------------------------------
-- BINARY / MULTIPLE STAR SYSTEMS
-- ----------------------------------------------------------

CREATE TABLE system_star (
    id_system       BIGINT       NOT NULL REFERENCES system(id) ON DELETE CASCADE,
    id_star         BIGINT       NOT NULL REFERENCES star(id),
    dt_created      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    dt_updated      TIMESTAMPTZ  NOT NULL DEFAULT now(),

    semi_major_axis INTEGER      NOT NULL DEFAULT 0,
    eccentricity    INTEGER      NOT NULL DEFAULT 0,
    inclination     INTEGER      NOT NULL DEFAULT 0,
    orbital_period  INTEGER      NOT NULL DEFAULT 0,

    is_primary      BOOLEAN      NOT NULL DEFAULT TRUE,

    PRIMARY KEY (id_system, id_star)
);

-- ----------------------------------------------------------
-- INDEXES
-- ----------------------------------------------------------

-- 3D spatial indexes (N-dimensional GiST)
CREATE INDEX idx_star_pos   ON star   USING GIST (pos gist_geometry_ops_nd);
CREATE INDEX idx_system_pos ON system USING GIST (pos gist_geometry_ops_nd);

-- classification lookups
CREATE INDEX idx_star_spectral   ON star (id_spectral_class);
CREATE INDEX idx_star_luminosity ON star (id_luminosity_class);
CREATE INDEX idx_star_lifecycle  ON star (id_lifecycle_stage);

-- system membership
CREATE INDEX idx_system_star_star   ON system_star (id_star);
CREATE INDEX idx_system_star_system ON system_star (id_system);

-- ----------------------------------------------------------
-- DERIVATION NOTES
-- ----------------------------------------------------------
--
-- HABITABLE ZONE:
--   inner = sqrt(luminosity / 1000) * 950   (milliau)
--   outer = sqrt(luminosity / 1000) * 1370  (milliau)
--
-- FROST LINE:
--   frost_line = sqrt(luminosity / 1000) * 2700  (milliau)
--
-- BINARY SYSTEMS:
--   separation >> hz → S-type (planets orbit one star)
--   separation << hz → P-type (planets orbit barycenter)
--
-- ============================================================
