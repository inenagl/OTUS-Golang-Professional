package internalhttp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Server struct {
	host   string
	port   int
	logger zap.Logger
	app    Application
	server *http.Server
}

type Application interface { // TODO
}

func NewServer(host string, port int, logger zap.Logger, app Application) *Server {
	return &Server{
		host:   host,
		port:   port,
		logger: logger,
		app:    app,
	}
}

func (s *Server) Start(ctx context.Context) error {
	timeout := 10 * time.Second
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.host, s.port),
		Handler:      s.router(),
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
	}
	s.server = server

	server.ListenAndServe()

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) router() *mux.Router {
	rtr := mux.NewRouter()
	rtr.Use(s.loggingMiddleware)

	rtr.HandleFunc("/hello", s.hello)

	return rtr
}

func (s *Server) hello(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)

	_, err := w.Write([]byte("Hello, Stranger!"))
	if err != nil {
		s.logger.Error(err.Error())
	}
}
