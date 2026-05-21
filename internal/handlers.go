package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Server struct {
	cfg    *Config
	docker *DockerManager
}

func NewServer(cfg *Config, docker *DockerManager) *Server {
	return &Server{cfg: cfg, docker: docker}
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/projects", s.handleList)
	mux.HandleFunc("POST /api/projects", s.handleCreate)
	mux.HandleFunc("GET /api/projects/{id}", s.handleGet)
	mux.HandleFunc("POST /api/projects/{id}/start", s.handleStart)
	mux.HandleFunc("POST /api/projects/{id}/stop", s.handleStop)
	mux.HandleFunc("POST /api/projects/{id}/upgrade", s.handleUpgrade)
	mux.HandleFunc("DELETE /api/projects/{id}", s.handleDelete)
}

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	projects, err := s.docker.ListProjects(r.Context())
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.writeJSON(w, http.StatusOK, projects)
}

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		s.writeError(w, http.StatusBadRequest, fmt.Errorf("name is required"))
		return
	}
	project, err := s.docker.CreateProject(r.Context(), &req)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.writeJSON(w, http.StatusCreated, project)
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	project, err := s.docker.GetProject(r.Context(), id)
	if err != nil {
		s.writeError(w, http.StatusNotFound, err)
		return
	}
	s.writeJSON(w, http.StatusOK, project)
}

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	project, err := s.docker.StartProject(r.Context(), id)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.writeJSON(w, http.StatusOK, project)
}

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	project, err := s.docker.StopProject(r.Context(), id)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.writeJSON(w, http.StatusOK, project)
}

func (s *Server) handleUpgrade(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	project, err := s.docker.UpgradeProject(r.Context(), id)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.writeJSON(w, http.StatusOK, project)
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.docker.DeleteProject(r.Context(), id); err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (s *Server) writeError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}
