package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type ExpenseCategorySeed struct {
	Name  string
	Color string
}

var defaultExpenseCategories = []ExpenseCategorySeed{
	{Name: "Housing", Color: "#4F46E5"},
	{Name: "Groceries", Color: "#10B981"},
	{Name: "Food & Dining", Color: "#F59E0B"},
	{Name: "Transport", Color: "#06B6D4"},
	{Name: "Health", Color: "#EF4444"},
	{Name: "Utilities", Color: "#8B5CF6"},
	{Name: "Subscriptions", Color: "#EC4899"},
	{Name: "Shopping", Color: "#F97316"},
	{Name: "Entertainment", Color: "#14B8A6"},
	{Name: "Education", Color: "#3B82F6"},
	{Name: "Travel", Color: "#22C55E"},
	{Name: "Miscellaneous", Color: "#64748B"},
}

var supportedBaseCurrencies = []string{
	"IDR",
	"USD",
	"SGD",
	"EUR",
	"JPY",
	"GBP",
	"AUD",
	"HKD",
}

type expenseCategoryExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func DefaultExpenseCategories() []ExpenseCategorySeed {
	return append([]ExpenseCategorySeed(nil), defaultExpenseCategories...)
}

func SupportedBaseCurrencies() []string {
	return append([]string(nil), supportedBaseCurrencies...)
}

func SeedDefaultExpenseCategories(ctx context.Context, executor expenseCategoryExecutor, userID uuid.UUID) error {
	if executor == nil {
		return errors.New("seed executor is required")
	}

	for _, category := range defaultExpenseCategories {
		if err := ensureExpenseCategory(ctx, executor, userID, category); err != nil {
			return err
		}
	}

	return nil
}

func ensureExpenseCategory(ctx context.Context, executor expenseCategoryExecutor, userID uuid.UUID, category ExpenseCategorySeed) error {
	if strings.TrimSpace(category.Name) == "" {
		return errors.New("expense category name is required")
	}

	var categoryID uuid.UUID
	err := executor.QueryRowContext(ctx, `
SELECT id
FROM expense_categories
WHERE user_id = $1
  AND parent_id IS NULL
  AND name = $2
`, userID, category.Name).Scan(&categoryID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("lookup expense category %q: %w", category.Name, err)
	}

	if err == nil {
		if _, err := executor.ExecContext(ctx, `
UPDATE expense_categories
SET color = $2,
    is_system = true,
    updated_at = now()
WHERE id = $1
`, categoryID, nullableString(category.Color)); err != nil {
			return fmt.Errorf("update expense category %q: %w", category.Name, err)
		}
		return nil
	}

	if _, err := executor.ExecContext(ctx, `
INSERT INTO expense_categories (user_id, parent_id, name, color, is_system)
VALUES ($1, NULL, $2, $3, true)
`, userID, category.Name, nullableString(category.Color)); err != nil {
		return fmt.Errorf("insert expense category %q: %w", category.Name, err)
	}

	return nil
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
