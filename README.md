# xarela-backend

Go backend scaffold for Xarela.

## Run

```bash
go run ./cmd/server
```

## Database migrations

```bash
go run ./cmd/migrate up
```

Set `DATABASE_DSN` before running the migration command.

## Seed data

```bash
go run ./cmd/seed categories --user-id <user-uuid>
go run ./cmd/seed currencies
```

`categories` seeds the default expense categories for a user and is safe to run repeatedly. `currencies` prints the supported base currency codes.

## Environment

- `PORT` default: `8080`
- `LOG_LEVEL` default: `info`
- `DATABASE_DSN` default: empty
