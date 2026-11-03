package sqlservice

import (
	"context"
	"testing"
)

func TestUserRegistrationValidation(t *testing.T) {
	svc := NewUserService(nil)

	// Missing username
	_, err := svc.RegisterUser(context.Background(), "", "test@example.com", "admin")
	if err == nil {
		t.Fatalf("expected error for empty username, got nil")
	}

	// Invalid email
	_, err = svc.RegisterUser(context.Background(), "alice", "invalid-email", "admin")
	if err == nil {
		t.Fatalf("expected error for invalid email, got nil")
	}
}

func TestLookupValidation(t *testing.T) {
	svc := NewUserService(nil)
	_, err := svc.LookupUser(context.Background(), "  ")
	if err == nil {
		t.Fatalf("expected error for whitespace username query")
	}
}

