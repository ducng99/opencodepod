package handlers

import (
	"encoding/json"
	"net/http"

	"opencodepod/internal/config"
	"opencodepod/internal/docker"
)

type Server struct {
	cfg    *config.Config
	docker *docker.DockerManager
}

func NewServer(cfg *config.Config, docker *docker.DockerManager) *Server {
	return &Server{cfg: cfg, docker: docker}
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/projects", s.handleList)
	mux.HandleFunc("POST /api/projects", s.handleCreate)
	mux.HandleFunc("GET /api/projects/{id}", s.handleGet)
	mux.HandleFunc("PATCH /api/projects/{id}", s.handleUpdate)
	mux.HandleFunc("POST /api/projects/{id}/start", s.handleStart)
	mux.HandleFunc("POST /api/projects/{id}/stop", s.handleStop)
	mux.HandleFunc("POST /api/projects/{id}/upgrade", s.handleUpgrade)
	mux.HandleFunc("DELETE /api/projects/{id}", s.handleDelete)
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (s *Server) writeError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}
