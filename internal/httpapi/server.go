package httpapi

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"

	"github.com/keepdevops/cofiswarm-slot-manager/internal/pressure"
)

type Server struct {
	mu        sync.RWMutex
	endpoints []pressure.Endpoint
}

func New(path string) (*Server, error) {
	s := &Server{}
	if path != "" {
		if err := s.LoadEndpoints(path); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *Server) LoadEndpoints(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var cfg struct {
		Endpoints []pressure.Endpoint `json:"endpoints"`
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		return err
	}
	s.mu.Lock()
	s.endpoints = cfg.Endpoints
	s.mu.Unlock()
	return nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/api/pressure", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		s.mu.RLock()
		eps := append([]pressure.Endpoint(nil), s.endpoints...)
		s.mu.RUnlock()
		enc := json.NewEncoder(w)
		_ = enc.Encode(pressure.Snapshot(eps))
	})
	mux.HandleFunc("/api/pressure/evict", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"accepted","note":"port to legacy/cpp/coordinator_kv_ops"}`))
	})
	mux.HandleFunc("/v1/endpoints", func(w http.ResponseWriter, _ *http.Request) {
		s.mu.RLock()
		defer s.mu.RUnlock()
		_ = json.NewEncoder(w).Encode(map[string]any{"endpoints": s.endpoints})
	})
	return mux
}
