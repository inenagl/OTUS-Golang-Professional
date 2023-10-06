package internalhttp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/storage"
	"go.uber.org/zap"
)

type Server struct {
	host   string
	port   int
	logger zap.Logger
	app    Application
	server *http.Server
}

type Application interface {
	GetEvent(id uuid.UUID, userID uuid.UUID) (storage.Event, error)
	UpdateEvent(id uuid.UUID, userID uuid.UUID, event storage.Event) (storage.Event, error)
	CreateEvent(userID uuid.UUID, event storage.Event) (storage.Event, error)
	DeleteEvent(id uuid.UUID, userID uuid.UUID) error
	GetEventsForPeriod(userID uuid.UUID, start, end time.Time) ([]storage.Event, error)
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

	// Оставлю как открытую часть API
	rtr.HandleFunc("/hello", s.hello)

	// Делаем саброутер для закрытой части апи (требующей передачи id юзера в заголовке)
	restricted := rtr.NewRoute().Subrouter()
	restricted.Use(s.userMiddleware)
	restricted.HandleFunc("/event/{eventId}", s.getEvent).Methods("GET")
	restricted.HandleFunc("/event/{eventId}", s.updateEvent).Methods("POST")
	restricted.HandleFunc("/event/{eventId}", s.deleteEvent).Methods("DELETE")
	restricted.HandleFunc("/event", s.createEvent).Methods("POST")
	restricted.HandleFunc("/events/day/{date}", s.getForDay).Methods("GET")
	restricted.HandleFunc("/events/week/{date}", s.getForWeek).Methods("GET")
	restricted.HandleFunc("/events/month/{date}", s.getForMonth).Methods("GET")

	return rtr
}

func (s *Server) hello(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)

	_, err := w.Write([]byte("Hello, Stranger!"))
	if err != nil {
		s.logger.Error(err.Error())
	}
}
