package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type UserSettingsRecord struct {
	BaseCurrency string
	Timezone     string
	FullName     string
}

type UserSettingsUpdate struct {
	BaseCurrency *string
	Timezone     *string
	FullName     *string
}

type UserSettingsRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (UserSettingsRecord, error)
	UpdateByUserID(ctx context.Context, userID uuid.UUID, update UserSettingsUpdate) (UserSettingsRecord, error)
}

type PostgresUserSettingsRepository struct {
	dsn string
}

func NewUserSettingsRepository(dsn string) UserSettingsRepository {
	return PostgresUserSettingsRepository{dsn: strings.TrimSpace(dsn)}
}

func (r PostgresUserSettingsRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (UserSettingsRecord, error) {
	database, err := sql.Open("postgres", r.dsn)
	if err != nil {
		return UserSettingsRecord{}, fmt.Errorf("open database: %w", err)
	}
	defer database.Close()

	if err := database.PingContext(ctx); err != nil {
		return UserSettingsRecord{}, fmt.Errorf("ping database: %w", err)
	}

	var record UserSettingsRecord
	var fullName sql.NullString
	var timezone sql.NullString
	err = database.QueryRowContext(ctx, `
SELECT base_currency, full_name, timezone
FROM users
WHERE id = $1
`, userID).Scan(&record.BaseCurrency, &fullName, &timezone)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UserSettingsRecord{}, err
		}
		return UserSettingsRecord{}, fmt.Errorf("load user settings: %w", err)
	}

	record.FullName = fullName.String
	record.Timezone = timezone.String
	return record, nil
}

func (r PostgresUserSettingsRepository) UpdateByUserID(ctx context.Context, userID uuid.UUID, update UserSettingsUpdate) (UserSettingsRecord, error) {
	database, err := sql.Open("postgres", r.dsn)
	if err != nil {
		return UserSettingsRecord{}, fmt.Errorf("open database: %w", err)
	}
	defer database.Close()

	if err := database.PingContext(ctx); err != nil {
		return UserSettingsRecord{}, fmt.Errorf("ping database: %w", err)
	}

	query, args := buildUserSettingsUpdateQuery(userID, update)
	if _, err := database.ExecContext(ctx, query, args...); err != nil {
		return UserSettingsRecord{}, fmt.Errorf("update user settings: %w", err)
	}

	return r.GetByUserID(ctx, userID)
}

func buildUserSettingsUpdateQuery(userID uuid.UUID, update UserSettingsUpdate) (string, []any) {
	setClauses := make([]string, 0, 4)
	args := make([]any, 0, 5)
	args = append(args, userID)

	if update.BaseCurrency != nil {
		setClauses = append(setClauses, fmt.Sprintf("base_currency = $%d", len(args)+1))
		args = append(args, *update.BaseCurrency)
	}
	if update.Timezone != nil {
		setClauses = append(setClauses, fmt.Sprintf("timezone = $%d", len(args)+1))
		args = append(args, nullableString(*update.Timezone))
	}
	if update.FullName != nil {
		setClauses = append(setClauses, fmt.Sprintf("full_name = $%d", len(args)+1))
		args = append(args, nullableString(*update.FullName))
	}

	setClauses = append(setClauses, "updated_at = now()")

	query := fmt.Sprintf(`
UPDATE users
SET %s
WHERE id = $1
`, strings.Join(setClauses, ",\n    "))
	return query, args
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
