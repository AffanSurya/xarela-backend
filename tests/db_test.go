package tests

import (
	"testing"

	"github.com/AffanSurya/xarela-backend/internal/db"
)

func TestDefaultExpenseCategoriesAreUnique(t *testing.T) {
	seen := make(map[string]struct{})
	for _, category := range db.DefaultExpenseCategories() {
		if category.Name == "" {
			t.Fatal("category name must not be empty")
		}
		if _, exists := seen[category.Name]; exists {
			t.Fatalf("duplicate category name %q", category.Name)
		}
		seen[category.Name] = struct{}{}
	}
}

func TestSupportedBaseCurrenciesAreUnique(t *testing.T) {
	seen := make(map[string]struct{})
	for _, currency := range db.SupportedBaseCurrencies() {
		if currency == "" {
			t.Fatal("currency code must not be empty")
		}
		if _, exists := seen[currency]; exists {
			t.Fatalf("duplicate currency %q", currency)
		}
		seen[currency] = struct{}{}
	}
}