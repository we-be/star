package gen

import (
	"math"
	"testing"
)

func TestMassToLuminosity_Sun(t *testing.T) {
	// Sun: mass=1000 (1.0 solar), expected luminosity ≈ 1000 (1.0 solar)
	l := MassToLuminosity(1000)
	if l < 900 || l > 1100 {
		t.Errorf("Sun luminosity: got %d, want ~1000", l)
	}
}

func TestMassToLuminosity_Ranges(t *testing.T) {
	tests := []struct {
		name    string
		mass    int32  // solar × 1000
		wantMin int32
		wantMax int32
	}{
		{"red dwarf (0.1 solar)", 100, 1, 5},
		{"Sun (1.0 solar)", 1000, 900, 1100},
		{"Sirius-like (2.0 solar)", 2000, 14000, 18000},
		{"massive (20 solar)", 20000, 30000000, 100000000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MassToLuminosity(tt.mass)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("MassToLuminosity(%d) = %d, want [%d, %d]", tt.mass, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestMassToRadius_Sun(t *testing.T) {
	r := MassToRadius(1000)
	// M=1.0 → R=1.0^0.8 = 1.0 (at the boundary, uses M<1 branch)
	// Actually 999/1000 < 1.0, so it uses 0.999^0.8 ≈ 0.999
	if r < 900 || r > 1100 {
		t.Errorf("Sun radius: got %d, want ~1000", r)
	}
}

func TestHabitableZone_Sun(t *testing.T) {
	inner, outer := HabitableZone(1000)
	// Sun: inner ≈ 950 milliau (0.95 AU), outer ≈ 1370 milliau (1.37 AU)
	if math.Abs(float64(inner)-950) > 50 {
		t.Errorf("HZ inner: got %d, want ~950", inner)
	}
	if math.Abs(float64(outer)-1370) > 50 {
		t.Errorf("HZ outer: got %d, want ~1370", outer)
	}
}

func TestFrostLine_Sun(t *testing.T) {
	fl := FrostLine(1000)
	// Sun: frost ≈ 2700 milliau (2.7 AU)
	if math.Abs(float64(fl)-2700) > 100 {
		t.Errorf("frost line: got %d, want ~2700", fl)
	}
}

func TestMassToLuminosity_Monotonic(t *testing.T) {
	// Luminosity should increase monotonically with mass
	prev := int32(0)
	for mass := int32(80); mass <= 150000; mass += 100 {
		l := MassToLuminosity(mass)
		if l < prev {
			t.Errorf("non-monotonic at mass=%d: L=%d < prev=%d", mass, l, prev)
		}
		prev = l
	}
}
