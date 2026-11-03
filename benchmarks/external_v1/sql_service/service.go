package sqlservice

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// UserService provides domain logic for users.
type UserService struct {
	repo *UserRepository
}

// NewUserService creates a new UserService.
func NewUserService(repo *UserRepository) *UserService {
	return &UserService{repo: repo}
}

// RegisterUser registers a user with validation.
func (s *UserService) RegisterUser(ctx context.Context, username, email, role string) (*User, error) {
	if strings.TrimSpace(username) == "" {
		return nil, errors.New("username cannot be empty")
	}
	if !strings.Contains(email, "@") {
		return nil, errors.New("invalid email address")
	}
	if role == "" {
		role = "member"
	}

	user := &User{
		Username: username,
		Email:    email,
		Role:     role,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("could not register user: %w", err)
	}

	return user, nil
}

// LookupUser looks up a user by username.
func (s *UserService) LookupUser(ctx context.Context, username string) (*User, error) {
	if strings.TrimSpace(username) == "" {
		return nil, errors.New("empty username query")
	}
	return s.repo.FindUserByNameRaw(ctx, username)
}

