package gen

import (
	"math"
	"math/rand/v2"
)

// UniverseOpts controls procedural generation of a region of space.
type UniverseOpts struct {
	CenterX, CenterY, CenterZ int32
	Radius                     int32  // coordinate units
	NumSystems                 int
	Seed                       uint64 // for reproducibility
}

// GenerateUniverse fills a region with star systems using a
// galactic disk density profile with spiral arm structure.
func GenerateUniverse(rng *rand.Rand, cfg *StarConfig, opts UniverseOpts) []SystemResult {
	systems := make([]SystemResult, 0, opts.NumSystems)
	for range opts.NumSystems {
		x, y, z := samplePosition(rng, opts)
		sys := GenerateSystem(rng, cfg, x, y, z)
		systems = append(systems, sys)
	}
	return systems
}

// samplePosition generates a point in a galactic disk with:
//   - exponential radial falloff (scale length = radius/3)
//   - 2-arm logarithmic spiral perturbation
//   - thin Gaussian vertical distribution (5% of radius)
func samplePosition(rng *rand.Rand, opts UniverseOpts) (x, y, z int32) {
	r := float64(opts.Radius)
	scale := r / 3.0

	for {
		// Exponential radial distance
		radial := rng.ExpFloat64() * scale
		if radial > r {
			continue
		}

		// Uniform azimuth
		theta := rng.Float64() * 2 * math.Pi

		// Spiral arm perturbation (2 arms, logarithmic)
		armPhase := theta + (radial/r)*4*math.Pi
		armPull := 0.3 * r * math.Exp(-radial/(r*0.7)) * math.Cos(2*armPhase)
		radial += armPull
		if radial < 0 || radial > r {
			continue
		}

		// Thin disk vertical
		vertical := rng.NormFloat64() * r * 0.05

		x = opts.CenterX + int32(radial*math.Cos(theta))
		y = opts.CenterY + int32(radial*math.Sin(theta))
		z = opts.CenterZ + int32(vertical)
		return
	}
}
