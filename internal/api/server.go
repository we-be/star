package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/we-be/star/internal/service"
)

// Server holds the service layer and HTTP handler.
type Server struct {
	svc *service.Service
	mux *http.ServeMux
}

// New creates a Server and registers all routes.
func New(svc *service.Service) *Server {
	s := &Server{svc: svc, mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	// Stars
	s.mux.HandleFunc("GET /stars/{id}", s.getStar)
	s.mux.HandleFunc("POST /stars", s.insertStar)
	s.mux.HandleFunc("PUT /stars/{id}/lifecycle", s.updateStarLifecycle)

	// Spatial
	s.mux.HandleFunc("GET /stars/nearby", s.findStarsWithinRadius)
	s.mux.HandleFunc("GET /stars/nearest", s.findNearestStars)
	s.mux.HandleFunc("GET /systems/nearby", s.findSystemsWithinRadius)
	s.mux.HandleFunc("GET /systems/nearest", s.findNearestSystems)

	// Systems
	s.mux.HandleFunc("GET /systems/{id}", s.getSystem)
	s.mux.HandleFunc("POST /systems", s.insertSystem)
	s.mux.HandleFunc("GET /systems/{id}/stars", s.getSystemStars)
	s.mux.HandleFunc("POST /systems/{id}/stars", s.insertSystemStar)
	s.mux.HandleFunc("GET /systems/{id}/environment", s.getStellarEnvironment)

	// Generation
	s.mux.HandleFunc("POST /generate/system", s.generateSystem)
	s.mux.HandleFunc("POST /generate/universe", s.generateUniverse)

	// Lookups
	s.mux.HandleFunc("GET /spectral-classes", s.listSpectralClasses)
	s.mux.HandleFunc("GET /spectral-classes/{id}", s.getSpectralClass)
	s.mux.HandleFunc("GET /luminosity-classes", s.listLuminosityClasses)
	s.mux.HandleFunc("GET /lifecycle-stages", s.listLifecycleStages)

	// Health
	s.mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Static UI
	s.mux.Handle("GET /", http.FileServer(http.Dir("static")))
}

// --- helpers ---

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("json encode: %v", err)
	}
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func pathID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

func queryInt32(r *http.Request, key string, fallback int32) int32 {
	v := r.URL.Query().Get(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return fallback
	}
	return int32(n)
}

func queryFloat64(r *http.Request, key string, fallback float64) float64 {
	v := r.URL.Query().Get(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fallback
	}
	return n
}
