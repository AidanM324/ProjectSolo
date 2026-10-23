package store

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
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

// InsertJob returns the job ID and whether a new row was created.
// If key is non-empty and already used, it returns the original job's ID
// with created=false instead of inserting a duplicate.
func InsertJob(ctx context.Context, pool *pgxpool.Pool, jobType string, payload []byte, key string) (string, bool, error) {
	var id string

	if key == "" {
		err := pool.QueryRow(ctx,
			"INSERT INTO jobs (type, payload) VALUES ($1, $2) RETURNING id",
			jobType, payload).Scan(&id)
		if err != nil {
			return "", false, fmt.Errorf("failed to insert job: %w", err)
		}
		return id, true, nil
	}

	err := pool.QueryRow(ctx,
		`INSERT INTO jobs (type, payload, idempotency_key) VALUES ($1, $2, $3)
		 ON CONFLICT (idempotency_key) DO NOTHING
		 RETURNING id`,
		jobType, payload, key).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		// Key already used: look up the original job.
		err = pool.QueryRow(ctx,
			"SELECT id FROM jobs WHERE idempotency_key = $1", key).Scan(&id)
		if err != nil {
			return "", false, fmt.Errorf("failed to look up existing job: %w", err)
		}
		return id, false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("failed to insert job: %w", err)
	}
	return id, true, nil
}

func UpdateJobStatus(ctx context.Context, pool *pgxpool.Pool, id string, status string) error {
	_, err := pool.Exec(ctx, "UPDATE jobs SET status = $1, updated_at = now() WHERE id = $2", status, id)
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}
	return nil
}

func GetJobStatus(ctx context.Context, pool *pgxpool.Pool, id string) (string, error) {
	var status string
	err := pool.QueryRow(ctx, "SELECT status FROM jobs WHERE id = $1", id).Scan(&status)
	if err != nil {
		return "", fmt.Errorf("failed to get job status: %w", err)
	}
	return status, nil
}