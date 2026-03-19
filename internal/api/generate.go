package api

import (
	"encoding/json"
	"net/http"

	"github.com/we-be/star/internal/gen"
)

type generateSystemRequest struct {
	X int32 `json:"x"`
	Y int32 `json:"y"`
	Z int32 `json:"z"`
}

func (s *Server) generateSystem(w http.ResponseWriter, r *http.Request) {
	var req generateSystemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json: "+err.Error())
		return
	}
	sys, err := s.svc.GenerateSystem(r.Context(), req.X, req.Y, req.Z)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, sys)
}

type generateUniverseRequest struct {
	CenterX    int32  `json:"center_x"`
	CenterY    int32  `json:"center_y"`
	CenterZ    int32  `json:"center_z"`
	Radius     int32  `json:"radius"`
	NumSystems int    `json:"num_systems"`
	Seed       uint64 `json:"seed"`
}

func (s *Server) generateUniverse(w http.ResponseWriter, r *http.Request) {
	var req generateUniverseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json: "+err.Error())
		return
	}
	if req.NumSystems <= 0 || req.NumSystems > 100000 {
		writeErr(w, http.StatusBadRequest, "num_systems must be 1–100000")
		return
	}
	if req.Radius <= 0 {
		writeErr(w, http.StatusBadRequest, "radius must be positive")
		return
	}
	result, err := s.svc.GenerateUniverse(r.Context(), gen.UniverseOpts{
		CenterX:    req.CenterX,
		CenterY:    req.CenterY,
		CenterZ:    req.CenterZ,
		Radius:     req.Radius,
		NumSystems: req.NumSystems,
		Seed:       req.Seed,
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, result)
}
