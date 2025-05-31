package postgres

import (
	"context"
	"fmt"
	"quote-service/pkg/logger"

	"github.com/jackc/pgx/v5"
)

type Postgres struct {
	DB *pgx.Conn
}

func NewPostgres(opts Options) (*Postgres, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		opts.Host, opts.Port, opts.User, opts.Password, opts.DBName)
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		logger.Errorf("Ошибка подключения к PostgreSQL: %v", err)
		return nil, err
	}
	return &Postgres{DB: conn}, nil
}

func (p *Postgres) Close() {
	if p.DB != nil {
		p.DB.Close(context.Background())
	}
}
