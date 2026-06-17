package api

import (
	"encoding/json"
	"net/http"

	"github.com/teddyandria/sentinel/internal/domain"
)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleConfig expose le token Mapbox au front (token pk.* public, restreint par domaine).
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"mapboxToken": s.mapboxToken})
}

// handleTopics renvoie la liste des sujets disponibles, pour que le frontend
// construise ses boutons de filtre sans les coder en dur.
func (s *Server) handleTopics(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, domain.AllowedTopics)
}

func (s *Server) handleListArticles(w http.ResponseWriter, r *http.Request) {
	// ?topic= filtre la carte ; absent = tous les sujets. On valide les valeurs connues.
	topic := r.URL.Query().Get("topic")
	if topic != "" && !domain.IsAllowedTopic(topic) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "topic inconnu"})
		return
	}

	articles, err := s.store.ListGeolocated(r.Context(), topic)
	if err != nil {
		s.log.Error("list articles", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, articles)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
