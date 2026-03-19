-- ============================================================
-- XANDARIS — Star Service Seed Data
-- ============================================================
-- Units reminder:
--   mass, luminosity, radius: solar × 1000
--   temp: Kelvin
--   luminosity_class modifiers: × 1000 multiplier (1000 = 1.0x)
-- ============================================================

-- ----------------------------------------------------------
-- LUMINOSITY CLASSES
-- ----------------------------------------------------------
-- Modifiers scale the spectral_class min/max ranges.
-- mass_modifier and radius_modifier are × 1000 (1000 = 1.0x).
-- Spectral class ranges assume main sequence (V) as baseline.

INSERT INTO luminosity_class (id, numeral, name, mass_modifier, radius_modifier) VALUES
    (1, 'Ia',  'bright_supergiant', 2000,  100000),  -- 2.0x mass, 100x radius
    (2, 'Ib',  'supergiant',        1500,  50000),    -- 1.5x mass, 50x radius
    (3, 'II',  'bright_giant',       1200,  15000),    -- 1.2x mass, 15x radius
    (4, 'III', 'giant',              1000,  5000),     -- 1.0x mass, 5x radius
    (5, 'IV',  'subgiant',           1000,  1500),     -- 1.0x mass, 1.5x radius
    (6, 'V',   'main_sequence',      1000,  1000),     -- baseline
    (7, 'VI',  'subdwarf',           800,   800),      -- 0.8x mass, 0.8x radius
    (8, 'VII', 'white_dwarf',        600,   15);       -- 0.6x mass, 0.015x radius

-- ----------------------------------------------------------
-- LIFECYCLE STAGES
-- ----------------------------------------------------------
-- mod_luminosity / mod_temp are additive hints (solar×1000 / Kelvin).
-- Actual lifecycle transitions use UpdateStarLifecycle which sets
-- absolute values directly. These mods are for quick approximation
-- or procedural generation tweaks.

INSERT INTO lifecycle_stage (id, name, is_active, mod_luminosity, mod_temp) VALUES
    (1, 'protostar',        true,   0,      -2000),  -- cooler, embedded
    (2, 'main_sequence',    true,   0,      0),       -- baseline
    (3, 'subgiant',         true,   500,    -300),    -- slightly brighter, slightly cooler
    (4, 'red_giant',        true,   5000,   -2000),   -- much brighter, much cooler
    (5, 'horizontal_branch',true,   2000,   1000),    -- helium burning, hotter than RGB
    (6, 'asymptotic_giant', true,   10000,  -3000),   -- very bright, very cool
    (7, 'white_dwarf',      false,  -800,   10000),   -- dim but very hot surface
    (8, 'neutron_star',     false,  -950,   50000),   -- tiny, extremely hot
    (9, 'black_hole',       false,  -1000,  0);       -- no luminosity

-- ----------------------------------------------------------
-- SPECTRAL CLASSES (70 rows: OBAFGKM × subtypes 0-9)
-- ----------------------------------------------------------
-- Generated via interpolation within each letter class.
-- Subtype 0 = hottest/most massive end, subtype 9 = coolest/least massive.
-- All ranges are for main sequence (luminosity class V).

INSERT INTO spectral_class (id, letter, subtype, name,
    min_temp, max_temp, min_mass, max_mass,
    min_luminosity, max_luminosity, min_radius, max_radius)
SELECT
    ((ord - 1) * 10 + n + 1)::bigint AS id,
    letter,
    n::smallint AS subtype,
    letter || n::text AS name,
    -- Each subtype is a band: subtype 0 gets the top 10%, subtype 9 gets the bottom 10%
    (lo_t + ((hi_t - lo_t)::bigint * (9 - n) / 10))::integer AS min_temp,
    (lo_t + ((hi_t - lo_t)::bigint * (10 - n) / 10))::integer AS max_temp,
    (lo_m + ((hi_m - lo_m)::bigint * (9 - n) / 10))::integer AS min_mass,
    (lo_m + ((hi_m - lo_m)::bigint * (10 - n) / 10))::integer AS max_mass,
    (lo_l + ((hi_l - lo_l)::bigint * (9 - n) / 10))::integer AS min_luminosity,
    (lo_l + ((hi_l - lo_l)::bigint * (10 - n) / 10))::integer AS max_luminosity,
    (lo_r + ((hi_r - lo_r)::bigint * (9 - n) / 10))::integer AS min_radius,
    (lo_r + ((hi_r - lo_r)::bigint * (10 - n) / 10))::integer AS max_radius
FROM (
    VALUES
        -- letter, ord, hi_temp, lo_temp, hi_mass, lo_mass, hi_lum, lo_lum, hi_rad, lo_rad
        --                (K)     (K)    (sol×1k) (sol×1k) (sol×1k) (sol×1k) (sol×1k) (sol×1k)
        ('O'::char(1), 1, 50000, 28000, 150000, 16000,  500000000, 30000000, 15000, 6600),
        ('B'::char(1), 2, 28000, 10000,  16000,  2100,   30000000,    25000,  6600, 1800),
        ('A'::char(1), 3, 10000,  7500,   2100,  1400,      25000,     5000,  1800, 1400),
        ('F'::char(1), 4,  7500,  6000,   1400,  1040,       5000,     1500,  1400, 1150),
        ('G'::char(1), 5,  6000,  5200,   1040,   800,       1500,      600,  1150,  960),
        ('K'::char(1), 6,  5200,  3700,    800,   450,        600,       80,   960,  700),
        ('M'::char(1), 7,  3700,  2400,    450,    80,         80,        1,   700,  100)
) AS classes(letter, ord, hi_t, lo_t, hi_m, lo_m, hi_l, lo_l, hi_r, lo_r)
CROSS JOIN generate_series(0, 9) AS s(n)
ORDER BY ord, n;
