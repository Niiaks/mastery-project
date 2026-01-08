package database

import (
	"context"
	"fmt"
	"log/slog"
	"mastery-project/internal/config"
	"net"
	"net/url"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	Pool *pgxpool.Pool
}

type multiTracer struct {
	tracers []any
}

// TraceQueryStart implements pgx tracer interface
func (mt *multiTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	for _, tracer := range mt.tracers {
		if t, ok := tracer.(interface {
			TraceQueryStart(context.Context, *pgx.Conn, pgx.TraceQueryStartData) context.Context
		}); ok {
			ctx = t.TraceQueryStart(ctx, conn, data)
		}
	}
	return ctx
}

// TraceQueryEnd implements pgx tracer interface
func (mt *multiTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	for _, tracer := range mt.tracers {
		if t, ok := tracer.(interface {
			TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData)
		}); ok {
			t.TraceQueryEnd(ctx, conn, data)
		}
	}
}

func New(cfg *config.Config) (*Database, error) {
	hostPort := net.JoinHostPort(cfg.Database.DBHost, strconv.Itoa(cfg.Database.DBPort))

	encodedPassword := url.QueryEscape(cfg.Database.DBPass)

	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		cfg.Database.DBUser,
		encodedPassword,
		hostPort,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database connection configuration: %s", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %s", err)
	}

	err2 := pool.Ping(context.Background())
	if err2 != nil {
		return nil, fmt.Errorf("failed to ping database: %s", err)
	}
	return &Database{Pool: pool}, nil
}

func (db *Database) Close() error {
	slog.Info("database closing")
	db.Pool.Close()
	return nil
}
