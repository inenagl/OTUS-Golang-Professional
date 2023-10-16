package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/app"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	host   string
	port   int
	logger zap.Logger
	app    app.Application
	server *grpc.Server
}

func NewServer(host string, port int, logger zap.Logger, app app.Application) *Server {
	return &Server{
		host:   host,
		port:   port,
		logger: logger,
		app:    app,
	}
}

func (s *Server) Start(_ context.Context) error {
	lsn, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		return err
	}

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(loggingInterceptor(&s.logger)),
	)
	service := NewService(s.app, s.logger)
	RegisterCalendarServer(server, service)
	s.server = server

	s.logger.Debug(fmt.Sprintf("starting grpc server on %s", lsn.Addr().String()))
	if err := server.Serve(lsn); err != nil {
		s.logger.Error(err.Error())
	}

	return nil
}

func (s *Server) Stop(_ context.Context) error {
	s.logger.Debug("GRPC stop")
	s.server.GracefulStop()
	return nil
}
