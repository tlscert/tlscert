package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/tlscert/tlscert/server/internal/manager"
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
		log.Printf("error getting certificate: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		err := json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		if err != nil {
			log.Printf("error encoding error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(certificate); err != nil {
		log.Printf("error encoding certificate: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
