package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type User struct {
	ID        string
	Username  string
	Email     string
	Role      string
	CreatedAt int64
}

// UserServiceServer implements production in-memory safe user management
type UserServiceServer struct {
	mu    sync.RWMutex
	users map[string]*User
}

func NewUserServiceServer() *UserServiceServer {
	return &UserServiceServer{
		users: make(map[string]*User),
	}
}

func (s *UserServiceServer) CreateUser(ctx context.Context, username, email, role string) (*User, error) {
	if username == "" || email == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username and email cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, u := range s.users {
		if u.Email == email {
			return nil, status.Errorf(codes.AlreadyExists, "user with email %s already exists", email)
		}
	}

	userID := fmt.Sprintf("usr_%d", time.Now().UnixNano())
	newUser := &User{
		ID:        userID,
		Username:  username,
		Email:     email,
		Role:      role,
		CreatedAt: time.Now().Unix(),
	}

	s.users[userID] = newUser
	return newUser, nil
}

func (s *UserServiceServer) GetUser(ctx context.Context, userID string) (*User, error) {
	if userID == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	u, exists := s.users[userID]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "user %s not found", userID)
	}

	return u, nil
}
