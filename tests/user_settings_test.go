package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/AffanSurya/xarela-backend/internal/repository"
	"github.com/AffanSurya/xarela-backend/internal/service"
	"github.com/google/uuid"
)

type fakeUserSettingsRepo struct {
	record       repository.UserSettingsRecord
	updateCalled bool
	lastUpdate   repository.UserSettingsUpdate
	getCalled    bool
}

func (r *fakeUserSettingsRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (repository.UserSettingsRecord, error) {
	r.getCalled = true
	return r.record, nil
}

func (r *fakeUserSettingsRepo) UpdateByUserID(ctx context.Context, userID uuid.UUID, update repository.UserSettingsUpdate) (repository.UserSettingsRecord, error) {
	r.updateCalled = true
	r.lastUpdate = update
	if update.BaseCurrency != nil {
		r.record.BaseCurrency = *update.BaseCurrency
	}
	if update.Timezone != nil {
		r.record.Timezone = *update.Timezone
	}
	if update.FullName != nil {
		r.record.FullName = *update.FullName
	}
	return r.record, nil
}

func TestUserSettingsUpdateRejectsInvalidCurrency(t *testing.T) {
	repo := &fakeUserSettingsRepo{record: repository.UserSettingsRecord{BaseCurrency: "IDR", Timezone: "Asia/Jakarta", FullName: "Eko"}}
	serviceUnderTest := service.NewUserSettingsService(repo)

	_, err := serviceUnderTest.Update(context.Background(), uuid.New(), service.UpdateUserSettingsRequest{
		BaseCurrency: stringPtr("XXX"),
	})
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr, ok := err.(*service.ValidationError)
	if !ok {
		t.Fatalf("expected validation error, got %T", err)
	}
	if validationErr.Code != "invalid_currency" {
		t.Fatalf("unexpected code %q", validationErr.Code)
	}
	if validationErr.Fields["base_currency"] != "unsupported" {
		t.Fatalf("unexpected fields %#v", validationErr.Fields)
	}
	if repo.updateCalled {
		t.Fatal("repo should not be called for invalid currency")
	}
}

func TestUserSettingsUpdateRejectsInvalidTimezone(t *testing.T) {
	repo := &fakeUserSettingsRepo{record: repository.UserSettingsRecord{BaseCurrency: "IDR", Timezone: "Asia/Jakarta", FullName: "Eko"}}
	serviceUnderTest := service.NewUserSettingsService(repo)

	_, err := serviceUnderTest.Update(context.Background(), uuid.New(), service.UpdateUserSettingsRequest{
		Timezone: stringPtr("Unknown/Timezone"),
	})
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr, ok := err.(*service.ValidationError)
	if !ok {
		t.Fatalf("expected validation error, got %T", err)
	}
	if validationErr.Code != "invalid_timezone" {
		t.Fatalf("unexpected code %q", validationErr.Code)
	}
	if validationErr.Fields["timezone"] != "unknown_timezone" {
		t.Fatalf("unexpected fields %#v", validationErr.Fields)
	}
	if repo.updateCalled {
		t.Fatal("repo should not be called for invalid timezone")
	}
}

func TestUserSettingsUpdatePartialSuccess(t *testing.T) {
	repo := &fakeUserSettingsRepo{record: repository.UserSettingsRecord{BaseCurrency: "IDR", Timezone: "Asia/Jakarta", FullName: "Eko"}}
	serviceUnderTest := service.NewUserSettingsService(repo)

	result, err := serviceUnderTest.Update(context.Background(), uuid.New(), service.UpdateUserSettingsRequest{
		BaseCurrency: stringPtr("USD"),
		Timezone:     stringPtr("Asia/Singapore"),
		FullName:     stringPtr("Eko Prasetyo"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.updateCalled {
		t.Fatal("expected repo update to be called")
	}
	if repo.lastUpdate.BaseCurrency == nil || *repo.lastUpdate.BaseCurrency != "USD" {
		t.Fatalf("unexpected base currency update %#v", repo.lastUpdate.BaseCurrency)
	}
	if repo.lastUpdate.Timezone == nil || *repo.lastUpdate.Timezone != "Asia/Singapore" {
		t.Fatalf("unexpected timezone update %#v", repo.lastUpdate.Timezone)
	}
	if repo.lastUpdate.FullName == nil || *repo.lastUpdate.FullName != "Eko Prasetyo" {
		t.Fatalf("unexpected full name update %#v", repo.lastUpdate.FullName)
	}
	if result.BaseCurrency != "USD" || result.Timezone != "Asia/Singapore" || result.FullName != "Eko Prasetyo" {
		t.Fatalf("unexpected update result %#v", result)
	}
}

func TestUserSettingsGetReturnsStoredProfile(t *testing.T) {
	repo := &fakeUserSettingsRepo{record: repository.UserSettingsRecord{BaseCurrency: "USD", Timezone: "Asia/Jakarta", FullName: "Eko Prasetyo"}}
	serviceUnderTest := service.NewUserSettingsService(repo)

	result, err := serviceUnderTest.Get(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.getCalled {
		t.Fatal("expected repo get to be called")
	}
	if result.BaseCurrency != "USD" || result.Timezone != "Asia/Jakarta" || result.FullName != "Eko Prasetyo" {
		t.Fatalf("unexpected get result %#v", result)
	}
}

func TestUserSettingsValidationRejectsShortName(t *testing.T) {
	repo := &fakeUserSettingsRepo{record: repository.UserSettingsRecord{BaseCurrency: "USD", Timezone: "Asia/Jakarta", FullName: "Eko"}}
	serviceUnderTest := service.NewUserSettingsService(repo)

	_, err := serviceUnderTest.Update(context.Background(), uuid.New(), service.UpdateUserSettingsRequest{
		FullName: stringPtr("A"),
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	var validationErr *service.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected validation error type, got %T", err)
	}
}

func stringPtr(value string) *string {
	return &value
}
