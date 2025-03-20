package api

import (
	"encoding/json"
	"net/http"

	"github.com/tlscert/backend/internal/manager"
)

type Server struct {
	cm *manager.CertificateManager
}

func NewServer(cm *manager.CertificateManager) *Server {
	return &Server{
		cm: cm,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/certificate":
		s.handleGetCertificate(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleGetCertificate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	certificate, err := s.cm.GetCertificate(r.Context())

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(certificate)
}
