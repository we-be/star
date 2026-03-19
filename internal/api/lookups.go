package api

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
)

func (s *Server) listSpectralClasses(w http.ResponseWriter, r *http.Request) {
	classes, err := s.svc.Queries().ListSpectralClasses(r.Context())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, classes)
}

func (s *Server) getSpectralClass(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid spectral class id")
		return
	}
	class, err := s.svc.Queries().GetSpectralClass(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		writeErr(w, http.StatusNotFound, "spectral class not found")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, class)
}

func (s *Server) listLuminosityClasses(w http.ResponseWriter, r *http.Request) {
	classes, err := s.svc.Queries().ListLuminosityClasses(r.Context())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, classes)
}

func (s *Server) listLifecycleStages(w http.ResponseWriter, r *http.Request) {
	stages, err := s.svc.Queries().ListLifecycleStages(r.Context())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, stages)
}
