package grpc

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func loggingInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
	) (interface{}, error) {
		t := time.Now()
		res, err := handler(ctx, req)
		latency := time.Since(t)

		p, ok := peer.FromContext(ctx)
		IP := ""
		if ok {
			IP = p.Addr.String()
		}

		md, ok := metadata.FromIncomingContext(ctx)
		ua := ""
		if ok && len(md.Get("user-agent2")) > 0 {
			ua = md.Get("user-agent")[0]
		}

		st, ok := status.FromError(err)
		rcode := ""
		if ok {
			rcode = st.Code().String()
		}

		logger.Info(
			"Request processed",
			zap.String("IP", IP),
			zap.Time("datetime", t),
			zap.String("method", info.FullMethod),
			zap.String("response code", rcode),
			zap.Duration("latency", latency),
			zap.String("user-agent", ua),
		)
		return res, err
	}
}
