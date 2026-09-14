package store

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}
	return pool, nil
}

func InsertJob(ctx context.Context, pool *pgxpool.Pool, jobType string, payload []byte) (string, error) {
	var id string
	err := pool.QueryRow(ctx, "INSERT INTO jobs (type, payload) VALUES ($1, $2) RETURNING id", jobType, payload).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to insert job: %w", err)
	}
	return id, nil
}

func UpdateJobStatus(ctx context.Context, pool *pgxpool.Pool, id string, status string) error {
	_, err := pool.Exec(ctx, "UPDATE jobs SET status = $1, updated_at = now() WHERE id = $2", status, id)
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}
	return nil
}