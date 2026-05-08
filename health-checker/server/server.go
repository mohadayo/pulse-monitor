package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mohadayo/pulse-monitor/health-checker/checker"
)

type Server struct {
	checker *checker.Checker
	router  *chi.Mux
	logger  *slog.Logger
}

type HealthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
}

type CheckRequest struct {
	URL string `json:"url"`
}

func New(c *checker.Checker, logger *slog.Logger) *Server {
	s := &Server{
		checker: c,
		router:  chi.NewRouter(),
		logger:  logger,
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.router.Get("/health", s.handleHealth)
	s.router.Post("/check", s.handleCheck)
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("health check requested")
	resp := HealthResponse{
		Status:    "ok",
		Service:   "health-checker",
		Version:   "1.0.0",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleCheck(w http.ResponseWriter, r *http.Request) {
	var req CheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("invalid request body", "error", err)
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		s.logger.Error("missing url in request")
		http.Error(w, `{"error":"url is required"}`, http.StatusBadRequest)
		return
	}

	s.logger.Info("performing health check", "target_url", req.URL)
	result := s.checker.Check(context.Background(), req.URL)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
