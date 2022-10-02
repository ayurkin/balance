package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
	"net"
	"net/http"
)

type Server struct {
	server *http.Server
	logger *zap.SugaredLogger
}

func New(logger *zap.SugaredLogger) *Server {
	return &Server{server: &http.Server{}, logger: logger}
}

func (s *Server) Start(port string) error {
	listen, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %v", err, port)
	}

	s.server.Handler = s.routes()

	if err := s.server.Serve(listen); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to serve http server over port %s: %v", err, port)
	}
	return nil
}

func (s *Server) routes() http.Handler {
	r := chi.NewMux()
	r.Get("/health", s.healthHandler)
	return r
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
