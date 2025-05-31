package postgres

import (
	"context"
	"fmt"
	"quote-service/internal/domain"
	"quote-service/internal/models"
	"quote-service/pkg/logger"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DBConn interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
}

type Storage struct {
	db DBConn
}

func NewStorage(db DBConn) *Storage {
	return &Storage{db: db}
}

func (s *Storage) Create(ctx context.Context, quote *models.Quote) error {
	var newID int
	query := `
        WITH RECURSIVE ids AS (
            SELECT generate_series(1, (SELECT COALESCE(MAX(id), 0) + 1 FROM quotes)) AS id
        )
        SELECT MIN(id)
        FROM ids
        WHERE NOT EXISTS (SELECT 1 FROM quotes WHERE quotes.id = ids.id)
    `
	err := s.db.QueryRow(ctx, query).Scan(&newID)
	if err != nil {
		if err == pgx.ErrNoRows {
			newID = 1
		} else {
			logger.Errorf("Ошибка поиска свободного ID: %v", err)
			return err
		}
	}

	query = `INSERT INTO quotes (id, author, quote) VALUES ($1, $2, $3) RETURNING created_at`
	err = s.db.QueryRow(ctx, query, newID, quote.Author, quote.Quote).Scan(&quote.CreatedAt)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			logger.Errorf("Конфликт ID: %d уже занят", newID)
			return fmt.Errorf("ID %d already exists", newID)
		}
		logger.Errorf("Ошибка создания цитаты: %v", err)
		return err
	}
	quote.ID = newID
	return nil
}

func (s *Storage) GetAll(ctx context.Context) ([]models.Quote, error) {
	query := `SELECT id, author, quote, created_at FROM quotes ORDER BY id`
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		logger.Errorf("Ошибка получения всех цитат: %v", err)
		return nil, err
	}
	defer rows.Close()

	var quotes []models.Quote
	for rows.Next() {
		var q models.Quote
		if err := rows.Scan(&q.ID, &q.Author, &q.Quote, &q.CreatedAt); err != nil {
			logger.Errorf("Ошибка сканирования строки: %v", err)
			return nil, err
		}
		quotes = append(quotes, q)
	}
	if err := rows.Err(); err != nil {
		logger.Errorf("Ошибка при итерации строк: %v", err)
		return nil, err
	}
	return quotes, nil
}

func (s *Storage) GetRandom(ctx context.Context) (*models.Quote, error) {
	query := `SELECT id, author, quote, created_at FROM quotes ORDER BY RANDOM() LIMIT 1`
	var q models.Quote
	err := s.db.QueryRow(ctx, query).Scan(&q.ID, &q.Author, &q.Quote, &q.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		logger.Errorf("Ошибка получения случайной цитаты: %v", err)
		return nil, err
	}
	return &q, nil
}

func (s *Storage) GetByAuthor(ctx context.Context, author string) ([]models.Quote, error) {
	query := `SELECT id, author, quote, created_at FROM quotes WHERE author = $1 ORDER BY id`
	rows, err := s.db.Query(ctx, query, author)
	if err != nil {
		logger.Errorf("Ошибка получения цитат по автору: %v", err)
		return nil, err
	}
	defer rows.Close()

	var quotes []models.Quote
	for rows.Next() {
		var q models.Quote
		if err := rows.Scan(&q.ID, &q.Author, &q.Quote, &q.CreatedAt); err != nil {
			logger.Errorf("Ошибка сканирования строки: %v", err)
			return nil, err
		}
		quotes = append(quotes, q)
	}
	if err := rows.Err(); err != nil {
		logger.Errorf("Ошибка при итерации строк: %v", err)
		return nil, err
	}
	return quotes, nil
}

func (s *Storage) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM quotes WHERE id = $1`
	result, err := s.db.Exec(ctx, query, id)
	if err != nil {
		logger.Errorf("Ошибка удаления цитаты: %v", err)
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *Storage) Exists(ctx context.Context, author, quote string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM quotes WHERE author = $1 AND quote = $2)`
	err := s.db.QueryRow(ctx, query, author, quote).Scan(&exists)
	if err != nil {
		logger.Errorf("Ошибка проверки существования цитаты: %v", err)
		return false, err
	}
	return exists, nil
}
