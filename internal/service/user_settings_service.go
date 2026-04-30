package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/AffanSurya/xarela-backend/internal/db"
	"github.com/AffanSurya/xarela-backend/internal/repository"
	"github.com/google/uuid"
)

type UserSettings struct {
	BaseCurrency string `json:"base_currency"`
	Timezone     string `json:"timezone"`
	FullName     string `json:"full_name"`
}

type UpdateUserSettingsRequest struct {
	BaseCurrency *string
	Timezone     *string
	FullName     *string
}

type UserSettingsService struct {
	repo repository.UserSettingsRepository
}

func NewUserSettingsService(repo repository.UserSettingsRepository) UserSettingsService {
	return UserSettingsService{repo: repo}
}

func (s UserSettingsService) Get(ctx context.Context, userID uuid.UUID) (UserSettings, error) {
	if s.repo == nil {
		return UserSettings{}, errors.New("user settings repository is required")
	}

	record, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return UserSettings{}, err
	}

	return UserSettings{
		BaseCurrency: record.BaseCurrency,
		Timezone:     record.Timezone,
		FullName:     record.FullName,
	}, nil
}

func (s UserSettingsService) Update(ctx context.Context, userID uuid.UUID, request UpdateUserSettingsRequest) (UserSettings, error) {
	if s.repo == nil {
		return UserSettings{}, errors.New("user settings repository is required")
	}

	validated, validationErr := validateUserSettingsRequest(request)
	if validationErr != nil {
		return UserSettings{}, validationErr
	}

	record, err := s.repo.UpdateByUserID(ctx, userID, validated)
	if err != nil {
		return UserSettings{}, err
	}

	return UserSettings{
		BaseCurrency: record.BaseCurrency,
		Timezone:     record.Timezone,
		FullName:     record.FullName,
	}, nil
}

func validateUserSettingsRequest(request UpdateUserSettingsRequest) (repository.UserSettingsUpdate, *ValidationError) {
	validated := repository.UserSettingsUpdate{}
	fields := make(map[string]string)

	if request.BaseCurrency != nil {
		currency := strings.ToUpper(strings.TrimSpace(*request.BaseCurrency))
		if !isSupportedBaseCurrency(currency) {
			fields["base_currency"] = "unsupported"
		} else {
			validated.BaseCurrency = &currency
		}
	}

	if request.Timezone != nil {
		timezone := strings.TrimSpace(*request.Timezone)
		if timezone == "" {
			fields["timezone"] = "unknown_timezone"
		} else if _, err := time.LoadLocation(timezone); err != nil {
			fields["timezone"] = "unknown_timezone"
		} else {
			validated.Timezone = &timezone
		}
	}

	if request.FullName != nil {
		fullName := strings.TrimSpace(*request.FullName)
		if len(fullName) < 2 {
			fields["full_name"] = "too_short"
		} else {
			validated.FullName = &fullName
		}
	}

	if len(fields) > 0 {
		return repository.UserSettingsUpdate{}, &ValidationError{
			Code:    chooseSettingsValidationCode(fields),
			Message: chooseSettingsValidationMessage(fields),
			Fields:  fields,
		}
	}

	return validated, nil
}

func isSupportedBaseCurrency(currency string) bool {
	for _, supported := range db.SupportedBaseCurrencies() {
		if supported == currency {
			return true
		}
	}
	return false
}

func chooseSettingsValidationCode(fields map[string]string) string {
	if _, ok := fields["base_currency"]; ok {
		return "invalid_currency"
	}
	if _, ok := fields["timezone"]; ok {
		return "invalid_timezone"
	}
	if _, ok := fields["full_name"]; ok {
		return "validation_error"
	}
	return "validation_error"
}

func chooseSettingsValidationMessage(fields map[string]string) string {
	if _, ok := fields["base_currency"]; ok {
		return "Currency is not supported"
	}
	if _, ok := fields["timezone"]; ok {
		return "Timezone is not supported"
	}
	if _, ok := fields["full_name"]; ok {
		return "Invalid input"
	}
	return "Invalid input"
}
