package tests

import (
	"testing"

	"github.com/AffanSurya/xarela-backend/internal/service"
)

func TestRegisterRejectsInvalidInput(t *testing.T) {
	serviceUnderTest := service.NewAuthService("")

	_, err := serviceUnderTest.Register(t.Context(), service.RegisterRequest{
		Email:    "bad",
		Password: "short",
		FullName: "A",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr, ok := err.(*service.ValidationError)
	if !ok {
		t.Fatalf("expected validation error type, got %T", err)
	}
	if validationErr.Code != "validation_error" {
		t.Fatalf("unexpected code %q", validationErr.Code)
	}
	if validationErr.Fields["email"] != "invalid_format" {
		t.Fatalf("expected invalid email error, got %#v", validationErr.Fields)
	}
	if validationErr.Fields["password"] != "too_short" {
		t.Fatalf("expected short password error, got %#v", validationErr.Fields)
	}
	if validationErr.Fields["full_name"] != "too_short" {
		t.Fatalf("expected short full name error, got %#v", validationErr.Fields)
	}
}

func TestVerifyEmailRejectsEmptyToken(t *testing.T) {
	serviceUnderTest := service.NewAuthService("")

	_, err := serviceUnderTest.VerifyEmail(t.Context(), service.VerifyEmailRequest{})
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr, ok := err.(*service.ValidationError)
	if !ok {
		t.Fatalf("expected validation error type, got %T", err)
	}
	if validationErr.Fields["verification_token"] != "required" {
		t.Fatalf("expected verification token required error, got %#v", validationErr.Fields)
	}
}