package main

import (
	"context"
	"fmt"
	"os"

	"github.com/AffanSurya/xarela-backend/internal/config"
	"github.com/AffanSurya/xarela-backend/internal/db"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] != "up" {
		fmt.Fprintln(os.Stderr, "usage: go run ./cmd/migrate up")
		os.Exit(1)
	}

	cfg := config.Load()
	if err := db.Up(context.Background(), cfg.DatabaseDSN); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("migrations applied successfully")
}
