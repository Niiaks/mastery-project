package repository

import (
	"context"
	"errors"
	"fmt"
	"mastery-project/internal/model"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepository struct {
	pool *pgxpool.Pool
}

type SessionRepo interface {
	CreateSession(ctx context.Context, session *model.Session) error
	DeleteSession(ctx context.Context, session *model.Session) error
	GetSession(ctx context.Context, sessionID string) (string, error)
	GetUserBySessionID(ctx context.Context, sessionID string) (*model.User, error)
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

func (r *SessionRepository) CreateSession(ctx context.Context, session *model.Session) error {
	expiresAt := time.Now().Add(30 * time.Minute)
	sql := `INSERT INTO sessions (session_id, user_id,expires_at) VALUES ($1, $2,$3) RETURNING session_id`

	err := r.pool.QueryRow(ctx, sql, session.SessionID, session.UserID, expiresAt).Scan(&session.SessionID)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

func (r *SessionRepository) DeleteSession(ctx context.Context, session *model.Session) error {
	sql := `DELETE FROM sessions WHERE session_id = $1`
	_, err := r.pool.Exec(ctx, sql, session.ID)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (r *SessionRepository) GetSession(ctx context.Context, sessionID string) (string, error) {
	sql := `SELECT * FROM sessions WHERE session_id = $1 AND expires_at > $2`
	err := r.pool.QueryRow(ctx, sql, sessionID, time.Now()).Scan(&sessionID)
	if err != nil {
		return "", fmt.Errorf("get session: %w", err)
	}
	return sessionID, nil
}

func (r *SessionRepository) GetUserBySessionID(
	ctx context.Context,
	sessionID string,
) (*model.User, error) {

	sql := `
	SELECT 
		u.id,
		u.name,
		u.email
	FROM sessions s
	INNER JOIN users u ON s.user_id = u.id
	WHERE s.session_id = $1
	  AND s.expires_at > $2;
	`

	var user model.User

	err := r.pool.QueryRow(
		ctx,
		sql,
		sessionID,
		time.Now(),
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("session invalid or expired")
		}
		return nil, fmt.Errorf("get user by session: %w", err)
	}

	return &user, nil
}
