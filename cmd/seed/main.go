package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/AffanSurya/xarela-backend/internal/config"
	"github.com/AffanSurya/xarela-backend/internal/db"
	"github.com/google/uuid"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "categories":
		runCategories(os.Args[2:])
	case "currencies":
		runCurrencies()
	default:
		usage()
		os.Exit(1)
	}
}

func runCategories(args []string) {
	flags := flag.NewFlagSet("categories", flag.ExitOnError)
	dsn := flags.String("database-dsn", "", "PostgreSQL connection string")
	userID := flags.String("user-id", "", "User UUID to seed default categories for")
	_ = flags.Parse(args)

	if strings.TrimSpace(*dsn) == "" {
		cfg := config.Load()
		*dsn = cfg.DatabaseDSN
	}
	if strings.TrimSpace(*dsn) == "" {
		fmt.Fprintln(os.Stderr, "database dsn is required")
		os.Exit(1)
	}
	if strings.TrimSpace(*userID) == "" {
		fmt.Fprintln(os.Stderr, "user-id is required")
		os.Exit(1)
	}

	parsedUserID, err := uuid.Parse(*userID)
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid user-id:", err)
		os.Exit(1)
	}

	if err := db.SeedBaseData(context.Background(), *dsn, parsedUserID); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("default expense categories seeded successfully")
}

func runCurrencies() {
	for _, currency := range db.SupportedBaseCurrencies() {
		fmt.Println(currency)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  go run ./cmd/seed categories --user-id <uuid> [--database-dsn <dsn>]")
	fmt.Fprintln(os.Stderr, "  go run ./cmd/seed currencies")
}
