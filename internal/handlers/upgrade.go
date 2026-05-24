package handlers

import "net/http"

func (s *Server) handleUpgrade(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	project, err := s.docker.UpgradeProject(r.Context(), id)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.writeJSON(w, http.StatusOK, project)
}
