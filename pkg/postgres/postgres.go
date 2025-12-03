package postgres

import (
	"time"

	"github.com/Masterminds/squirrel"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/mathbdw/book/internal/interfaces/observability"
)

var sqlxOpen = sqlx.Open

// Postgres -.
type Postgres struct {
	dsn             string
	driver          string
	maxOpenConns    int
	maxIdleConns    int
	connMaxIdleTime time.Duration
	connMaxLifeTime time.Duration

	Builder squirrel.StatementBuilderType
	Sqlx    *sqlx.DB
}

// New -.
func New(log observability.Logger, opts ...Option) (*Postgres, error) {
	pg := &Postgres{}

	for _, opt := range opts {
		opt(pg)
	}

	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	db, err := sqlxOpen(pg.driver, pg.dsn)
	if err != nil {
		log.Error("postgres.New: failed to create database connection", map[string]any{"error": err.Error()})

		return nil, err
	}

	db.SetMaxOpenConns(pg.maxOpenConns)
	db.SetMaxIdleConns(pg.maxIdleConns)
	db.SetConnMaxIdleTime(pg.connMaxIdleTime)
	db.SetConnMaxLifetime(pg.connMaxLifeTime)

	if err = db.Ping(); err != nil {
		log.Error("postgres.New: failed ping the database", map[string]any{"error": err.Error()})

		return nil, err
	}

	pg.Sqlx = db

	return pg, nil
}
