package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"opencodepod/internal/project"
)

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req project.CreateRequest
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
