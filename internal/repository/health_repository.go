package repository

import "context"

type HealthRepository interface {
	Ping(ctx context.Context) error
}

type NoopHealthRepository struct{}

func NewHealthRepository() HealthRepository {
	return NoopHealthRepository{}
}

func (NoopHealthRepository) Ping(ctx context.Context) error {
	return nil
}
