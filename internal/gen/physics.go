package gen

import "math"

// MassToLuminosity returns luminosity (solar × 1000) from mass (solar × 1000)
// using the empirical mass-luminosity relation for main sequence stars.
//
//	M < 0.43:  L = 0.23 × M^2.3
//	0.43–2.0:  L = M^4
//	2.0–55:    L = 1.4 × M^3.5
//	M > 55:    L = 32000 × M  (Eddington limit)
func MassToLuminosity(mass int32) int32 {
	m := float64(mass) / 1000.0
	var l float64
	switch {
	case m < 0.43:
		l = 0.23 * math.Pow(m, 2.3)
	case m < 2.0:
		l = math.Pow(m, 4.0)
	case m < 55.0:
		l = 1.4 * math.Pow(m, 3.5)
	default:
		l = 32000.0 * m
	}
	v := l * 1000
	if v > math.MaxInt32 {
		return math.MaxInt32
	}
	if v < 1 {
		return 1
	}
	return int32(v)
}

// MassToRadius returns radius (solar × 1000) from mass (solar × 1000).
//
//	M < 1:  R = M^0.8
//	M >= 1: R = M^0.57
func MassToRadius(mass int32) int32 {
	m := float64(mass) / 1000.0
	var r float64
	if m < 1.0 {
		r = math.Pow(m, 0.8)
	} else {
		r = math.Pow(m, 0.57)
	}
	result := int32(r * 1000)
	if result < 1 {
		result = 1
	}
	return result
}

// HabitableZone returns (inner, outer) in milliau from luminosity (solar × 1000).
//
//	inner = sqrt(L) × 0.95 AU
//	outer = sqrt(L) × 1.37 AU
func HabitableZone(luminosity int32) (inner, outer int32) {
	sqrtL := math.Sqrt(float64(luminosity) / 1000.0)
	return int32(sqrtL * 950), int32(sqrtL * 1370)
}

// FrostLine returns the frost line distance in milliau from luminosity (solar × 1000).
//
//	frost = sqrt(L) × 2.7 AU
func FrostLine(luminosity int32) int32 {
	return int32(math.Sqrt(float64(luminosity)/1000.0) * 2700)
}
