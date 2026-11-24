package service

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ServingStatus represents health status enumeration
type ServingStatus int32

const (
	StatusUnknown ServingStatus = 0
	StatusServing ServingStatus = 1
	StatusNotServing ServingStatus = 2
)

type HealthCheckRequest struct {
	Service string
}

type HealthCheckResponse struct {
	Status    ServingStatus
	Message   string
	Timestamp int64
}

// HealthServiceServer implements enterprise health check probes for Kubernetes / AWS ALB
type HealthServiceServer struct {
	mu       sync.RWMutex
	statuses map[string]ServingStatus
}

func NewHealthServiceServer() *HealthServiceServer {
	h := &HealthServiceServer{
		statuses: make(map[string]ServingStatus),
	}
	h.SetStatus("", StatusServing)
	h.SetStatus("api.v1.UserService", StatusServing)
	return h
}

func (s *HealthServiceServer) SetStatus(service string, st ServingStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.statuses[service] = st
}

func (s *HealthServiceServer) Check(ctx context.Context, req *HealthCheckRequest) (*HealthCheckResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	st, exists := s.statuses[req.Service]
	if !exists {
		return &HealthCheckResponse{
			Status:    StatusNotServing,
			Message:   "service not found or uninitialized",
			Timestamp: time.Now().Unix(),
		}, status.Errorf(codes.NotFound, "unknown service: %s", req.Service)
	}

	return &HealthCheckResponse{
		Status:    st,
		Message:   "service is operational and serving traffic",
		Timestamp: time.Now().Unix(),
	}, nil
}
