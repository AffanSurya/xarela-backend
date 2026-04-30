package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func SeedBaseData(ctx context.Context, dsn string, userID uuid.UUID) error {
	if strings.TrimSpace(dsn) == "" {
		return errors.New("database dsn is required")
	}

	database, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer database.Close()

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := database.PingContext(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin seed transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if err := SeedDefaultExpenseCategories(ctx, tx, userID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit seed transaction: %w", err)
	}

	return nil
}
