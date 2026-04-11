package internalhttp

import (
	"context"
	"net"
	"net/http"

	"github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/server/http/handlers"
	"github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/server/http/router"
)

type Server struct {
	logger     Logger
	httpLogger Logger
	app        handlers.Application
	httpServer *http.Server
	address    string
}

type Logger interface {
	Info(msg string)
	Infof(format string, args ...interface{})
	Error(msg string)
	Errorf(format string, args ...interface{})
}

func NewServer(logger Logger, httpLogger Logger, app handlers.Application, host, port string) *Server {
	address := net.JoinHostPort(host, port)

	return &Server{
		logger:     logger,
		httpLogger: httpLogger,
		app:        app,
		address:    address,
	}
}

func (s *Server) Start(ctx context.Context) error {
	handler := router.NewRouter(s.httpLogger, s.app)

	s.httpServer = &http.Server{
		Addr:    s.address,
		Handler: handler,
	}

	s.logger.Info("starting HTTP server on " + s.address)

	errChan := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return nil
	}
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("stopping HTTP server...")

	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			return err
		}
	}

	s.logger.Info("HTTP server stopped")
	return nil
}
