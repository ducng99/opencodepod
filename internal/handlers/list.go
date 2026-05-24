package handlers

import "net/http"

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	projects, err := s.docker.ListProjects(r.Context())
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.writeJSON(w, http.StatusOK, projects)
}
