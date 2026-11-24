package interceptor

import (
	"context"
	"log"
	"runtime/debug"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LoggingUnaryInterceptor records RPC method, execution duration, and response status code.
func LoggingUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		statusCode := codes.OK
		if err != nil {
			if st, ok := status.FromError(err); ok {
				statusCode = st.Code()
			} else {
				statusCode = codes.Unknown
			}
		}

		log.Printf("[gRPC] method=%s duration=%s status=%s error=%v",
			info.FullMethod, duration, statusCode, err)
		return resp, err
	}
}

// RecoveryUnaryInterceptor recovers from panics, logs the stack trace, and returns codes.Internal.
func RecoveryUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC RECOVERED] method=%s panic=%v\nstack:\n%s",
					info.FullMethod, r, debug.Stack())
				err = status.Errorf(codes.Internal, "an unexpected internal error occurred")
			}
		}()
		return handler(ctx, req)
	}
}
