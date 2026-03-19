package gen

import (
	"fmt"
	"math"
	"math/rand/v2"

	"github.com/we-be/star/internal/db"
)

// StarConfig holds cached reference data needed for generation.
type StarConfig struct {
	SpectralClasses   []db.SpectralClass
	LuminosityClasses []db.LuminosityClass
	LifecycleStages   []db.LifecycleStage
}

// StarResult holds a generated star's properties before persistence.
type StarResult struct {
	ID                int64
	Name              string
	SpectralClassID   int64
	LuminosityClassID int64
	LifecycleStageID  int64
	Mass              int32
	Radius            int32
	Temp              int32
	Luminosity        int32
	Metallicity       int32
	Age               int32
	FlareFrequency    int16
	SolarWind         int16
	Variability       int16
	HabitableInner    int32
	HabitableOuter    int32
	FrostLine         int32
	PosX, PosY, PosZ  int32
}

// Gameplay-weighted spectral class distribution.
// Real IMF is ~76% M, ~0.00003% O — boring for a game.
var spectralWeight = map[byte]int{
	'O': 2, 'B': 3, 'A': 5, 'F': 10,
	'G': 20, 'K': 25, 'M': 35,
}

// Deterministic letter iteration order.
var letterOrder = []byte{'O', 'B', 'A', 'F', 'G', 'K', 'M'}

// GenerateStar creates a star with properties derived from
// spectral class ranges, mass-luminosity relation, and activity models.
func GenerateStar(rng *rand.Rand, cfg *StarConfig, x, y, z int32) StarResult {
	sc := pickSpectralClass(rng, cfg.SpectralClasses)
	lc := pickLuminosityClass(rng, cfg.LuminosityClasses)
	ls := findStage(cfg.LifecycleStages, "main_sequence")

	// Roll mass within spectral class range, scaled by luminosity class
	baseMass := randBetween(rng, sc.MinMass, sc.MaxMass)
	mass := int32(clamp64(int64(baseMass)*int64(lc.MassModifier)/1000, 1, math.MaxInt32))

	// Derive from mass
	luminosity := MassToLuminosity(mass)
	baseRadius := MassToRadius(mass)
	radius := int32(clamp64(int64(baseRadius)*int64(lc.RadiusModifier)/1000, 1, math.MaxInt32))
	temp := randBetween(rng, sc.MinTemp, sc.MaxTemp)

	// Derived zones
	hzInner, hzOuter := HabitableZone(luminosity)
	frostLine := FrostLine(luminosity)

	// Activity
	flareFreq := flareFrequency(rng, sc.Letter[0])
	solarWind := solarWindStrength(rng, mass)

	// Metallicity: [Fe/H] × 1000, roughly -0.5 to +0.5 dex
	metallicity := int32(rng.IntN(1001) - 500)

	// Age: massive stars burn fast
	age := starAge(rng, mass)

	return StarResult{
		ID:                NextID(),
		Name:              fmt.Sprintf("%s %s-%d", sc.Name, string([]byte{byte('A' + rng.IntN(26))}), rng.IntN(900)+100),
		SpectralClassID:   sc.ID,
		LuminosityClassID: lc.ID,
		LifecycleStageID:  ls.ID,
		Mass:              mass,
		Radius:            radius,
		Temp:              temp,
		Luminosity:        luminosity,
		Metallicity:       metallicity,
		Age:               age,
		FlareFrequency:    int16(flareFreq),
		SolarWind:         int16(solarWind),
		Variability:       int16(rng.IntN(20)),
		HabitableInner:    hzInner,
		HabitableOuter:    hzOuter,
		FrostLine:         frostLine,
		PosX:              x, PosY: y, PosZ: z,
	}
}

func pickSpectralClass(rng *rand.Rand, classes []db.SpectralClass) db.SpectralClass {
	// Group by letter
	byLetter := make(map[byte][]db.SpectralClass)
	for _, c := range classes {
		byLetter[c.Letter[0]] = append(byLetter[c.Letter[0]], c)
	}

	// Weighted pick of letter (deterministic order)
	total := 0
	for _, l := range letterOrder {
		total += spectralWeight[l]
	}
	roll := rng.IntN(total)
	var letter byte
	for _, l := range letterOrder {
		roll -= spectralWeight[l]
		if roll < 0 {
			letter = l
			break
		}
	}

	// Random subtype within letter
	pool := byLetter[letter]
	if len(pool) == 0 {
		return classes[0] // fallback
	}
	return pool[rng.IntN(len(pool))]
}

func pickLuminosityClass(rng *rand.Rand, classes []db.LuminosityClass) db.LuminosityClass {
	// Heavily weighted toward main sequence.
	weights := map[string]int{
		"Ia": 2, "Ib": 8, "II": 15, "III": 50,
		"IV": 70, "V": 850, "VI": 10, "VII": 0,
	}
	total := 0
	for _, lc := range classes {
		total += weights[lc.Numeral]
	}
	if total == 0 {
		return classes[0]
	}
	roll := rng.IntN(total)
	for _, lc := range classes {
		roll -= weights[lc.Numeral]
		if roll < 0 {
			return lc
		}
	}
	return classes[0]
}

func findStage(stages []db.LifecycleStage, name string) db.LifecycleStage {
	for _, s := range stages {
		if s.Name == name {
			return s
		}
	}
	return stages[0]
}

func flareFrequency(rng *rand.Rand, letter byte) int {
	base := map[byte]int{
		'O': 0, 'B': 0, 'A': 1, 'F': 2,
		'G': 3, 'K': 5, 'M': 15,
	}
	b := base[letter]
	return b + rng.IntN(b+1)
}

func solarWindStrength(rng *rand.Rand, mass int32) int {
	base := int(mass) / 10 // 1 solar → 100
	if base < 10 {
		base = 10
	}
	return base + rng.IntN(base/2+1)
}

func starAge(rng *rand.Rand, mass int32) int32 {
	// Main sequence lifetime ∝ M/L ∝ M^-2.5 (roughly)
	// Sun ≈ 10000 Myr
	m := float64(mass) / 1000.0
	if m < 0.1 {
		m = 0.1
	}
	maxAge := 10000.0 / (m * m * math.Sqrt(m))
	if maxAge > 13000 {
		maxAge = 13000 // age of universe
	}
	if maxAge < 10 {
		maxAge = 10
	}
	minAge := maxAge * 0.1
	return int32(minAge + rng.Float64()*(maxAge-minAge))
}

func randBetween(rng *rand.Rand, min, max int32) int32 {
	if max <= min {
		return min
	}
	return min + rng.Int32N(max-min+1)
}

func clamp64(v, lo, hi int64) int64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
