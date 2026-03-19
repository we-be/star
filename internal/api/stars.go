package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/we-be/star/internal/db"
)

func (s *Server) getStar(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid star id")
		return
	}
	star, err := s.svc.Queries().GetStar(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		writeErr(w, http.StatusNotFound, "star not found")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, star)
}

func (s *Server) insertStar(w http.ResponseWriter, r *http.Request) {
	var p db.InsertStarParams
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json: "+err.Error())
		return
	}
	star, err := s.svc.Queries().InsertStar(r.Context(), p)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, star)
}

func (s *Server) updateStarLifecycle(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid star id")
		return
	}
	var p db.UpdateStarLifecycleParams
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json: "+err.Error())
		return
	}
	p.ID = id
	if err := s.svc.Queries().UpdateStarLifecycle(r.Context(), p); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) findStarsWithinRadius(w http.ResponseWriter, r *http.Request) {
	stars, err := s.svc.Queries().FindStarsWithinRadius(r.Context(), db.FindStarsWithinRadiusParams{
		Column1: queryFloat64(r, "x", 0),
		Column2: queryFloat64(r, "y", 0),
		Column3: queryFloat64(r, "z", 0),
		Column4: queryFloat64(r, "radius", 100),
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, stars)
}

func (s *Server) findNearestStars(w http.ResponseWriter, r *http.Request) {
	stars, err := s.svc.Queries().FindNearestStars(r.Context(), db.FindNearestStarsParams{
		Column1: queryFloat64(r, "x", 0),
		Column2: queryFloat64(r, "y", 0),
		Column3: queryFloat64(r, "z", 0),
		Limit:   queryInt32(r, "limit", 10),
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, stars)
}
