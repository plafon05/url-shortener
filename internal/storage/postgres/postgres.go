package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

// Close закрывает пул соединений с базой данных
func (s *Storage) Close() error {
	return s.db.Close()
}

func New(storagePath string) (*Storage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const op = "storage.postgres.New"

	db, err := sql.Open("postgres", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = db.ExecContext(ctx, `
	CREATE TABLE IF NOT EXISTS url(
		id SERIAL PRIMARY KEY,
		alias TEXT NOT NULL UNIQUE,
		url TEXT NOT NULL)
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = db.ExecContext(ctx, "CREATE INDEX IF NOT EXISTS ind_alias ON url(alias)")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	storage := &Storage{
		db: db,
	}

	return storage, nil
}
