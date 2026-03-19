package gen

import (
	"math"
	"math/rand/v2"
)

// Binary probability by primary spectral type.
var binaryProbability = map[byte]int{
	'O': 75, 'B': 60, 'A': 50, 'F': 45,
	'G': 45, 'K': 40, 'M': 30,
}

// SystemResult holds a generated system and its stars.
type SystemResult struct {
	ID               int64
	Name             string
	PosX, PosY, PosZ int32
	Stars            []SystemStarResult
}

// SystemStarResult is a star with its orbital relationship to the barycenter.
type SystemStarResult struct {
	Star          StarResult
	SemiMajorAxis int32 // milliau
	Eccentricity  int32 // × 10000
	Inclination   int32 // degrees × 100
	OrbitalPeriod int32 // game-years × 1000
	IsPrimary     bool
}

// GenerateSystem creates a star system at the given position.
// Rolls for binary based on primary star type.
func GenerateSystem(rng *rand.Rand, cfg *StarConfig, x, y, z int32) SystemResult {
	primary := GenerateStar(rng, cfg, x, y, z)

	sys := SystemResult{
		ID:   NextID(),
		Name: SystemName(rng),
		PosX: x, PosY: y, PosZ: z,
		Stars: []SystemStarResult{{
			Star:      primary,
			IsPrimary: true,
		}},
	}

	// Look up primary's spectral letter for binary probability
	letter := spectralLetter(cfg, primary.SpectralClassID)
	if rng.IntN(100) < binaryProbability[letter] {
		addCompanion(rng, cfg, &sys)
	}

	return sys
}

func addCompanion(rng *rand.Rand, cfg *StarConfig, sys *SystemResult) {
	primary := &sys.Stars[0]
	secondary := GenerateStar(rng, cfg, sys.PosX, sys.PosY, sys.PosZ)

	// Companion should be less massive — reroll mass if needed
	if secondary.Mass > primary.Star.Mass {
		// Constrain to half-to-full primary mass
		maxM := primary.Star.Mass
		minM := maxM / 2
		if minM < 80 { // minimum stellar mass ~0.08 solar
			minM = 80
		}
		secondary.Mass = randBetween(rng, minM, maxM)
		secondary.Luminosity = MassToLuminosity(secondary.Mass)
		secondary.Radius = MassToRadius(secondary.Mass)
		hzI, hzO := HabitableZone(secondary.Luminosity)
		secondary.HabitableInner = hzI
		secondary.HabitableOuter = hzO
		secondary.FrostLine = FrostLine(secondary.Luminosity)
	}

	// Orbital parameters
	separation := binarySeparation(rng) // milliau
	ecc := int32(rng.IntN(6000))        // 0.0 – 0.6
	incl := int32(rng.IntN(18000))      // 0 – 180 deg × 100

	// Kepler's 3rd: P² = a³ / M_total (years, AU, solar masses)
	aAU := float64(separation) / 1000.0
	mTotal := float64(primary.Star.Mass+secondary.Mass) / 1000.0
	if mTotal < 0.01 {
		mTotal = 0.01
	}
	pYears := math.Sqrt(aAU * aAU * aAU / mTotal)
	period := int32(pYears * 1000)

	// Both stars orbit barycenter. Primary SMA = sep × m2/(m1+m2).
	totalMass := float64(primary.Star.Mass + secondary.Mass)
	primarySMA := int32(float64(separation) * float64(secondary.Mass) / totalMass)
	secondarySMA := separation - primarySMA

	primary.SemiMajorAxis = primarySMA
	primary.Eccentricity = ecc
	primary.Inclination = incl
	primary.OrbitalPeriod = period

	sys.Stars = append(sys.Stars, SystemStarResult{
		Star:          secondary,
		SemiMajorAxis: secondarySMA,
		Eccentricity:  ecc,
		Inclination:   incl,
		OrbitalPeriod: period,
		IsPrimary:     false,
	})
}

// binarySeparation returns a separation in milliau drawn from a
// log-normal distribution centered around ~50 AU.
func binarySeparation(rng *rand.Rand) int32 {
	// log10(50) ≈ 1.7
	logSep := 1.7 + rng.NormFloat64()*1.0
	if logSep < -1 { // min 0.1 AU
		logSep = -1
	}
	if logSep > 4 { // max 10000 AU
		logSep = 4
	}
	return int32(math.Pow(10, logSep) * 1000)
}

func spectralLetter(cfg *StarConfig, classID int64) byte {
	for _, sc := range cfg.SpectralClasses {
		if sc.ID == classID {
			return sc.Letter[0]
		}
	}
	return 'G'
}
