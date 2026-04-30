package service

import "testing"

func TestValidateRegisterRequest(t *testing.T) {
	err := validateRegisterRequest(RegisterRequest{Email: "bad", Password: "short", FullName: "A"})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if err.Code != "validation_error" {
		t.Fatalf("unexpected code %q", err.Code)
	}
	if err.Fields["email"] != "invalid_format" {
		t.Fatalf("expected invalid email error, got %#v", err.Fields)
	}
	if err.Fields["password"] != "too_short" {
		t.Fatalf("expected short password error, got %#v", err.Fields)
	}
	if err.Fields["full_name"] != "too_short" {
		t.Fatalf("expected short full name error, got %#v", err.Fields)
	}
}

func TestHashRefreshTokenIsDeterministic(t *testing.T) {
	first := hashRefreshToken("refresh-token")
	second := hashRefreshToken("refresh-token")
	if first != second {
		t.Fatal("expected refresh token hash to be deterministic")
	}
}
