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

## Environment

- `PORT` default: `8080`
- `LOG_LEVEL` default: `info`
- `DATABASE_DSN` default: empty
