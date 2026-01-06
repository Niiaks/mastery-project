package repository

import (
	"context"
	"errors"
	"fmt"
	"mastery-project/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

type UserRepo interface {
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	CreateUser(ctx context.Context, user *model.User) error
	EmailExists(ctx context.Context, email string) (bool, error)
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (ur *UserRepository) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	err := ur.db.QueryRow(ctx, "SELECT id,name,email,password FROM users WHERE id = $1", id).Scan(&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
	)
	if err != nil {
		return nil, fmt.Errorf("query user by id: %w", err)
	}
	return &user, nil
}

func (ur *UserRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	//scan will place the db query into the user table
	err := ur.db.QueryRow(
		ctx,
		`SELECT id, name, email, password
	 FROM users
	 WHERE email = $1`,
		email,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("query user by email: %w", err)
	}
	return &user, nil
}

func (ur *UserRepository) CreateUser(ctx context.Context, user *model.User) error {
	sql := "INSERT INTO users (name, email, password) VALUES ($1, $2,$3) RETURNING id"

	err := ur.db.QueryRow(ctx, sql, user.Name, user.Email, user.Password).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("error creating task: %w", err)
	}

	return nil
}

func (ur *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	sql := `SELECT COUNT(*) FROM users WHERE email = $1`

	var count int
	err := ur.db.QueryRow(ctx, sql, email).Scan(&count)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
	}
	return count > 0, nil
}
