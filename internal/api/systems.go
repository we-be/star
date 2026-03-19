package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/we-be/star/internal/db"
)

func (s *Server) getSystem(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid system id")
		return
	}
	sys, err := s.svc.Queries().GetSystem(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		writeErr(w, http.StatusNotFound, "system not found")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sys)
}

func (s *Server) insertSystem(w http.ResponseWriter, r *http.Request) {
	var p db.InsertSystemParams
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json: "+err.Error())
		return
	}
	sys, err := s.svc.Queries().InsertSystem(r.Context(), p)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, sys)
}

func (s *Server) getSystemStars(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid system id")
		return
	}
	stars, err := s.svc.Queries().GetSystemStars(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, stars)
}

func (s *Server) insertSystemStar(w http.ResponseWriter, r *http.Request) {
	sysID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid system id")
		return
	}
	var p db.InsertSystemStarParams
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json: "+err.Error())
		return
	}
	p.IDSystem = sysID
	if err := s.svc.Queries().InsertSystemStar(r.Context(), p); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) getStellarEnvironment(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid system id")
		return
	}
	env, err := s.svc.Queries().GetStellarEnvironment(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		writeErr(w, http.StatusNotFound, "system not found or has no primary star")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, env)
}

func (s *Server) findSystemsWithinRadius(w http.ResponseWriter, r *http.Request) {
	systems, err := s.svc.Queries().FindSystemsWithinRadius(r.Context(), db.FindSystemsWithinRadiusParams{
		Column1: queryFloat64(r, "x", 0),
		Column2: queryFloat64(r, "y", 0),
		Column3: queryFloat64(r, "z", 0),
		Column4: queryFloat64(r, "radius", 100),
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, systems)
}

func (s *Server) findNearestSystems(w http.ResponseWriter, r *http.Request) {
	systems, err := s.svc.Queries().FindNearestSystems(r.Context(), db.FindNearestSystemsParams{
		Column1: queryFloat64(r, "x", 0),
		Column2: queryFloat64(r, "y", 0),
		Column3: queryFloat64(r, "z", 0),
		Limit:   queryInt32(r, "limit", 10),
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, systems)
}
