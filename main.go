package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tayyabmasood911/go-grpc-microservice-boilerplate/internal/interceptor"
	"github.com/tayyabmasood911/go-grpc-microservice-boilerplate/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

func main() {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50051"
	}
	addr := fmt.Sprintf(":%s", port)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("[FATAL] Failed to bind TCP listener on %s: %v", addr, err)
	}

	// Server options: interceptors, keepalive enforcement, max message limits
	serverOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			interceptor.RecoveryUnaryInterceptor(),
			interceptor.LoggingUnaryInterceptor(),
		),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     15 * time.Minute,
			MaxConnectionAge:      30 * time.Minute,
			MaxConnectionAgeGrace: 5 * time.Second,
			Time:                  5 * time.Minute,
			Timeout:               20 * time.Second,
		}),
		grpc.MaxRecvMsgSize(1024 * 1024 * 16), // 16 MB max recv
		grpc.MaxSendMsgSize(1024 * 1024 * 16), // 16 MB max send
	}

	grpcServer := grpc.NewServer(serverOpts...)

	// Initialize core microservice business services
	healthService := service.NewHealthServiceServer()
	userService := service.NewUserServiceServer()

	// Seed sample domain user
	_, _ = userService.CreateUser(context.Background(), "tayyabmasood911", "tayyabmasood911@gmail.com", "platform-architect")

	// Enable standard gRPC Server Reflection for grpcurl / Postman inspection
	reflection.Register(grpcServer)

	// Graceful shutdown channel
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("[INFO] Enterprise gRPC Microservice listening on %s (Reflection Enabled)", addr)
		if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			log.Fatalf("[FATAL] gRPC server runtime failure: %v", err)
		}
	}()

	sig := <-stopChan
	log.Printf("[INFO] Received signal %v. Initiating graceful gRPC server termination...", sig)
	healthService.SetStatus("", service.StatusNotServing)

	// Gracefully drain active RPC connections
	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		log.Println("[INFO] gRPC server stopped cleanly.")
	case <-time.After(10 * time.Second):
		log.Println("[WARN] Graceful stop timed out after 10s. Forcing immediate termination.")
		grpcServer.Stop()
	}
}
