package service

import (
	"context"
	"fmt"
	"math/rand/v2"

	"github.com/we-be/star/internal/db"
	"github.com/we-be/star/internal/gen"
)

type Service struct {
	q   *db.Queries
	rng *rand.Rand
	cfg *gen.StarConfig
}

func New(q *db.Queries, seed uint64) *Service {
	return &Service{
		q:   q,
		rng: rand.New(rand.NewPCG(seed, seed^0xCAFEBABE)),
	}
}

// Queries exposes the raw DB layer for CRUD operations.
func (s *Service) Queries() *db.Queries { return s.q }

// LoadReferenceData caches spectral classes, luminosity classes, and lifecycle
// stages from the database. Must be called before any generation.
func (s *Service) LoadReferenceData(ctx context.Context) error {
	sc, err := s.q.ListSpectralClasses(ctx)
	if err != nil {
		return fmt.Errorf("load spectral classes: %w", err)
	}
	lc, err := s.q.ListLuminosityClasses(ctx)
	if err != nil {
		return fmt.Errorf("load luminosity classes: %w", err)
	}
	ls, err := s.q.ListLifecycleStages(ctx)
	if err != nil {
		return fmt.Errorf("load lifecycle stages: %w", err)
	}
	if len(sc) == 0 || len(lc) == 0 || len(ls) == 0 {
		return fmt.Errorf("reference data not seeded (spectral=%d, luminosity=%d, lifecycle=%d)",
			len(sc), len(lc), len(ls))
	}
	s.cfg = &gen.StarConfig{
		SpectralClasses:   sc,
		LuminosityClasses: lc,
		LifecycleStages:   ls,
	}
	return nil
}

// GenerateSystem creates a single star system at the given coordinates,
// persists it, and returns the result.
func (s *Service) GenerateSystem(ctx context.Context, x, y, z int32) (*gen.SystemResult, error) {
	sys := gen.GenerateSystem(s.rng, s.cfg, x, y, z)
	if err := s.persistSystem(ctx, &sys); err != nil {
		return nil, err
	}
	return &sys, nil
}

// UniverseResult summarizes a universe generation run.
type UniverseResult struct {
	Systems int `json:"systems"`
	Stars   int `json:"stars"`
	Binary  int `json:"binary"`
}

// GenerateUniverse fills a region of space with star systems.
// Uses a dedicated RNG seeded from opts.Seed for reproducibility.
func (s *Service) GenerateUniverse(ctx context.Context, opts gen.UniverseOpts) (*UniverseResult, error) {
	rng := rand.New(rand.NewPCG(opts.Seed, opts.Seed^0xDEADBEEF))
	systems := gen.GenerateUniverse(rng, s.cfg, opts)

	result := &UniverseResult{}
	for i := range systems {
		if err := s.persistSystem(ctx, &systems[i]); err != nil {
			return nil, fmt.Errorf("system %d/%d: %w", i+1, len(systems), err)
		}
		result.Systems++
		result.Stars += len(systems[i].Stars)
		if len(systems[i].Stars) > 1 {
			result.Binary++
		}
	}
	return result, nil
}

func (s *Service) persistSystem(ctx context.Context, sys *gen.SystemResult) error {
	_, err := s.q.InsertSystem(ctx, db.InsertSystemParams{
		ID:   sys.ID,
		Name: sys.Name,
		PosX: sys.PosX, PosY: sys.PosY, PosZ: sys.PosZ,
	})
	if err != nil {
		return fmt.Errorf("insert system %q: %w", sys.Name, err)
	}

	for _, ss := range sys.Stars {
		_, err := s.q.InsertStar(ctx, db.InsertStarParams{
			ID:                ss.Star.ID,
			Name:              ss.Star.Name,
			IDSpectralClass:   ss.Star.SpectralClassID,
			IDLuminosityClass: ss.Star.LuminosityClassID,
			IDLifecycleStage:  ss.Star.LifecycleStageID,
			Mass:              ss.Star.Mass,
			Radius:            ss.Star.Radius,
			Temp:              ss.Star.Temp,
			Luminosity:        ss.Star.Luminosity,
			Metallicity:       ss.Star.Metallicity,
			Age:               ss.Star.Age,
			FlareFrequency:    ss.Star.FlareFrequency,
			SolarWind:         ss.Star.SolarWind,
			Variability:       ss.Star.Variability,
			HabitableInner:    ss.Star.HabitableInner,
			HabitableOuter:    ss.Star.HabitableOuter,
			FrostLine:         ss.Star.FrostLine,
			PosX:              ss.Star.PosX,
			PosY:              ss.Star.PosY,
			PosZ:              ss.Star.PosZ,
		})
		if err != nil {
			return fmt.Errorf("insert star %q: %w", ss.Star.Name, err)
		}

		err = s.q.InsertSystemStar(ctx, db.InsertSystemStarParams{
			IDSystem:      sys.ID,
			IDStar:        ss.Star.ID,
			SemiMajorAxis: ss.SemiMajorAxis,
			Eccentricity:  ss.Eccentricity,
			Inclination:   ss.Inclination,
			OrbitalPeriod: ss.OrbitalPeriod,
			IsPrimary:     ss.IsPrimary,
		})
		if err != nil {
			return fmt.Errorf("link star %q to system: %w", ss.Star.Name, err)
		}
	}
	return nil
}
