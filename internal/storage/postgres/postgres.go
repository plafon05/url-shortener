package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"httpServer_project/internal/storage"
	"log"
	"time"

	"github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

// Close закрывает пул соединений с базой данных
func (s *Storage) Close() error {
	return s.db.Close()
}

func New(storagePath string) (*Storage, error) {

	const op = "storage.postgres.New"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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

func (s *Storage) SaveURL(ctx context.Context, urlToSave, alias string) (int64, error) {

	const op = "storage.postgres.SaveURL"

	var id int64
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO url(alias, url)
		VALUES($1, $2)
		RETURNING id
	`, alias, urlToSave).Scan(&id)

	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrAliasExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetURL(ctx context.Context, alias string) (string, error) {

	const op = "storage.postgres.GetURL"

	var longURL string
	err := s.db.QueryRowContext(ctx, `
		SELECT url FROM url
		WHERE alias=$1
		`, alias).Scan(&longURL)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return longURL, nil
}

func (s *Storage) DeleteURL(ctx context.Context, alias string) error {

	const op = "storage.postgres.DeleteURL"

	deletedURL, err := s.db.ExecContext(ctx, "DELETE FROM url WHERE alias=$1", alias)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	row, errAffected := deletedURL.RowsAffected()
	if errAffected != nil {
		return fmt.Errorf("%s: %w", op, errAffected)
	}

	if row == 0 {
		return storage.ErrAliasNotFound
	}

	log.Println("URL был успешно удалён")
	return nil
}

func (s *Storage) GetAliasesByURL(ctx context.Context, urlToFind string) ([]string, error) {

	const op = "storage.postgres.GetAliasesByURL"

	rows, err := s.db.QueryContext(ctx, `
		SELECT alias FROM url
		WHERE url=$1
		`, urlToFind)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var aliases []string
	for rows.Next() {
		var alias string
		if err := rows.Scan(&alias); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		aliases = append(aliases, alias)
	}

	if len(aliases) == 0 {
		return nil, storage.ErrURLNotFound
	}

	return aliases, nil
}
